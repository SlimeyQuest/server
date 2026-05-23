package idle

import (
	"context"
	"time"

	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/services/player"
)

// BuildIdleState maps preview data into an API idle state.
func BuildIdleState(preview Preview, bundle *apitypes.RewardBundle, profile *apitypes.PlayerProfile) *apitypes.IdleState {
	return &apitypes.IdleState{
		OfflineSeconds: preview.OfflineSeconds,
		RewardStartMs:  preview.RewardStart.UnixMilli(),
		RewardEndMs:    preview.RewardEnd.UnixMilli(),
		EfficiencyBps:  preview.EfficiencyBps,
		PreviewReward:  bundle,
		PlayerSnapshot: profile,
	}
}

// PreviewForLogin builds idle state without advancing last_claim_at.
func (s *Service) PreviewForLogin(_ context.Context, state *player.ProgressState, now time.Time) *apitypes.IdleState {
	preview := ComputePreview(state, s.cfg, now)
	bundle := PreviewBundle(preview)
	profile := player.ToProfile(state, s.cfg)
	return BuildIdleState(preview, bundle, profile)
}
