package player

import "github.com/slimeyquest/server/internal/gameplayconfig"

// ComputeCombatPower returns deterministic combat power from level and equipment.
func ComputeCombatPower(state *ProgressState, cfg *gameplayconfig.Config) int64 {
	if state == nil || cfg == nil {
		return 0
	}
	var weaponAtk, armorHP, ringBonus int64
	for slot, uid := range state.Equipment.Equipped {
		if uid == 0 {
			continue
		}
		inst, ok := state.Equipment.Instances[uid]
		if !ok {
			continue
		}
		switch slot {
		case 1: // weapon
			weaponAtk += inst.Attack
		case 2: // armor
			armorHP += inst.HP
		case 3: // ring
			ringBonus += inst.Attack
			if inst.BonusAttackPct > 0 && weaponAtk > 0 {
				ringBonus += weaponAtk * int64(inst.BonusAttackPct) / 10000
			}
		}
	}
	levelBonus := int64(state.Level-1) * 5
	power := weaponAtk + int64(float64(armorHP)*cfg.Globals.KArmor) + ringBonus + levelBonus
	if power < 0 {
		return 0
	}
	return power
}
