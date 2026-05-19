package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	gatewayv1 "github.com/slimeyquest/proto/gen/go/gateway"
	idlev1 "github.com/slimeyquest/proto/gen/go/idle"
	loginv1 "github.com/slimeyquest/proto/gen/go/login"
	stagev1 "github.com/slimeyquest/proto/gen/go/stage"
)

const (
	readTimeout       = 10 * time.Second
	stage1ClearGold   = int64(50)
)

type savedState struct {
	gold         int64
	combatPower  int64
	stageCleared int32
	stageIndex   int32
}

func main() {
	addr := flag.String("addr", "ws://localhost:8080/ws", "WebSocket server address")
	flag.Parse()

	deviceID := fmt.Sprintf("smoke-%d", time.Now().UnixNano())
	fmt.Println("ws-smoke: device_id =", deviceID)

	conn, cleanup, err := dial(*addr)
	if err != nil {
		fail("connect", err)
	}
	defer cleanup()

	loginRes, err := guestLogin(conn, deviceID)
	if err != nil {
		fail("login", err)
	}
	pass("login: session_token, profile, idle_state, stage_state present")

	goldBefore := loginRes.GetProfile().GetGold()
	stageBefore := loginRes.GetStageState().GetHighestStageCleared()
	combatPower := loginRes.GetProfile().GetCombatPower()

	// New players start with last_claim_at=now; wait long enough for at least 1 gold tick.
	time.Sleep(5 * time.Second)

	claimRes, err := claimIdle(conn, time.Now().UnixMilli())
	if err != nil {
		fail("idle claim", err)
	}
	if !claimRes.GetSuccess() {
		fail("idle claim", fmt.Errorf("success=false"))
	}
	goldAfterClaim := claimRes.GetIdleState().GetPlayerSnapshot().GetGold()
	if goldAfterClaim <= goldBefore {
		fail("idle claim", fmt.Errorf("gold did not increase: before=%d after=%d", goldBefore, goldAfterClaim))
	}
	pass(fmt.Sprintf("idle claim: gold %d -> %d", goldBefore, goldAfterClaim))
	pass("connection alive after login (idle claim succeeded on same socket)")

	dupRes, err := claimIdle(conn, time.Now().UnixMilli())
	if err != nil {
		fail("duplicate idle claim", err)
	}
	goldAfterDup := dupRes.GetIdleState().GetPlayerSnapshot().GetGold()
	if goldAfterDup != goldAfterClaim {
		fail("duplicate idle claim", fmt.Errorf("gold changed: before=%d after=%d", goldAfterClaim, goldAfterDup))
	}
	pass("duplicate idle claim: no extra gold granted")

	failPush, err := pushStage(conn, 2)
	if err != nil {
		fail("stage push fail", err)
	}
	if failPush.GetSuccess() {
		fail("stage push fail", fmt.Errorf("expected success=false for wrong target"))
	}
	if failPush.GetStageState().GetHighestStageCleared() != stageBefore {
		fail("stage push fail", fmt.Errorf("stage progression mutated on failed push"))
	}
	pass("stage push fail: wrong target rejected without mutation")

	okPush, err := pushStage(conn, 1)
	if err != nil {
		fail("stage push ok", err)
	}
	if !okPush.GetSuccess() {
		fail("stage push ok", fmt.Errorf("expected success=true for stage 1"))
	}
	if okPush.GetStageState().GetHighestStageCleared() <= stageBefore {
		fail("stage push ok", fmt.Errorf("highest_stage_cleared did not advance: before=%d after=%d", stageBefore, okPush.GetStageState().GetHighestStageCleared()))
	}
	pass(fmt.Sprintf("stage push ok: cleared stage, highest=%d", okPush.GetStageState().GetHighestStageCleared()))

	state := savedState{
		gold:         goldAfterDup + stage1ClearGold,
		combatPower:  combatPower,
		stageCleared: okPush.GetStageState().GetHighestStageCleared(),
		stageIndex:   okPush.GetStageState().GetStageIndex(),
	}

	cleanup()
	time.Sleep(200 * time.Millisecond)

	conn2, cleanup2, err := dial(*addr)
	if err != nil {
		fail("reconnect", err)
	}
	defer cleanup2()

	resumeRes, err := guestLogin(conn2, deviceID)
	if err != nil {
		fail("persistence login", err)
	}

	if resumeRes.GetProfile().GetGold() != state.gold {
		fail("persistence gold", fmt.Errorf("expected %d got %d", state.gold, resumeRes.GetProfile().GetGold()))
	}
	if resumeRes.GetStageState().GetHighestStageCleared() != state.stageCleared {
		fail("persistence stage", fmt.Errorf("expected cleared=%d got %d", state.stageCleared, resumeRes.GetStageState().GetHighestStageCleared()))
	}
	if resumeRes.GetProfile().GetCombatPower() != state.combatPower {
		fail("persistence combat power", fmt.Errorf("expected %d got %d", state.combatPower, resumeRes.GetProfile().GetCombatPower()))
	}
	if resumeRes.GetIdleState().GetOfflineSeconds() < 0 {
		fail("persistence idle", fmt.Errorf("invalid idle offline_seconds"))
	}
	pass("persistence: gold, stage, idle state, combat power restored after reconnect")

	fmt.Println("\nws-smoke: ALL PASS")
}

