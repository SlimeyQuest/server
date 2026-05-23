package network

import (
	"context"

	gatewayv1 "github.com/slimeyquest/proto/gen/go/gateway"
	idlev1 "github.com/slimeyquest/proto/gen/go/idle"
	stagev1 "github.com/slimeyquest/proto/gen/go/stage"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/stage"
)

// Gameplay handles authenticated gameplay messages.
type Gameplay struct {
	idle    *idle.Service
	stage   *stage.Service
	loop    *player.ClosedLoopService
	session SessionValidator
}

// SessionValidator checks active session tokens.
type SessionValidator interface {
	Validate(playerID int64, token string) bool
}

// NewGameplay creates a gameplay message handler.
func NewGameplay(idleSvc *idle.Service, stageSvc *stage.Service, loopSvc *player.ClosedLoopService, sessions SessionValidator) *Gameplay {
	return &Gameplay{idle: idleSvc, stage: stageSvc, loop: loopSvc, session: sessions}
}

func (g *Gameplay) handleClaimIdle(c *Conn, req *idlev1.ClaimIdleRewardsReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		c.log.Warn("gameplay_rejected", "reason", "invalid_session")
		return errInvalidSession
	}

	res, err := g.idle.Claim(context.Background(), c.PlayerID(), req.GetClaimedThroughMs())
	if err != nil {
		c.log.Error("claim_idle_failed", "player_id", c.PlayerID(), "error", err)
		return c.sendServerMessage(&gatewayv1.ServerMessage{
			Payload: &gatewayv1.ServerMessage_ClaimIdleRewards{
				ClaimIdleRewards: &idlev1.ClaimIdleRewardsRes{Success: false},
			},
		})
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{
		Payload: &gatewayv1.ServerMessage_ClaimIdleRewards{ClaimIdleRewards: res},
	})
}

func (g *Gameplay) handlePushStage(c *Conn, req *stagev1.PushStageReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		c.log.Warn("gameplay_rejected", "reason", "invalid_session")
		return errInvalidSession
	}

	res, err := g.stage.PushStage(context.Background(), c.PlayerID(), req.GetTargetStageIndex())
	if err != nil {
		c.log.Error("push_stage_failed", "player_id", c.PlayerID(), "error", err)
		return c.sendServerMessage(&gatewayv1.ServerMessage{
			Payload: &gatewayv1.ServerMessage_PushStage{
				PushStage: &stagev1.PushStageRes{Success: false},
			},
		})
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{
		Payload: &gatewayv1.ServerMessage_PushStage{PushStage: res},
	})
}
