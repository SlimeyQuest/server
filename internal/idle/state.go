package idle

import (
	idlev1 "github.com/slimeyquest/proto/gen/go/idle"
	playerv1 "github.com/slimeyquest/proto/gen/go/player"
	rewardv1 "github.com/slimeyquest/proto/gen/go/reward"
)

// BuildIdleState maps preview data into protobuf IdleState.
func BuildIdleState(preview Preview, previewBundle *rewardv1.RewardBundle, profile *playerv1.PlayerProfile) *idlev1.IdleState {
	return &idlev1.IdleState{
		OfflineSeconds: preview.OfflineSeconds,
		RewardStartMs:  preview.RewardStart.UnixMilli(),
		RewardEndMs:    preview.RewardEnd.UnixMilli(),
		EfficiencyBps:  preview.EfficiencyBps,
		PreviewReward:  previewBundle,
		PlayerSnapshot: profile,
	}
}
