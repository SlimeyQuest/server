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
	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	"github.com/slimeyquest/server/internal/data/playerrepo"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/login"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
	"github.com/slimeyquest/server/internal/services/session"
	"github.com/slimeyquest/server/internal/services/stage"
)

type fakeConn struct {
	id     string
	closed bool
}

func (f *fakeConn) Close() { f.closed = true }

func binding(conn *fakeConn) login.SessionBinding {
	return login.SessionBinding{ID: conn.id, Handle: conn}
}

func applyAuth(conn *fakeConn, auth *login.AuthResult) {
	if auth == nil || auth.ReplacedBinding == nil {
		return
	}
	if oldConn, ok := auth.ReplacedBinding.Handle.(*fakeConn); ok {
		oldConn.Close()
	}
}

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

	conn1 := &fakeConn{id: "conn-1"}
	res, auth := svc.GuestLogin(ctx, binding(conn1), &loginv1.GuestLoginReq{
		DeviceId:      "device-a",
		ClientVersion: "1.0.0",
	})
	applyAuth(conn1, auth)
	if !login.IsSuccess(res) {
		t.Fatalf("expected success, got %#v", res.GetError())
	}
	if res.GetPlayerId() == 0 || res.GetSessionToken() == "" {
		t.Fatal("expected player id and session token")
	}
	if auth == nil || res.GetSessionToken() != auth.SessionToken || res.GetPlayerId() != auth.PlayerID {
		t.Fatal("expected auth result to match login response")
	}
	if !player.ValidateDefaultNicknamePattern(res.GetProfile().GetDisplayName()) {
		t.Fatalf("unexpected nickname: %s", res.GetProfile().GetDisplayName())
	}
	if res.GetIdleState() == nil || res.GetStageState() == nil {
		t.Fatal("expected idle and stage snapshots on login")
	}
	if res.GetProfile().GetCombatPower() < 100 {
		t.Fatalf("expected starter combat power, got %d", res.GetProfile().GetCombatPower())
	}

	conn2 := &fakeConn{id: "conn-2"}
	res2, auth2 := svc.GuestLogin(ctx, binding(conn2), &loginv1.GuestLoginReq{DeviceId: "device-a"})
	applyAuth(conn2, auth2)
	if !login.IsSuccess(res2) {
		t.Fatalf("expected resume success, got %#v", res2.GetError())
	}
	if res2.GetPlayerId() != res.GetPlayerId() {
		t.Fatalf("expected same player id, got %d and %d", res2.GetPlayerId(), res.GetPlayerId())
	}
	if res2.GetSessionToken() == res.GetSessionToken() {
		t.Fatal("expected a new token on reconnect")
	}
	if !conn1.closed {
		t.Fatal("expected previous connection to be closed on replacement")
	}
}

func TestGuestLoginRejectsEmptyDeviceID(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := newTestLoginService(t, client)

	res, _ := svc.GuestLogin(context.Background(), binding(&fakeConn{id: "conn-1"}), &loginv1.GuestLoginReq{})
	if login.IsSuccess(res) {
		t.Fatal("expected invalid request failure")
	}
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("unexpected error code: %v", res.GetError().GetCode())
	}
}

func TestPhoneAuthCreatesSessionAndRejectsInvalidCode(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:phone?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := newTestLoginService(t, client)
	ctx := context.Background()

	bad, _ := svc.PhoneRegister(ctx, binding(&fakeConn{id: "bad"}), &loginv1.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "999999"})
	if bad.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("expected invalid code rejection, got %v", bad.GetError().GetCode())
	}

	conn := &fakeConn{id: "phone-1"}
	res, auth := svc.PhoneRegister(ctx, binding(conn), &loginv1.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "000000"})
	applyAuth(conn, auth)
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("expected phone register success, got %#v", res.GetError())
	}
	if res.GetPlayerId() == 0 || res.GetSessionToken() == "" || auth == nil || auth.SessionToken == "" {
		t.Fatal("expected player id and session token")
	}

	loginConn := &fakeConn{id: "phone-2"}
	loginRes, loginAuth := svc.PhoneLogin(ctx, binding(loginConn), &loginv1.PhoneLoginReq{Phone: "13800000000", VerifyCode: "123456"})
	applyAuth(loginConn, loginAuth)
	if loginRes.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("expected phone login success, got %#v", loginRes.GetError())
	}
	if loginRes.GetPlayerId() != res.GetPlayerId() {
		t.Fatalf("expected same player id, got %d and %d", loginRes.GetPlayerId(), res.GetPlayerId())
	}
	if !conn.closed {
		t.Fatal("expected previous phone connection to be replaced")
	}
}
