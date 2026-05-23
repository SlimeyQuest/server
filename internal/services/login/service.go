package login

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/slimeyquest/ent"
	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/internal/services/stage"
)

// AuthResult contains successful session identity for interface adapters.
type AuthResult struct {
	PlayerID     int64
	SessionToken string
}

// Service handles guest and phone login.
type Service struct {
	log      *slog.Logger
	players  player.Repository
	sessions *session.Manager
	idle     *idle.Service
	stage    *stage.Service
}

// NewService creates a login service.
func NewService(
	log *slog.Logger,
	players player.Repository,
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
func (s *Service) GuestLogin(ctx context.Context, req *apitypes.GuestLoginReq) (*apitypes.AuthResponse, *AuthResult) {
	if req == nil {
		return authError(apitypes.ErrorCodeInvalidRequest, "missing guest login payload"), nil
	}
	clientVersion := req.ClientVersion
	externalID := req.DeviceID
	if externalID == "" {
		return authError(apitypes.ErrorCodeInvalidRequest, "device_id is required"), nil
	}

	p, created, err := s.loadOrCreatePlayer(ctx, PlatformGuest, externalID)
	if err != nil {
		s.log.Error("guest login failed",
			"platform", PlatformGuest,
			"external_id", externalID,
			"client_version", clientVersion,
			"error", err,
		)
		return authError(apitypes.ErrorCodeInternal, "internal error"), nil
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
		return authError(apitypes.ErrorCodeInternal, "internal error"), nil
	}

	return s.finishLogin(ctx, p, created, PlatformGuest, externalID, clientVersion)
}

func (s *Service) finishLogin(ctx context.Context, p *ent.Player, created bool, platform, externalID, clientVersion string) (*apitypes.AuthResponse, *AuthResult) {
	state := player.FromEntity(p)
	now := time.Now().UTC()
	profile := player.ToProfile(state, s.players.Cfg())
	idleState := s.idle.PreviewForLogin(ctx, state, now)
	idleState.PlayerSnapshot = profile
	stageState := s.stage.BuildStageState(state)

	playerID := int64(p.ID)
	newSession, replaced := s.sessions.Bind(playerID)
	if replaced != nil {
		s.log.Info("session replacement",
			"player_id", playerID,
			"platform", platform,
			"external_id", externalID,
			"client_version", clientVersion,
			"old_token", replaced.Token,
			"new_token", newSession.Token,
		)
	}

	if created {
		s.log.Info("player created",
			"player_id", playerID,
			"platform", platform,
			"external_id", externalID,
			"client_version", clientVersion,
		)
	} else {
		s.log.Info("player loaded",
			"player_id", playerID,
			"platform", platform,
			"external_id", externalID,
			"client_version", clientVersion,
		)
	}

	s.log.Info("login success",
		"player_id", playerID,
		"token", newSession.Token,
		"platform", platform,
		"external_id", externalID,
		"client_version", clientVersion,
		"created", created,
	)

	return &apitypes.AuthResponse{
			SessionToken: newSession.Token,
			PlayerID:     playerID,
			Profile:      profile,
			IdleState:    idleState,
			StageState:   stageState,
		}, &AuthResult{
			PlayerID:     playerID,
			SessionToken: newSession.Token,
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

func authError(code, message string) *apitypes.AuthResponse {
	return &apitypes.AuthResponse{Error: apitypes.Err(code, message)}
}

// IsSuccess reports whether a login response represents success.
func IsSuccess(res *apitypes.AuthResponse) bool {
	return res != nil && !apitypes.HasError(res.Error)
}

// ValidateResponse returns an error for failed login responses.
func ValidateResponse(res *apitypes.AuthResponse) error {
	if IsSuccess(res) {
		return nil
	}
	if res == nil || res.Error == nil {
		return errors.New("login failed")
	}
	return fmt.Errorf("login failed: %s", res.Error.Message)
}
