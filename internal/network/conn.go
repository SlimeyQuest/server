package network

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const pingPeriod = 54 * time.Second

// Conn represents a single WebSocket client connection.
type Conn struct {
	id     string
	log    *slog.Logger
	ws     *websocket.Conn
	hub    *Hub
	send   chan []byte
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

func newConn(id string, log *slog.Logger, ws *websocket.Conn, hub *Hub) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		id:     id,
		log:    log.With("conn_id", id),
		ws:     ws,
		hub:    hub,
		send:   make(chan []byte, 16),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ID returns the connection identifier.
func (c *Conn) ID() string {
	return c.id
}

// Serve starts read and write pumps until the connection closes.
func (c *Conn) Serve() {
	c.hub.Register(c)
	defer func() {
		c.hub.Unregister(c)
		c.Close()
	}()

	go c.writePump()
	c.readPump()
}

func (c *Conn) readPump() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, message, err := c.ws.Read(c.ctx)
		if err != nil {
			if !isBenignClose(err) && !errors.Is(err, context.Canceled) {
				c.log.Debug("websocket read ended", "error", err)
			}
			return
		}

		// Foundation only: discard client frames until protobuf handlers exist.
		c.log.Debug("websocket frame received", "bytes", len(message))
	}
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.ws.Write(c.ctx, websocket.MessageBinary, message); err != nil {
				c.log.Debug("websocket write failed", "error", err)
				return
			}
		case <-ticker.C:
			if err := c.ws.Ping(c.ctx); err != nil {
				return
			}
		}
	}
}

// Close cancels the connection context and closes the underlying socket.
func (c *Conn) Close() {
	c.once.Do(func() {
		c.cancel()
		_ = c.ws.Close(websocket.StatusNormalClosure, "")
	})
}

func isBenignClose(err error) bool {
	code := websocket.CloseStatus(err)
	return code == websocket.StatusNormalClosure || code == websocket.StatusGoingAway
}
