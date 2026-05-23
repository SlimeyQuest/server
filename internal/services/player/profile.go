package player

import (
	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

// ToProfile maps ProgressState into an API player profile.
func ToProfile(state *ProgressState, cfg *gameplayconfig.Config) *apitypes.PlayerProfile {
	if state == nil {
		return &apitypes.PlayerProfile{}
	}
	return &apitypes.PlayerProfile{
		PlayerID:            state.PlayerID,
		DisplayName:         state.DisplayName,
		Gold:                state.Gold,
		Gems:                state.Gems,
		CombatPower:         ComputeCombatPower(state, cfg),
		AdventureID:         state.AdventureID,
		StageIndex:          state.StageIndex,
		HighestStageCleared: state.HighestStageCleared,
		EquippedSlots:       state.Equipment.EquippedSlotsAPI(),
		CreatedAtMs:         state.CreatedAt.UnixMilli(),
		LastLoginAtMs:       lastLoginAtMs(state),
		HeroClass:           apitypes.HeroClassWarrior,
		HeroLevel:           state.Level,
		ZoneID:              TestZoneID,
		ProfessionSkill:     &apitypes.SkillInfo{SkillID: 1, Name: "Warrior Strike", Quality: 1},
		EquippedSkills:      []apitypes.SkillInfo{},
		Companions:          []apitypes.CompanionInfo{},
		ChestLevel:          state.ChestLevel(),
		BoxCount:            state.BoxCount(),
	}
}

func lastLoginAtMs(state *ProgressState) int64 {
	if state.LastLoginAt != nil {
		return state.LastLoginAt.UnixMilli()
	}
	return state.CreatedAt.UnixMilli()
}
