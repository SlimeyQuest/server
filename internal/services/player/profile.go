package player

import (
	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/config"
)

// ToProfile maps ProgressState into an API player profile.
func ToProfile(state *ProgressState, cfg *config.GameplayConfig) *entity.PlayerProfile {
	if state == nil {
		return &entity.PlayerProfile{}
	}
	return &entity.PlayerProfile{
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
		HeroClass:           entity.HeroClassWarrior,
		HeroLevel:           state.Level,
		ZoneID:              TestZoneID,
		ProfessionSkill:     &entity.SkillInfo{SkillID: 1, Name: "Warrior Strike", Quality: 1},
		EquippedSkills:      []entity.SkillInfo{},
		Companions:          []entity.CompanionInfo{},
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
