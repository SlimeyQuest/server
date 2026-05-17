package network

import (
	"context"

	"google.golang.org/protobuf/proto"

	idlev1 "github.com/slimeyquest/proto/gen/go/idle"
	stagev1 "github.com/slimeyquest/proto/gen/go/stage"
	"github.com/slimeyquest/server/internal/idle"
	"github.com/slimeyquest/server/internal/stage"
)

// Gameplay handles authenticated gameplay messages.
type Gameplay struct {
	idle    *idle.Service
	stage   *stage.Service
	session SessionValidator
}

// SessionValidator checks active session tokens.
type SessionValidator interface {
	Validate(playerID int64, token string) bool
}

// NewGameplay creates a gameplay message handler.
func NewGameplay(idleSvc *idle.Service, stageSvc *stage.Service, sessions SessionValidator) *Gameplay {
	return &Gameplay{idle: idleSvc, stage: stageSvc, session: sessions}
}

func (g *Gameplay) handleMessage(c *Conn, data []byte) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		c.log.Warn("gameplay_rejected", "reason", "invalid_session")
		return errInvalidSession
	}

	kind, field := peekInt64Field(data)
	switch kind {
	case wireKindIdleClaim:
		req := &idlev1.ClaimIdleRewardsReq{ClaimedThroughMs: field}
		res, err := g.idle.Claim(context.Background(), c.PlayerID(), req.GetClaimedThroughMs())
		if err != nil {
			c.log.Error("claim_idle_failed", "player_id", c.PlayerID(), "error", err)
			return c.sendProto(&idlev1.ClaimIdleRewardsRes{Success: false})
		}
		return c.sendProto(res)
	case wireKindPushStage:
		req := &stagev1.PushStageReq{TargetStageIndex: int32(field)}
		res, err := g.stage.PushStage(context.Background(), c.PlayerID(), req.GetTargetStageIndex())
		if err != nil {
			c.log.Error("push_stage_failed", "player_id", c.PlayerID(), "error", err)
			return c.sendProto(&stagev1.PushStageRes{Success: false})
		}
		return c.sendProto(res)
	default:
		c.log.Warn("gameplay_rejected", "reason", "unknown_message", "field", field)
		return errUnknownMessage
	}
}

type wireKind int

const (
	wireKindUnknown wireKind = iota
	wireKindIdleClaim
	wireKindPushStage
)

// peekInt64Field reads protobuf field 1 varint and classifies the message.
// Claim idle uses millisecond timestamps; push stage uses small stage indexes.
func peekInt64Field(data []byte) (wireKind, int64) {
	var fieldNum int
	var value int64
	var shift uint
	for i := 0; i < len(data); i++ {
		b := data[i]
		if fieldNum == 0 {
			key := b
			fieldNum = int(key >> 3)
			if fieldNum != 1 {
				return wireKindUnknown, 0
			}
			if key&0x07 != 0 {
				return wireKindUnknown, 0
			}
			continue
		}
		value |= int64(b&0x7f) << shift
		if b < 0x80 {
			break
		}
		shift += 7
	}
	if fieldNum != 1 {
		return wireKindUnknown, 0
	}
	if value >= 1_000_000_000_000 {
		return wireKindIdleClaim, value
	}
	if value >= 1 && value <= 10 {
		return wireKindPushStage, value
	}
	return wireKindUnknown, value
}

func (c *Conn) sendProto(msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		c.log.Error("marshal gameplay response failed", "error", err)
		return err
	}
	c.sendResponse(payload)
	return nil
}
