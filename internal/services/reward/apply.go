package reward

import (
	"context"
	"fmt"
	"log/slog"

	rewardv1 "github.com/slimeyquest/proto/gen/go/reward"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/player"
)

// Applier applies explicit reward grants to player state.
type Applier struct {
	log *slog.Logger
	cfg *gameplayconfig.Config
}

// NewApplier creates a reward applier.
func NewApplier(log *slog.Logger, cfg *gameplayconfig.Config) *Applier {
	return &Applier{log: log, cfg: cfg}
}

// Apply mutates state in memory and returns the granted bundle.
func (a *Applier) Apply(ctx context.Context, state *player.ProgressState, req ApplyRequest) (*ApplyResult, error) {
	if state == nil {
		return nil, fmt.Errorf("apply reward: nil state")
	}
	if req.Source == rewardv1.RewardSource_REWARD_SOURCE_UNSPECIFIED {
		return nil, fmt.Errorf("apply reward: unspecified source")
	}

	bundle := &rewardv1.RewardBundle{Source: req.Source}
	if req.GoldDelta > 0 {
		state.Gold += req.GoldDelta
		bundle.Items = append(bundle.Items, &rewardv1.RewardItem{
			Type: rewardv1.RewardType_REWARD_TYPE_GOLD,
			Gold: req.GoldDelta,
		})
	}

	equipCount := 0
	for _, grant := range req.EquipmentGrants {
		row := a.cfg.StarterWeapon
		if grant.ConfigID != 0 {
			row = gameplayconfig.DropRow{
				ConfigID:       grant.ConfigID,
				Rarity:         grant.Rarity,
				Slot:           grant.Slot,
				Attack:         grant.Attack,
				HP:             grant.HP,
				BonusAttackPct: grant.BonusAttackPct,
			}
		}
		inst := state.Equipment.AddInstance(row)
		equipCount++
		bundle.Items = append(bundle.Items, &rewardv1.RewardItem{
			Type:      rewardv1.RewardType_REWARD_TYPE_EQUIPMENT,
			Equipment: inst.ToProto(),
		})
	}

	a.log.InfoContext(ctx, "reward_applied",
		"player_id", state.PlayerID,
		"source", req.Source.String(),
		"gold", req.GoldDelta,
		"equipment_count", equipCount,
	)

	return &ApplyResult{
		State:         state,
		AppliedBundle: bundle,
	}, nil
}

// BundleFromGrants builds a proto bundle without applying (for previews).
func BundleFromGrants(source rewardv1.RewardSource, gold int64, instances []player.EquipmentInstance) *rewardv1.RewardBundle {
	bundle := &rewardv1.RewardBundle{Source: source}
	if gold > 0 {
		bundle.Items = append(bundle.Items, &rewardv1.RewardItem{
			Type: rewardv1.RewardType_REWARD_TYPE_GOLD,
			Gold: gold,
		})
	}
	for _, inst := range instances {
		bundle.Items = append(bundle.Items, &rewardv1.RewardItem{
			Type:      rewardv1.RewardType_REWARD_TYPE_EQUIPMENT,
			Equipment: inst.ToProto(),
		})
	}
	return bundle
}

// GrantsFromDrop builds equipment grants from config drop rows.
func GrantsFromDrop(rows []gameplayconfig.DropRow) []EquipmentGrant {
	grants := make([]EquipmentGrant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, EquipmentGrant{
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
