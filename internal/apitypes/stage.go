package apitypes

// StageState is the current stage progression snapshot.
type StageState struct {
	AdventureID            int32           `json:"adventureId"`
	StageIndex             int32           `json:"stageIndex"`
	HighestStageCleared    int32           `json:"highestStageCleared"`
	RecommendedCombatPower int64           `json:"recommendedCombatPower"`
	IsCleared              bool            `json:"isCleared"`
	MilestoneRewards       []*RewardBundle `json:"milestoneRewards,omitempty"`
}

// PushStageReq attempts to clear the current stage target.
type PushStageReq struct {
	TargetStageIndex int32 `json:"targetStageIndex"`
}

// PushStageRes returns stage push outcome.
type PushStageRes struct {
	Success         bool            `json:"success"`
	StageState      *StageState     `json:"stageState,omitempty"`
	MilestoneReward *RewardBundle   `json:"milestoneReward,omitempty"`
	BoxReward       *StageBoxReward `json:"boxReward,omitempty"`
}
