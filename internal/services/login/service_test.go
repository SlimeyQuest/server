package login_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	"github.com/slimeyquest/ent"
	"github.com/slimeyquest/ent/enttest"
	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/data/playerrepo"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/login"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/internal/services/stage"
)

func newTestLoginService(t *testing.T, client *ent.Client) *login.Service {
	t.Helper()
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := playerrepo.New(client, cfg)
	sessions := session.NewManager()
	rewardSvc := reward.NewService(log, cfg, repo)
	idleSvc := idle.NewService(log, cfg, repo, rewardSvc)
	stageSvc := stage.NewService(log, cfg, repo, rewardSvc)
	return login.NewService(log, repo, sessions, idleSvc, stageSvc)
}

func TestGuestLoginCreatesAndResumesPlayer(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := newTestLoginService(t, client)
	ctx := context.Background()

	res, auth := svc.GuestLogin(ctx, &apitypes.GuestLoginReq{
		DeviceID:      "device-a",
		ClientVersion: "1.0.0",
	})
	if !login.IsSuccess(res) {
		t.Fatalf("expected success, got %#v", res.Error)
	}
	if res.PlayerID == 0 || res.SessionToken == "" {
		t.Fatal("expected player id and session token")
	}
	if auth == nil || res.SessionToken != auth.SessionToken || res.PlayerID != auth.PlayerID {
		t.Fatal("expected auth result to match login response")
	}
	if !player.ValidateDefaultNicknamePattern(res.Profile.DisplayName) {
		t.Fatalf("unexpected nickname: %s", res.Profile.DisplayName)
	}
	if res.IdleState == nil || res.StageState == nil {
		t.Fatal("expected idle and stage snapshots on login")
	}
	if res.Profile.CombatPower < 100 {
		t.Fatalf("expected starter combat power, got %d", res.Profile.CombatPower)
	}

	res2, auth2 := svc.GuestLogin(ctx, &apitypes.GuestLoginReq{DeviceID: "device-a"})
	if !login.IsSuccess(res2) {
		t.Fatalf("expected resume success, got %#v", res2.Error)
	}
	if res2.PlayerID != res.PlayerID {
		t.Fatalf("expected same player id, got %d and %d", res2.PlayerID, res.PlayerID)
	}
	if res2.SessionToken == res.SessionToken {
		t.Fatal("expected a new token on reconnect")
	}
	if auth2 == nil || auth2.SessionToken != res2.SessionToken {
		t.Fatal("expected new auth token")
	}
}

func TestGuestLoginRejectsEmptyDeviceID(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := newTestLoginService(t, client)

	res, _ := svc.GuestLogin(context.Background(), &apitypes.GuestLoginReq{})
	if login.IsSuccess(res) {
		t.Fatal("expected invalid request failure")
	}
	if res.Error.Code != apitypes.ErrorCodeInvalidRequest {
		t.Fatalf("unexpected error code: %v", res.Error.Code)
	}
}

func TestPhoneAuthCreatesSessionAndRejectsInvalidCode(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:phone?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := newTestLoginService(t, client)
	ctx := context.Background()

	bad, _ := svc.PhoneRegister(ctx, &apitypes.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "999999"})
	if bad.Error.Code != apitypes.ErrorCodeInvalidRequest {
		t.Fatalf("expected invalid code rejection, got %v", bad.Error.Code)
	}

	res, auth := svc.PhoneRegister(ctx, &apitypes.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "000000"})
	if apitypes.HasError(res.Error) {
		t.Fatalf("expected phone register success, got %#v", res.Error)
	}
	if res.PlayerID == 0 || res.SessionToken == "" || auth == nil || auth.SessionToken == "" {
		t.Fatal("expected player id and session token")
	}

	loginRes, loginAuth := svc.PhoneLogin(ctx, &apitypes.PhoneLoginReq{Phone: "13800000000", VerifyCode: "123456"})
	if apitypes.HasError(loginRes.Error) {
		t.Fatalf("expected phone login success, got %#v", loginRes.Error)
	}
	if loginRes.PlayerID != res.PlayerID {
		t.Fatalf("expected same player id, got %d and %d", loginRes.PlayerID, res.PlayerID)
	}
	if loginAuth == nil || loginAuth.SessionToken != loginRes.SessionToken {
		t.Fatal("expected auth result to match login response")
	}
	if loginRes.SessionToken == res.SessionToken {
		t.Fatal("expected a new session token on phone login")
	}
}
