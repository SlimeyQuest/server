package reward

import (
	"github.com/slimeyquest/server/internal/apitypes"
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
	Source          string
	GoldDelta       int64
	EquipmentGrants []EquipmentGrant
}

// ApplyResult is the updated player state after applying rewards.
type ApplyResult struct {
	State         *player.ProgressState
	AppliedBundle *apitypes.RewardBundle
}
