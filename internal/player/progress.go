package player

import (
	"time"

	"github.com/slimeyquest/server/ent"
)

// ProgressState is the server-side gameplay snapshot for one player.
type ProgressState struct {
	PlayerID            int64
	Platform            string
	ExternalID          string
	DisplayName         string
	Level               int32
	Exp                 int64
	Gold                int64
	Gems                int64
	AdventureID         int32
	StageIndex          int32
	HighestStageCleared int32
	LastClaimAt         *time.Time
	LastLoginAt         *time.Time
	CreatedAt           time.Time
	Equipment           EquipmentData
	ClearedMilestones   []int32
}

// FromEntity maps ent Player into ProgressState.
func FromEntity(p *ent.Player) *ProgressState {
	state := &ProgressState{
		PlayerID:            int64(p.ID),
		Platform:            p.Platform,
		ExternalID:          p.ExternalID,
		DisplayName:         p.Nickname,
		Level:               p.Level,
		Exp:                 p.Exp,
		Gold:                p.Gold,
		Gems:                p.Gems,
		AdventureID:         p.AdventureID,
		StageIndex:          p.StageIndex,
		HighestStageCleared: p.HighestStageCleared,
		LastClaimAt:         p.LastClaimAt,
		LastLoginAt:         p.LastLoginAt,
		CreatedAt:           p.CreatedAt,
		ClearedMilestones:   append([]int32(nil), p.ClearedMilestones...),
	}
	state.Equipment = DecodeEquipment(p.EquipmentJSON)
	return state
}

// ClaimBaseline returns the timestamp used for idle accumulation.
func (s *ProgressState) ClaimBaseline() time.Time {
	if s.LastClaimAt != nil {
		return *s.LastClaimAt
	}
	return s.CreatedAt
}

// HasClearedMilestone reports whether a milestone flat was already rewarded.
func (s *ProgressState) HasClearedMilestone(flat int32) bool {
	for _, m := range s.ClearedMilestones {
		if m == flat {
			return true
		}
	}
	return false
}

// MarkMilestone records a milestone flat as cleared.
func (s *ProgressState) MarkMilestone(flat int32) {
	if s.HasClearedMilestone(flat) {
		return
	}
	s.ClearedMilestones = append(s.ClearedMilestones, flat)
}
