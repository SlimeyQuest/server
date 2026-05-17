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
	"github.com/slimeyquest/server/ent/enttest"
	"github.com/slimeyquest/server/internal/login"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/session"
)

type fakeConn struct {
	id        string
	playerID  int64
	token     string
	closed    bool
}

func (f *fakeConn) ID() string { return f.id }
func (f *fakeConn) Close()     { f.closed = true }
func (f *fakeConn) SetAuthenticated(playerID int64, token string) {
	f.playerID = playerID
	f.token = token
}

func TestGuestLoginCreatesAndResumesPlayer(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := player.NewRepository(client)
	sessions := session.NewManager()
	svc := login.NewService(log, repo, sessions)
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

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := login.NewService(log, player.NewRepository(client), session.NewManager())

	res := svc.GuestLogin(context.Background(), &fakeConn{id: "conn-1"}, &loginv1.GuestLoginReq{})
	if login.IsSuccess(res) {
		t.Fatal("expected invalid request failure")
	}
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("unexpected error code: %v", res.GetError().GetCode())
	}
}
