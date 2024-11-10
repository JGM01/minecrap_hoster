package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Configures server-sent events streaming parameters
type SSEConfig struct {
	HeartbeatInterval time.Duration
	PollInterval      time.Duration
}

// Creates a configuration with safe default values
func DefaultSSEConfig() SSEConfig {
	return SSEConfig{
		HeartbeatInterval: 15 * time.Second,
		PollInterval:      100 * time.Millisecond,
	}
}

// Manages the state of an SSE connection
type sseConnection struct {
	writer   http.ResponseWriter
	flusher  http.Flusher
	seenLogs map[string]bool
	lastLen  int
	status   uint8
	config   SSEConfig
}

// Streams server logs using Server-Sent Events
func (h *Handler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	log.Printf("New SSE connection from %s", r.RemoteAddr)

	conn, err := newSSEConnection(w, DefaultSSEConfig())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := conn.initialize(h); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conn.streamLogs(h, r.Context().Done())
}

func newSSEConnection(w http.ResponseWriter, config SSEConfig) (*sseConnection, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}

	setSSEHeaders(w)

	return &sseConnection{
		writer:   w,
		flusher:  flusher,
		seenLogs: make(map[string]bool),
		config:   config,
	}, nil
}

func setSSEHeaders(w http.ResponseWriter) {
	headers := map[string]string{
		"Content-Type":                "text/event-stream",
		"Cache-Control":               "no-cache",
		"Connection":                  "keep-alive",
		"Access-Control-Allow-Origin": "*",
		"X-Accel-Buffering":           "no",
	}

	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

func (c *sseConnection) initialize(h *Handler) error {
	// Send connected event
	if err := c.sendEvent("connected", "Connected to log stream"); err != nil {
		return fmt.Errorf("failed to send connected event: %v", err)
	}

	// Send initial status
	c.status = h.server.Status
	if err := c.sendEvent("status", getStatusHTML(c.status)); err != nil {
		return fmt.Errorf("failed to send initial status: %v", err)
	}

	// Send initial logs
	if err := c.sendInitialLogs(h); err != nil {
		return fmt.Errorf("failed to send initial logs: %v", err)
	}

	return nil
}

func (c *sseConnection) sendInitialLogs(h *Handler) error {
	currentLogs := h.server.GetLogs()
	c.lastLen = len(currentLogs)

	newLogs := c.processNewLogs(currentLogs)
	if len(newLogs) > 0 {
		return c.sendLogBatch(newLogs)
	}
	return nil
}

func (c *sseConnection) streamLogs(h *Handler, done <-chan struct{}) {
	heartbeat := time.NewTicker(c.config.HeartbeatInterval)
	defer heartbeat.Stop()

	var mutex sync.Mutex // Protects access to connection state

	for {
		select {
		case <-done:
			log.Printf("SSE connection closed")
			return

		case <-heartbeat.C:
			mutex.Lock()
			err := c.sendEvent("heartbeat", "ping")
			mutex.Unlock()
			if err != nil {
				log.Printf("Error sending heartbeat: %v", err)
				return
			}

		default:
			mutex.Lock()
			if err := c.processUpdates(h); err != nil {
				log.Printf("Error processing updates: %v", err)
				mutex.Unlock()
				return
			}
			mutex.Unlock()
			time.Sleep(c.config.PollInterval)
		}
	}
}

func (c *sseConnection) processUpdates(h *Handler) error {
	// Check status changes
	if err := c.checkStatus(h); err != nil {
		return fmt.Errorf("status check failed: %v", err)
	}

	// Check for new logs
	if err := c.checkNewLogs(h); err != nil {
		return fmt.Errorf("log check failed: %v", err)
	}

	return nil
}

func (c *sseConnection) checkStatus(h *Handler) error {
	currentStatus := h.server.Status
	if currentStatus != c.status {
		if err := c.sendEvent("status", getStatusHTML(currentStatus)); err != nil {
			return err
		}
		c.status = currentStatus
	}
	return nil
}

func (c *sseConnection) checkNewLogs(h *Handler) error {
	currentLogs := h.server.GetLogs()
	currentLen := len(currentLogs)

	if currentLen > c.lastLen {
		newLogs := c.processNewLogs(currentLogs[c.lastLen:])
		if len(newLogs) > 0 {
			if err := c.sendLogBatch(newLogs); err != nil {
				return err
			}
		}
		c.lastLen = currentLen
	}
	return nil
}

func (c *sseConnection) processNewLogs(logs []string) []string {
	var newLogs []string
	for _, logLine := range logs {
		hash := hashLog(logLine)
		if !c.seenLogs[hash] {
			c.seenLogs[hash] = true
			newLogs = append(newLogs, logLine)
		}
	}
	return newLogs
}

func (c *sseConnection) sendEvent(event, data string) error {
	_, err := fmt.Fprintf(c.writer, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	c.flusher.Flush()
	return nil
}

func (c *sseConnection) sendLogBatch(logs []string) error {
	if len(logs) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.Grow(len(logs) * 64)

	for i, line := range logs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(line)
	}

	return c.sendEvent("log", sb.String())
}

// Generates a unique hash for a log line
func hashLog(log string) string {
	hash := sha256.Sum256([]byte(log))
	return hex.EncodeToString(hash[:])
}
