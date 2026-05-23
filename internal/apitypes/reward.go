package apitypes

// Reward source identifiers.
const (
	RewardSourceIdleClaim      = "IDLE_CLAIM"
	RewardSourceStageClear     = "STAGE_CLEAR"
	RewardSourceStageMilestone = "STAGE_MILESTONE"
)

// Reward type identifiers.
const (
	RewardTypeGold       = "GOLD"
	RewardTypeEquipment  = "EQUIPMENT"
)

// RewardItem is one grant entry in a bundle.
type RewardItem struct {
	Type      string         `json:"type"`
	Gold      int64          `json:"gold,omitempty"`
	Equipment *EquipmentInfo `json:"equipment,omitempty"`
}

// RewardBundle groups grants from one gameplay source.
type RewardBundle struct {
	Source string       `json:"source"`
	Items  []RewardItem `json:"items,omitempty"`
}
