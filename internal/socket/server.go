package socket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

// HandlerFunc is a function that handles a JSON-RPC method.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Server is the Agent Unix socket server.
type Server struct {
	socketPath string
	listener   net.Listener
	handlers   map[string]HandlerFunc
	mu         sync.RWMutex
	log        *slog.Logger
}

// NewServer creates a new Unix socket server.
func NewServer(socketPath string, log *slog.Logger) *Server {
	return &Server{
		socketPath: socketPath,
		handlers:   make(map[string]HandlerFunc),
		log:        log,
	}
}

// RegisterMethod registers a handler for a JSON-RPC method.
func (s *Server) RegisterMethod(method string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// Listen creates the Unix socket and starts listening.
func (s *Server) Listen() error {
	_ = os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen on unix socket: %w", err)
	}
	s.listener = listener

	if err := os.Chmod(s.socketPath, 0660); err != nil {
		return fmt.Errorf("chmod socket: %w", err)
	}

	if grp, err := userLookupGroup("juvia"); err == nil {
		gid, _ := strconv.Atoi(grp.Gid)
		if gid > 0 {
			os.Chown(s.socketPath, 0, gid)
		}
	}

	s.log.Info("agent socket listening", "path", s.socketPath)
	return nil
}

// Serve accepts connections and handles them.
func (s *Server) Serve(ctx context.Context) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				s.log.Error("accept connection", "error", err)
				continue
			}
		}
		go s.handleConn(ctx, conn)
	}
}

// Close stops the server.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			s.writeResponse(conn, NewErrorResponse("", "PARSE_ERROR", err.Error()))
			continue
		}

		s.mu.RLock()
		handler, ok := s.handlers[req.Method]
		s.mu.RUnlock()

		if !ok {
			s.writeResponse(conn, NewErrorResponse(req.ID, "METHOD_NOT_FOUND", "method not found: "+req.Method))
			continue
		}

		paramsBytes, _ := json.Marshal(req.Params)
		result, err := handler(ctx, paramsBytes)
		if err != nil {
			s.writeResponse(conn, NewErrorResponse(req.ID, "EXECUTION_ERROR", err.Error()))
			continue
		}

		s.writeResponse(conn, NewResponse(req.ID, result))
	}
}

func (s *Server) writeResponse(conn net.Conn, resp *Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	_, _ = conn.Write(data)
}

type userGroup struct {
	Name string
	Gid  string
}

func userLookupGroup(name string) (*userGroup, error) {
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 4)
		if len(parts) >= 4 && parts[0] == name {
			return &userGroup{Name: parts[0], Gid: parts[2]}, nil
		}
	}
	return nil, fmt.Errorf("group not found")
}
