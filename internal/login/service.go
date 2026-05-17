package login

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	"github.com/slimeyquest/server/ent"
	"github.com/slimeyquest/server/internal/idle"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/session"
	"github.com/slimeyquest/server/internal/stage"
)

// LiveConn is the websocket connection used during login.
type LiveConn interface {
	ID() string
	Close()
	SetAuthenticated(playerID int64, token string)
}

// Service handles guest login.
type Service struct {
	log      *slog.Logger
	players  *player.Repository
	sessions *session.Manager
	idle     *idle.Service
	stage    *stage.Service
}

// NewService creates a login service.
func NewService(
	log *slog.Logger,
	players *player.Repository,
	sessions *session.Manager,
	idleSvc *idle.Service,
	stageSvc *stage.Service,
) *Service {
	return &Service{
		log:      log,
		players:  players,
		sessions: sessions,
		idle:     idleSvc,
		stage:    stageSvc,
	}
}

// GuestLogin authenticates a guest and returns a login response.
func (s *Service) GuestLogin(ctx context.Context, conn LiveConn, req *loginv1.GuestLoginReq) *loginv1.GuestLoginRes {
	clientVersion := req.GetClientVersion()
	externalID := req.GetDeviceId()
	if externalID == "" {
		return errorRes(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "device_id is required")
	}

	p, created, err := s.loadOrCreatePlayer(ctx, PlatformGuest, externalID)
	if err != nil {
		s.log.Error("guest login failed",
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
			"error", err,
		)
		return errorRes(commonv1.ErrorCode_ERROR_CODE_INTERNAL, "internal error")
	}

	p, err = s.players.RecordLogin(ctx, p.ID)
	if err != nil {
		s.log.Error("guest login record failed",
			"player_id", p.ID,
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
			"error", err,
		)
		return errorRes(commonv1.ErrorCode_ERROR_CODE_INTERNAL, "internal error")
	}

	state := player.FromEntity(p)
	now := time.Now().UTC()
	profile := player.ToProfile(state, s.players.Cfg())
	idleState := s.idle.PreviewForLogin(ctx, state, now)
	idleState.PlayerSnapshot = profile
	stageState := s.stage.BuildStageState(state)

	playerID := int64(p.ID)
	newSession, replaced := s.sessions.Bind(playerID, conn)
	if replaced != nil {
		s.log.Info("reconnect replacement",
			"player_id", playerID,
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
			"old_token", replaced.Token,
			"new_token", newSession.Token,
			"old_conn_id", replaced.Conn.ID(),
			"new_conn_id", conn.ID(),
		)
		replaced.Conn.Close()
	}

	conn.SetAuthenticated(playerID, newSession.Token)

	if created {
		s.log.Info("player created",
			"player_id", playerID,
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
		)
	} else {
		s.log.Info("player loaded",
			"player_id", playerID,
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
		)
	}

	s.log.Info("login success",
		"player_id", playerID,
		"token", newSession.Token,
		"platform", PlatformGuest,
		"external_id", externalID,
		"client_version", clientVersion,
		"created", created,
		"conn_id", conn.ID(),
	)

	return &loginv1.GuestLoginRes{
		SessionToken: newSession.Token,
		PlayerId:     playerID,
		Profile:      profile,
		IdleState:    idleState,
		StageState:   stageState,
	}
}

func (s *Service) loadOrCreatePlayer(ctx context.Context, platform, externalID string) (*ent.Player, bool, error) {
	p, err := s.players.FindByExternalID(ctx, platform, externalID)
	if err == nil {
		return p, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, err
	}

	p, err = s.players.CreatePlayerForPlatform(ctx, platform, externalID, player.DefaultNickname())
	if err != nil {
		return nil, false, err
	}
	return p, true, nil
}

func errorRes(code commonv1.ErrorCode, message string) *loginv1.GuestLoginRes {
	return &loginv1.GuestLoginRes{
		Error: &commonv1.ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// IsSuccess reports whether a login response represents success.
func IsSuccess(res *loginv1.GuestLoginRes) bool {
	if res == nil || res.Error == nil {
		return true
	}
	return res.Error.Code == commonv1.ErrorCode_ERROR_CODE_OK || res.Error.Code == 0
}

// ValidateResponse returns an error for failed login responses.
func ValidateResponse(res *loginv1.GuestLoginRes) error {
	if IsSuccess(res) {
		return nil
	}
	if res.Error == nil {
		return errors.New("login failed")
	}
	return fmt.Errorf("login failed: %s", res.Error.Message)
}
