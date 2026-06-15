package ws

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type terminalMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type terminalInput struct {
	Input string `json:"input"`
}

type terminalResize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type terminalSession struct {
	conn   *websocket.Conn
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mu     sync.Mutex
	done   chan struct{}
}

func (s *Server) HandleTerminal(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.Error("upgrade websocket", "error", err)
		return
	}

	shell := "/bin/bash"
	if runtime.GOOS == "windows" {
		shell = "cmd"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		s.log.Error("stdin pipe", "error", err)
		conn.Close()
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.log.Error("stdout pipe", "error", err)
		stdin.Close()
		conn.Close()
		return
	}

	cmd.Stderr = cmd.Stdout

	sess := &terminalSession{
		conn:   conn,
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		done:   make(chan struct{}),
	}

	if err := cmd.Start(); err != nil {
		s.log.Error("start shell", "error", err)
		stdin.Close()
		stdout.Close()
		conn.Close()
		return
	}

	s.log.Info("terminal session started", "sessionId", sessionID, "pid", cmd.Process.Pid)

	go sess.readPump(s.log)
	go sess.writePump(s.log)

	<-sess.done
	sess.cleanup(s.log, sessionID)
}

func (sess *terminalSession) readPump(log *slog.Logger) {
	defer close(sess.done)

	for {
		_, message, err := sess.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error("ws read", "error", err)
			}
			return
		}

		var msg terminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Error("unmarshal message", "error", err)
			continue
		}

		switch msg.Type {
		case "input":
			var input terminalInput
			if err := json.Unmarshal(msg.Data, &input); err != nil {
				log.Error("unmarshal input", "error", err)
				continue
			}
			sess.mu.Lock()
			_, err := sess.stdin.Write([]byte(input.Input))
			sess.mu.Unlock()
			if err != nil {
				log.Error("write stdin", "error", err)
				return
			}

		case "resize":
			var resize terminalResize
			if err := json.Unmarshal(msg.Data, &resize); err != nil {
				log.Error("unmarshal resize", "error", err)
				continue
			}
			sess.mu.Lock()
			// TODO: send SIGWINCH to process when creack/pty is available
			_ = resize
			sess.mu.Unlock()

		case "ping":
			sess.mu.Lock()
			sess.conn.WriteMessage(websocket.PongMessage, nil)
			sess.mu.Unlock()
		}
	}
}

func (sess *terminalSession) writePump(log *slog.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	buf := make([]byte, 4096)
	for {
		select {
		case <-sess.done:
			return
		default:
		}

		n, err := sess.stdout.Read(buf)
		if n > 0 {
			output := make([]byte, n)
			copy(output, buf[:n])

			msg, _ := json.Marshal(terminalMessage{
				Type: "output",
				Data: output,
			})

			sess.mu.Lock()
			sess.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := sess.conn.WriteMessage(websocket.TextMessage, msg)
			sess.mu.Unlock()

			if err != nil {
				log.Error("ws write", "error", err)
				return
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Error("read stdout", "error", err)
			}
			return
		}
	}
}

func (sess *terminalSession) cleanup(log *slog.Logger, sessionID string) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.stdin != nil {
		sess.stdin.Close()
	}
	if sess.stdout != nil {
		sess.stdout.Close()
	}

	if sess.cmd.Process != nil {
		if err := sess.cmd.Process.Kill(); err != nil {
			log.Debug("kill process", "error", err)
		}
		sess.cmd.Wait()
	}

	sess.conn.Close()
	log.Info("terminal session ended", "sessionId", sessionID)
}
