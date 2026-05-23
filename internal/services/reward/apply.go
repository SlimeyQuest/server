package reward

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/player"
)

// Applier applies explicit reward grants to player state.
type Applier struct {
	log *slog.Logger
	cfg *config.GameplayConfig
}

// NewApplier creates a reward applier.
func NewApplier(log *slog.Logger, cfg *config.GameplayConfig) *Applier {
	return &Applier{log: log, cfg: cfg}
}

// Apply mutates state in memory and returns the granted bundle.
func (a *Applier) Apply(ctx context.Context, state *player.ProgressState, req ApplyRequest) (*ApplyResult, error) {
	if state == nil {
		return nil, fmt.Errorf("apply reward: nil state")
	}
	if req.Source == "" {
		return nil, fmt.Errorf("apply reward: unspecified source")
	}

	bundle := &entity.RewardBundle{Source: req.Source}
	if req.GoldDelta > 0 {
		state.Gold += req.GoldDelta
		bundle.Items = append(bundle.Items, entity.RewardItem{
			Type: entity.RewardTypeGold,
			Gold: req.GoldDelta,
		})
	}

	equipCount := 0
	for _, grant := range req.EquipmentGrants {
		row := a.cfg.StarterWeapon
		if grant.ConfigID != 0 {
			row = config.DropRow{
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
		eq := inst.ToAPI()
		bundle.Items = append(bundle.Items, entity.RewardItem{
			Type:      entity.RewardTypeEquipment,
			Equipment: &eq,
		})
	}

	a.log.InfoContext(ctx, "reward_applied",
		"player_id", state.PlayerID,
		"source", req.Source,
		"gold", req.GoldDelta,
		"equipment_count", equipCount,
	)

	return &ApplyResult{
		State:         state,
		AppliedBundle: bundle,
	}, nil
}

// BundleFromGrants builds a bundle without applying (for previews).
func BundleFromGrants(source string, gold int64, instances []player.EquipmentInstance) *entity.RewardBundle {
	bundle := &entity.RewardBundle{Source: source}
	if gold > 0 {
		bundle.Items = append(bundle.Items, entity.RewardItem{
			Type: entity.RewardTypeGold,
			Gold: gold,
		})
	}
	for _, inst := range instances {
		eq := inst.ToAPI()
		bundle.Items = append(bundle.Items, entity.RewardItem{
			Type:      entity.RewardTypeEquipment,
			Equipment: &eq,
		})
	}
	return bundle
}

// GrantsFromDrop builds equipment grants from config drop rows.
func GrantsFromDrop(rows []config.DropRow) []EquipmentGrant {
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
