package reward

import (
	rewardv1 "github.com/slimeyquest/proto/gen/go/reward"
	"github.com/slimeyquest/server/internal/services/player"
)

// EquipmentGrant is an explicit equipment reward entry.
type EquipmentGrant struct {
	ConfigID       int32
	Slot           int32
	Rarity         int32
	Attack         int64
	HP             int64
	BonusAttackPct int32
}

// ApplyRequest describes a server-authoritative reward application.
type ApplyRequest struct {
	PlayerID        int64
	Source          rewardv1.RewardSource
	GoldDelta       int64
	EquipmentGrants []EquipmentGrant
}

// ApplyResult is the updated player state after applying rewards.
type ApplyResult struct {
	State         *player.ProgressState
	AppliedBundle *rewardv1.RewardBundle
}
