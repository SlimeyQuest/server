package idle

import (
	"context"
	"time"

	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/services/player"
)

// BuildIdleState maps preview data into an API idle state.
func BuildIdleState(preview Preview, bundle *entity.RewardBundle, profile *entity.PlayerProfile) *entity.IdleState {
	return &entity.IdleState{
		OfflineSeconds: preview.OfflineSeconds,
		RewardStartMs:  preview.RewardStart.UnixMilli(),
		RewardEndMs:    preview.RewardEnd.UnixMilli(),
		EfficiencyBps:  preview.EfficiencyBps,
		PreviewReward:  bundle,
		PlayerSnapshot: profile,
	}
}

// PreviewForLogin builds idle state without advancing last_claim_at.
func (s *Service) PreviewForLogin(_ context.Context, state *player.ProgressState, now time.Time) *entity.IdleState {
	preview := ComputePreview(state, s.cfg, now)
	bundle := PreviewBundle(preview)
	profile := player.ToProfile(state, s.cfg)
	return BuildIdleState(preview, bundle, profile)
}
