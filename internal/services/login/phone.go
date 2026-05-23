package login

import (
	"context"
	"log/slog"

	"github.com/slimeyquest/server/internal/entity"
)

// PhoneRegister authenticates a phone account in the MVP test-code flow.
func (s *Service) PhoneRegister(ctx context.Context, req *entity.PhoneRegisterReq) (*entity.AuthResponse, *AuthResult) {
	if req == nil {
		return authError(entity.ErrorCodeInvalidRequest, "missing phone register payload"), nil
	}
	return s.phoneAuth(ctx, req.Phone, req.VerifyCode, req.ClientVersion, true)
}

// PhoneLogin authenticates an existing or test-created phone account.
func (s *Service) PhoneLogin(ctx context.Context, req *entity.PhoneLoginReq) (*entity.AuthResponse, *AuthResult) {
	if req == nil {
		return authError(entity.ErrorCodeInvalidRequest, "missing phone login payload"), nil
	}
	return s.phoneAuth(ctx, req.Phone, req.VerifyCode, req.ClientVersion, false)
}

func (s *Service) phoneAuth(ctx context.Context, phone, verifyCode, clientVersion string, register bool) (*entity.AuthResponse, *AuthResult) {
	if phone == "" {
		return authError(entity.ErrorCodeInvalidRequest, "phone is required"), nil
	}
	if !validTestVerifyCode(verifyCode) {
		return authError(entity.ErrorCodeInvalidRequest, "invalid verify code"), nil
	}

	p, created, err := s.loadOrCreatePlayer(ctx, PlatformPhone, phone)
	if err != nil {
		s.log.Error("phone auth failed", slog.String("phone", phone), slog.Bool("register", register), slog.String("client_version", clientVersion), slog.Any("error", err))
		return authError(entity.ErrorCodeInternal, "internal error"), nil
	}
	p, err = s.players.RecordLogin(ctx, p.ID)
	if err != nil {
		s.log.Error("phone auth record failed", slog.Int("player_id", p.ID), slog.String("phone", phone), slog.Any("error", err))
		return authError(entity.ErrorCodeInternal, "internal error"), nil
	}

	s.log.Info("phone auth", slog.Bool("register", register), slog.String("phone", phone))
	return s.finishLogin(ctx, p, created, PlatformPhone, phone, clientVersion)
}

func validTestVerifyCode(code string) bool {
	return code == "000000" || code == "123456"
}
