package idle

import (
	"math"
	"time"

	rewardv1 "github.com/slimeyquest/proto/gen/go/reward"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/reward"
)

// Preview holds lazily computed idle rewards.
type Preview struct {
	OfflineSeconds int64
	RewardStart    time.Time
	RewardEnd      time.Time
	EfficiencyBps  int32
	GoldTotal      int64
	EquipmentRolls int64
}

// AccumulatedSeconds caps elapsed seconds since last claim.
func AccumulatedSeconds(lastClaim, now time.Time, capHours float64) int64 {
	if now.Before(lastClaim) {
		return 0
	}
	raw := int64(now.Sub(lastClaim).Seconds())
	capSec := int64(capHours * 3600)
	if capSec > 0 && raw > capSec {
		return capSec
	}
	if raw < 0 {
		return 0
	}
	return raw
}

// ComputePreview calculates idle rewards without mutating state.
func ComputePreview(state *player.ProgressState, cfg *gameplayconfig.Config, now time.Time) Preview {
	start := state.ClaimBaseline()
	end := now
	seconds := AccumulatedSeconds(start, end, cfg.Globals.OfflineCapHours)
	scale := cfg.StageIdleScale(state.HighestStageCleared)
	goldPerSec := cfg.GoldPerSec(state.HighestStageCleared)
	effectiveRate := cfg.Globals.BaseOfflineRate * scale
	goldTotal := int64(math.Floor(goldPerSec * float64(seconds) * effectiveRate))

	rollInterval := cfg.Globals.EquipRollIntervalSec
	if rollInterval <= 0 {
		rollInterval = 120
	}
	rolls := int64(0)
	if rollInterval > 0 {
		rolls = seconds / rollInterval
	}
	if max := cfg.Globals.MaxEquipRollsPerClaim; max > 0 && rolls > max {
		rolls = max
	}

	effBps := int32(scale * 10000)
	if effBps < 0 {
		effBps = 0
	}

	return Preview{
		OfflineSeconds: seconds,
		RewardStart:    start,
		RewardEnd:      end,
		EfficiencyBps:  effBps,
		GoldTotal:      goldTotal,
		EquipmentRolls: rolls,
	}
}

// PreviewBundle builds a read-only gold preview bundle for login UI.
func PreviewBundle(preview Preview) *rewardv1.RewardBundle {
	return reward.BundleFromGrants(rewardv1.RewardSource_REWARD_SOURCE_IDLE_CLAIM, preview.GoldTotal, nil)
}

// ClaimGrants builds explicit equipment grants for a claim.
func ClaimGrants(state *player.ProgressState, cfg *gameplayconfig.Config, preview Preview) []reward.EquipmentGrant {
	grants := make([]reward.EquipmentGrant, 0, preview.EquipmentRolls)
	for i := int64(0); i < preview.EquipmentRolls; i++ {
		seed := state.PlayerID*1000 + preview.OfflineSeconds + i
		row := cfg.PickIdleDrop(seed)
		grants = append(grants, reward.EquipmentGrant{
			ConfigID:       row.ConfigID,
			Slot:           row.Slot,
			Rarity:         row.Rarity,
			Attack:         row.Attack,
			HP:             row.HP,
			BonusAttackPct: row.BonusAttackPct,
		})
	}
	return grants
}
