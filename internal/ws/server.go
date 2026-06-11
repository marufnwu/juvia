package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"juvia/internal/events"
	"juvia/internal/metrics"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server handles WebSocket connections.
type Server struct {
	log       *slog.Logger
	eventBus  *events.Bus
	clients   map[*client]struct{}
	mu        sync.RWMutex
	metrics   *metrics.Collector
}

type client struct {
	conn   *websocket.Conn
	server *Server
	send   chan []byte
}

// NewServer creates a new WebSocket server.
func NewServer(log *slog.Logger, eventBus *events.Bus) *Server {
	return &Server{
		log:      log,
		eventBus: eventBus,
		clients:  make(map[*client]struct{}),
		metrics:  metrics.NewCollector(),
	}
}

// HandleMetrics handles /ws/v1/metrics connections.
func (s *Server) HandleMetrics(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.Error("upgrade websocket", "error", err)
		return
	}
	cl := &client{conn: conn, server: s, send: make(chan []byte, 256)}
	s.mu.Lock()
	s.clients[cl] = struct{}{}
	s.mu.Unlock()

	go cl.writePump()
	go s.metricsPump(cl)
}

// HandleTasks handles /ws/v1/tasks/:taskId connections.
func (s *Server) HandleTasks(c *gin.Context) {
	taskID := c.Param("taskId")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.Error("upgrade websocket", "error", err)
		return
	}
	cl := &client{conn: conn, server: s, send: make(chan []byte, 256)}
	s.mu.Lock()
	s.clients[cl] = struct{}{}
	s.mu.Unlock()

	go cl.writePump()
	go s.taskPump(cl, taskID)
}

func (s *Server) metricsPump(cl *client) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m, err := s.metrics.Collect()
			if err != nil {
				continue
			}
			data, _ := json.Marshal(gin.H{"type": "metrics", "data": m})
			select {
			case cl.send <- data:
			default:
				return
			}
		}
	}
}

func (s *Server) taskPump(cl *client, taskID string) {
	ch := s.eventBus.Subscribe(events.EventTaskProgress)
	defer s.eventBus.Unsubscribe(events.EventTaskProgress, ch)
	ch2 := s.eventBus.Subscribe(events.EventTaskComplete)
	defer s.eventBus.Unsubscribe(events.EventTaskComplete, ch2)

	for {
		select {
		case ev := <-ch:
			data, _ := json.Marshal(ev)
			select {
			case cl.send <- data:
			default:
				return
			}
		case ev := <-ch2:
			data, _ := json.Marshal(ev)
			select {
			case cl.send <- data:
			default:
				return
			}
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.server.mu.Lock()
		delete(c.server.clients, c)
		c.server.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

// Broadcast sends a message to all connected clients.
func (s *Server) Broadcast(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for cl := range s.clients {
		select {
		case cl.send <- data:
		default:
		}
	}
}
