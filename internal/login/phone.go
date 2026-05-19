package login

import (
	"context"
	"log/slog"
	"time"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	"github.com/slimeyquest/server/internal/player"
)

// PhoneRegister authenticates a phone account in the MVP test-code flow.
func (s *Service) PhoneRegister(ctx context.Context, conn LiveConn, req *loginv1.PhoneRegisterReq) *loginv1.PhoneAuthRes {
	if req == nil {
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "missing phone register payload")
	}
	return s.phoneAuth(ctx, conn, req.GetPhone(), req.GetVerifyCode(), req.GetClientVersion(), true)
}

// PhoneLogin authenticates an existing or test-created phone account.
func (s *Service) PhoneLogin(ctx context.Context, conn LiveConn, req *loginv1.PhoneLoginReq) *loginv1.PhoneAuthRes {
	if req == nil {
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "missing phone login payload")
	}
	return s.phoneAuth(ctx, conn, req.GetPhone(), req.GetVerifyCode(), req.GetClientVersion(), false)
}

func (s *Service) phoneAuth(ctx context.Context, conn LiveConn, phone, verifyCode, clientVersion string, register bool) *loginv1.PhoneAuthRes {
	if phone == "" {
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "phone is required")
	}
	if !validTestVerifyCode(verifyCode) {
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "invalid verify code")
	}

	p, created, err := s.loadOrCreatePlayer(ctx, PlatformPhone, phone)
	if err != nil {
		s.log.Error("phone auth failed", slog.String("phone", phone), slog.Bool("register", register), slog.String("client_version", clientVersion), slog.Any("error", err))
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INTERNAL, "internal error")
	}
	p, err = s.players.RecordLogin(ctx, p.ID)
	if err != nil {
		s.log.Error("phone auth record failed", slog.Int("player_id", p.ID), slog.String("phone", phone), slog.Any("error", err))
		return phoneError(commonv1.ErrorCode_ERROR_CODE_INTERNAL, "internal error")
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
		replaced.Conn.Close()
	}
	conn.SetAuthenticated(playerID, newSession.Token)

	s.log.Info("phone auth success", slog.Int64("player_id", playerID), slog.String("phone", phone), slog.Bool("created", created), slog.Bool("register", register))

	return &loginv1.PhoneAuthRes{
		SessionToken: newSession.Token,
		PlayerId:     playerID,
		Profile:      profile,
		IdleState:    idleState,
		StageState:   stageState,
	}
}

func validTestVerifyCode(code string) bool {
	return code == "000000" || code == "123456"
}

func phoneError(code commonv1.ErrorCode, message string) *loginv1.PhoneAuthRes {
	return &loginv1.PhoneAuthRes{Error: &commonv1.ErrorInfo{Code: code, Message: message}}
}
