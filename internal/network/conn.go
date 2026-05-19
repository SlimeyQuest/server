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
	gatewayv1 "github.com/slimeyquest/proto/gen/go/gateway"
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
	gameplay      *Gameplay
	send          chan []byte
	ctx           context.Context
	cancel        context.CancelFunc
	once          sync.Once
	playerID      int64
	token         string
	authenticated bool
}

func newConn(id string, log *slog.Logger, ws *websocket.Conn, hub *Hub, loginSvc *login.Service, gameplay *Gameplay) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		id:       id,
		log:      log.With("conn_id", id),
		ws:       ws,
		hub:      hub,
		loginSvc: loginSvc,
		gameplay: gameplay,
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

		c.handleClientMessage(message)
	}
}

func (c *Conn) handleClientMessage(message []byte) {
	msg := &gatewayv1.ClientMessage{}
	if err := proto.Unmarshal(message, msg); err != nil {
		c.log.Warn("invalid client message", "error", err)
		if !c.authenticated {
			c.sendLoginResponse(errorRes(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "invalid client message"))
			c.log.Info("login_rejected", "reason", "invalid_payload")
			c.Close()
		}
		return
	}

	switch p := msg.Payload.(type) {
	case *gatewayv1.ClientMessage_GuestLogin:
		if c.authenticated {
			c.log.Warn("login_rejected", "reason", "already_authenticated")
			return
		}
		if c.handleGuestLogin(p.GuestLogin) {
			c.Close()
		}
	case *gatewayv1.ClientMessage_PhoneRegister:
		if c.authenticated {
			c.log.Warn("login_rejected", "reason", "already_authenticated")
			return
		}
		if c.handlePhoneRegister(p.PhoneRegister) {
			c.Close()
		}
	case *gatewayv1.ClientMessage_PhoneLogin:
		if c.authenticated {
			c.log.Warn("login_rejected", "reason", "already_authenticated")
			return
		}
		if c.handlePhoneLogin(p.PhoneLogin) {
			c.Close()
		}
	case *gatewayv1.ClientMessage_ClaimIdleRewards:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleClaimIdle(c, p.ClaimIdleRewards); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_PushStage:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handlePushStage(c, p.PushStage); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_CreateRole:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleCreateRole(c, p.CreateRole); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_ChestOpen:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleChestOpen(c, p.ChestOpen); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_DecomposeEquipment:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleDecomposeEquipment(c, p.DecomposeEquipment); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_UpgradeChest:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleUpgradeChest(c, p.UpgradeChest); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_DrawSkill:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleDrawSkill(c, p.DrawSkill); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_DrawCompanion:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleDrawCompanion(c, p.DrawCompanion); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	case *gatewayv1.ClientMessage_EquipItem:
		if !c.authenticated {
			c.log.Warn("gameplay_rejected", "reason", "unauthenticated")
			return
		}
		if err := c.gameplay.handleEquipItem(c, p.EquipItem); err != nil {
			c.log.Debug("gameplay message error", "error", err)
		}
	default:
		c.log.Warn("gameplay_rejected", "reason", "unknown_message")
	}
}

// handleGuestLogin processes guest login. Returns true when the connection should close.
func (c *Conn) handleGuestLogin(req *loginv1.GuestLoginReq) bool {
	if req == nil {
		c.log.Warn("invalid guest login payload", "error", "missing guest_login")
		c.sendLoginResponse(errorRes(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "missing guest login payload"))
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

func (c *Conn) handlePhoneRegister(req *loginv1.PhoneRegisterReq) bool {
	res := c.loginSvc.PhoneRegister(c.ctx, c, req)
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		c.sendPhoneRegisterResponse(res)
		return true
	}
	return !c.sendPhoneRegisterResponse(res)
}

func (c *Conn) handlePhoneLogin(req *loginv1.PhoneLoginReq) bool {
	res := c.loginSvc.PhoneLogin(c.ctx, c, req)
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		c.sendPhoneLoginResponse(res)
		return true
	}
	return !c.sendPhoneLoginResponse(res)
}

func (c *Conn) sendLoginResponse(res *loginv1.GuestLoginRes) bool {
	return c.sendServerMessage(&gatewayv1.ServerMessage{
		Payload: &gatewayv1.ServerMessage_GuestLogin{GuestLogin: res},
	}) == nil
}

func (c *Conn) sendPhoneRegisterResponse(res *loginv1.PhoneAuthRes) bool {
	return c.sendServerMessage(&gatewayv1.ServerMessage{
		Payload: &gatewayv1.ServerMessage_PhoneRegister{PhoneRegister: res},
	}) == nil
}

func (c *Conn) sendPhoneLoginResponse(res *loginv1.PhoneAuthRes) bool {
	return c.sendServerMessage(&gatewayv1.ServerMessage{
		Payload: &gatewayv1.ServerMessage_PhoneLogin{PhoneLogin: res},
	}) == nil
}

func (c *Conn) sendServerMessage(msg *gatewayv1.ServerMessage) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		c.log.Error("marshal server message failed", "error", err)
		return err
	}
	c.sendResponse(payload)
	return nil
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
