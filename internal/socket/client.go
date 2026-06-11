package socket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Client is the Panel Unix socket client.
type Client struct {
	socketPath string
	log        *slog.Logger
	mu         sync.Mutex
	conn       net.Conn
	reader     *bufio.Reader
	offline    bool
	failures   int
}

// NewClient creates a new socket client.
func NewClient(socketPath string, log *slog.Logger) *Client {
	return &Client{socketPath: socketPath, log: log}
}

// Call sends a JSON-RPC request and waits for a response.
func (c *Client) Call(ctx context.Context, method string, params interface{}) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.offline && c.failures >= 3 {
		return nil, fmt.Errorf("agent offline")
	}

	if c.conn == nil {
		if err := c.connect(); err != nil {
			c.failures++
			if c.failures >= 3 {
				c.offline = true
			}
			return nil, fmt.Errorf("connect to agent: %w", err)
		}
	}

	req := Request{
		ID:     fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Method: method,
		Params: params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		c.conn.Close()
		c.conn = nil
		c.failures++
		return nil, fmt.Errorf("write request: %w", err)
	}

	// Simple read with timeout
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		c.conn.Close()
		c.conn = nil
		c.failures++
		return nil, fmt.Errorf("read response: %w", err)
	}
	c.conn.SetReadDeadline(time.Time{})

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	c.failures = 0
	c.offline = false
	return &resp, nil
}

// IsOffline returns true if the agent is considered offline.
func (c *Client) IsOffline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.offline
}

func (c *Client) connect() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
