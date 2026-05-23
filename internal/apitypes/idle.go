package apitypes

// IdleState describes offline reward preview state.
type IdleState struct {
	OfflineSeconds  int64          `json:"offlineSeconds"`
	RewardStartMs   int64          `json:"rewardStartMs"`
	RewardEndMs     int64          `json:"rewardEndMs"`
	EfficiencyBps   int32          `json:"efficiencyBps"`
	PreviewReward   *RewardBundle  `json:"previewReward,omitempty"`
	PlayerSnapshot  *PlayerProfile `json:"playerSnapshot,omitempty"`
}

// ClaimIdleRewardsReq requests an idle claim.
type ClaimIdleRewardsReq struct {
	ClaimedThroughMs int64 `json:"claimedThroughMs"`
}

// ClaimIdleRewardsRes returns claim results.
type ClaimIdleRewardsRes struct {
	Success       bool          `json:"success"`
	ClaimedReward *RewardBundle `json:"claimedReward,omitempty"`
	IdleState     *IdleState    `json:"idleState,omitempty"`
}
