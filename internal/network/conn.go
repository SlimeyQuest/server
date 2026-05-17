package network

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	"github.com/slimeyquest/server/internal/login"
)

const pingPeriod = 54 * time.Second

// Conn represents a single WebSocket client connection.
type Conn struct {
	id            string
	log           *slog.Logger
	ws            *websocket.Conn
	hub           *Hub
	loginSvc      *login.Service
	send          chan []byte
	ctx           context.Context
	cancel        context.CancelFunc
	once          sync.Once
	playerID      int64
	token         string
	authenticated bool
}

func newConn(id string, log *slog.Logger, ws *websocket.Conn, hub *Hub, loginSvc *login.Service) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		id:       id,
		log:      log.With("conn_id", id),
		ws:       ws,
		hub:      hub,
		loginSvc: loginSvc,
		send:     make(chan []byte, 16),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// ID returns the connection identifier.
func (c *Conn) ID() string {
	return c.id
}

// PlayerID returns the authenticated player id, or zero if not logged in.
func (c *Conn) PlayerID() int64 {
	return c.playerID
}

// Token returns the active session token, or empty if not logged in.
func (c *Conn) Token() string {
	return c.token
}

// SetAuthenticated marks the connection as logged in.
func (c *Conn) SetAuthenticated(playerID int64, token string) {
	c.playerID = playerID
	c.token = token
	c.authenticated = true
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

		if c.authenticated {
			c.log.Info("login_rejected", "reason", "unexpected_message_after_login")
			return
		}

		if c.handleGuestLogin(message) {
			return
		}
	}
}

// handleGuestLogin processes the first login packet. Returns true when the connection should close.
func (c *Conn) handleGuestLogin(message []byte) bool {
	req := &loginv1.GuestLoginReq{}
	if err := proto.Unmarshal(message, req); err != nil {
		c.log.Warn("invalid guest login payload", "error", err)
		c.sendLoginResponse(errorRes(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "invalid guest login payload"))
		c.log.Info("login_rejected", "reason", "invalid_payload")
		return true
	}

	res := c.loginSvc.GuestLogin(c.ctx, c, req)
	if !login.IsSuccess(res) {
		c.sendLoginResponse(res)
		c.log.Info("login_rejected", "reason", "login_failed")
		return true
	}

	if !c.sendLoginResponse(res) {
		c.log.Info("login_rejected", "reason", "marshal_response_failed")
		return true
	}

	return false
}

func (c *Conn) sendLoginResponse(res *loginv1.GuestLoginRes) bool {
	payload, err := proto.Marshal(res)
	if err != nil {
		c.log.Error("marshal guest login response failed", "error", err)
		return false
	}
	c.sendResponse(payload)
	return true
}

func errorRes(code commonv1.ErrorCode, message string) *loginv1.GuestLoginRes {
	return &loginv1.GuestLoginRes{
		Error: &commonv1.ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

func (c *Conn) sendResponse(payload []byte) {
	select {
	case <-c.ctx.Done():
	case c.send <- payload:
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
