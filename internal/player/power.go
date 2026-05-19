package player

import "github.com/slimeyquest/server/internal/gameplayconfig"

// ComputeCombatPower returns deterministic combat power from level and equipment.
func ComputeCombatPower(state *ProgressState, cfg *gameplayconfig.Config) int64 {
	if state == nil || cfg == nil {
		return 0
	}
	var attack, hp, bonusAttackPct int64
	for _, uid := range state.Equipment.Equipped {
		if uid == 0 {
			continue
		}
		inst, ok := state.Equipment.Instances[uid]
		if !ok {
			continue
		}
		attack += inst.Attack
		hp += inst.HP
		if inst.BonusAttackPct > 0 {
			bonusAttackPct += int64(inst.BonusAttackPct)
		}
	}
	attack += attack * bonusAttackPct / 10000
	levelBonus := int64(state.Level-1) * 5
	power := attack + int64(float64(hp)*cfg.Globals.KArmor) + levelBonus
	if power < 0 {
		return 0
	}
	return power
}
