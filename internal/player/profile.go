package player

import (
	playerv1 "github.com/slimeyquest/proto/gen/go/player"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

// ToProfile maps ProgressState into a protobuf PlayerProfile.
func ToProfile(state *ProgressState, cfg *gameplayconfig.Config) *playerv1.PlayerProfile {
	if state == nil {
		return &playerv1.PlayerProfile{}
	}
	return &playerv1.PlayerProfile{
		PlayerId:            state.PlayerID,
		DisplayName:         state.DisplayName,
		Gold:                state.Gold,
		Gems:                state.Gems,
		CombatPower:         ComputeCombatPower(state, cfg),
		AdventureId:         state.AdventureID,
		StageIndex:          state.StageIndex,
		HighestStageCleared: state.HighestStageCleared,
		EquippedSlots:       state.Equipment.EquippedSlots(),
		CreatedAtMs:         state.CreatedAt.UnixMilli(),
		LastLoginAtMs:       lastLoginAtMs(state),
		HeroClass:           playerv1.HeroClass_HERO_CLASS_WARRIOR,
		HeroLevel:           state.Level,
		ZoneId:              TestZoneID,
		ProfessionSkill:     &playerv1.SkillInfo{SkillId: 1, Name: "Warrior Strike", Quality: 1},
		EquippedSkills:      []*playerv1.SkillInfo{},
		Companions:          []*playerv1.CompanionInfo{},
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