func dial(addr string) (*websocket.Conn, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		return nil, nil, err
	}
	return conn, func() { _ = conn.Close(websocket.StatusNormalClosure, "") }, nil
}

func sendClient(conn *websocket.Conn, msg *gatewayv1.ClientMessage) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return conn.Write(ctx, websocket.MessageBinary, payload)
}

func recvServer(conn *websocket.Conn) (*gatewayv1.ServerMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	msg := &gatewayv1.ServerMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func guestLogin(conn *websocket.Conn, deviceID string) (*loginv1.GuestLoginRes, error) {
	if err := sendClient(conn, &gatewayv1.ClientMessage{
		Payload: &gatewayv1.ClientMessage_GuestLogin{
			GuestLogin: &loginv1.GuestLoginReq{
				DeviceId:      deviceID,
				ClientVersion: "ws-smoke",
			},
		},
	}); err != nil {
		return nil, err
	}
	msg, err := recvServer(conn)
	if err != nil {
		return nil, err
	}
	res := msg.GetGuestLogin()
	if res == nil {
		return nil, fmt.Errorf("expected guest_login response, got %T", msg.Payload)
	}
	if res.GetError() != nil && res.GetError().GetCode() != 0 {
		return nil, fmt.Errorf("login error: %s", res.GetError().GetMessage())
	}
	if res.GetSessionToken() == "" {
		return nil, fmt.Errorf("missing session_token")
	}
	if res.GetProfile() == nil {
		return nil, fmt.Errorf("missing profile")
	}
	if res.GetIdleState() == nil {
		return nil, fmt.Errorf("missing idle_state")
	}
	if res.GetStageState() == nil {
		return nil, fmt.Errorf("missing stage_state")
	}
	return res, nil
}

func claimIdle(conn *websocket.Conn, claimedThroughMs int64) (*idlev1.ClaimIdleRewardsRes, error) {
	if err := sendClient(conn, &gatewayv1.ClientMessage{
		Payload: &gatewayv1.ClientMessage_ClaimIdleRewards{
			ClaimIdleRewards: &idlev1.ClaimIdleRewardsReq{
				ClaimedThroughMs: claimedThroughMs,
			},
		},
	}); err != nil {
		return nil, err
	}
	msg, err := recvServer(conn)
	if err != nil {
		return nil, err
	}
	res := msg.GetClaimIdleRewards()
	if res == nil {
		return nil, fmt.Errorf("expected claim_idle_rewards response, got %T", msg.Payload)
	}
	return res, nil
}

func pushStage(conn *websocket.Conn, target int32) (*stagev1.PushStageRes, error) {
	if err := sendClient(conn, &gatewayv1.ClientMessage{
		Payload: &gatewayv1.ClientMessage_PushStage{
			PushStage: &stagev1.PushStageReq{
				TargetStageIndex: target,
			},
		},
	}); err != nil {
		return nil, err
	}
	msg, err := recvServer(conn)
	if err != nil {
		return nil, err
	}
	res := msg.GetPushStage()
	if res == nil {
		return nil, fmt.Errorf("expected push_stage response, got %T", msg.Payload)
	}
	return res, nil
}

func pass(msg string) {
	fmt.Println("PASS:", msg)
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "FAIL [%s]: %v\n", step, err)
	os.Exit(1)
}
