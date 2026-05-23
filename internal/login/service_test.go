package login_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	"github.com/slimeyquest/ent"
	"github.com/slimeyquest/ent/enttest"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/idle"
	"github.com/slimeyquest/server/internal/login"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/reward"
	"github.com/slimeyquest/server/internal/session"
	"github.com/slimeyquest/server/internal/stage"
)

type fakeConn struct {
	id       string
	playerID int64
	token    string
	closed   bool
}

func (f *fakeConn) ID() string { return f.id }
func (f *fakeConn) Close()     { f.closed = true }
func (f *fakeConn) SetAuthenticated(playerID int64, token string) {
	f.playerID = playerID
	f.token = token
}

func newTestLoginService(t *testing.T, client *ent.Client) *login.Service {
	t.Helper()
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := player.NewRepository(client, cfg)
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
	res := svc.GuestLogin(ctx, conn1, &loginv1.GuestLoginReq{
		DeviceId:      "device-a",
		ClientVersion: "1.0.0",
	})
	if !login.IsSuccess(res) {
		t.Fatalf("expected success, got %#v", res.GetError())
	}
	if res.GetPlayerId() == 0 || res.GetSessionToken() == "" {
		t.Fatal("expected player id and session token")
	}
	if res.GetSessionToken() != conn1.token {
		t.Fatalf("expected conn token %q, got %q", res.GetSessionToken(), conn1.token)
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
	res2 := svc.GuestLogin(ctx, conn2, &loginv1.GuestLoginReq{DeviceId: "device-a"})
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

	res := svc.GuestLogin(context.Background(), &fakeConn{id: "conn-1"}, &loginv1.GuestLoginReq{})
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

	bad := svc.PhoneRegister(ctx, &fakeConn{id: "bad"}, &loginv1.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "999999"})
	if bad.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("expected invalid code rejection, got %v", bad.GetError().GetCode())
	}

	conn := &fakeConn{id: "phone-1"}
	res := svc.PhoneRegister(ctx, conn, &loginv1.PhoneRegisterReq{Phone: "13800000000", VerifyCode: "000000"})
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("expected phone register success, got %#v", res.GetError())
	}
	if res.GetPlayerId() == 0 || res.GetSessionToken() == "" || conn.token == "" {
		t.Fatal("expected player id and session token")
	}

	loginConn := &fakeConn{id: "phone-2"}
	loginRes := svc.PhoneLogin(ctx, loginConn, &loginv1.PhoneLoginReq{Phone: "13800000000", VerifyCode: "123456"})
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
