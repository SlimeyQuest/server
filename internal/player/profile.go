package player

import (
	playerv1 "github.com/slimeyquest/proto/gen/go/player"
	"github.com/slimeyquest/server/ent"
)

// ToProfile maps an ent Player into a protobuf PlayerProfile.
func ToProfile(p *ent.Player) *playerv1.PlayerProfile {
	return &playerv1.PlayerProfile{
		PlayerId:            int64(p.ID),
		DisplayName:         p.Nickname,
		Gold:                0,
		Gems:                0,
		CombatPower:         0,
		AdventureId:         0,
		StageIndex:          0,
		HighestStageCleared: 0,
		EquippedSlots:       nil,
		CreatedAtMs:         p.CreatedAt.UnixMilli(),
		LastLoginAtMs:       lastLoginAtMs(p),
	}
}

func lastLoginAtMs(p *ent.Player) int64 {
	if p.LastLoginAt != nil {
		return p.LastLoginAt.UnixMilli()
	}
	// MVP fallback for rows without login history yet.
	return p.CreatedAt.UnixMilli()
}
