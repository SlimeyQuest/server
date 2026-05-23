package player

import (
	"time"

	"github.com/slimeyquest/ent"
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

// ChestLevel returns the fixed opener area's MVP upgrade level.
func (s *ProgressState) ChestLevel() int32 {
	if s.Level <= 0 {
		return 1
	}
	return s.Level
}

// SetChestLevel stores the fixed opener area's MVP upgrade level.
func (s *ProgressState) SetChestLevel(level int32) {
	if level <= 0 {
		level = 1
	}
	s.Level = level
}

// BoxCount returns unopened stage boxes stored in the bottom opener area.
func (s *ProgressState) BoxCount() int32 {
	if s.Exp <= 0 {
		return 0
	}
	return int32(s.Exp)
}

// SetBoxCount stores unopened stage boxes in the MVP temporary exp field.
func (s *ProgressState) SetBoxCount(count int32) {
	if count < 0 {
		count = 0
	}
	s.Exp = int64(count)
}

// AddBoxes increases unopened stage boxes.
func (s *ProgressState) AddBoxes(count int32) {
	if count <= 0 {
		return
	}
	s.SetBoxCount(s.BoxCount() + count)
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
