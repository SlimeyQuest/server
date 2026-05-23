package idle

import (
	"context"
	"log/slog"
	"time"

	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
)

const claimTimeToleranceMs = 120_000

// Service handles idle reward preview and claims.
type Service struct {
	log     *slog.Logger
	cfg     *gameplayconfig.Config
	players player.Repository
	rewards *reward.Service
}

// NewService creates an idle service.
func NewService(log *slog.Logger, cfg *gameplayconfig.Config, players player.Repository, rewards *reward.Service) *Service {
	return &Service{log: log, cfg: cfg, players: players, rewards: rewards}
}

// Claim settles idle rewards for the player.
func (s *Service) Claim(ctx context.Context, playerID int64, claimedThroughMs int64) (*apitypes.ClaimIdleRewardsRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if claimedThroughMs > 0 {
		claimedAt := time.UnixMilli(claimedThroughMs).UTC()
		diff := now.Sub(claimedAt)
		if diff < 0 {
			diff = -diff
		}
		if diff > claimTimeToleranceMs {
			s.log.WarnContext(ctx, "idle_claim_time_skew",
				"player_id", playerID,
				"claimed_through_ms", claimedThroughMs,
				"server_now_ms", now.UnixMilli(),
			)
		}
	}

	preview := ComputePreview(state, s.cfg, now)
	grants := ClaimGrants(state, s.cfg, preview)
	applyReq := reward.ApplyRequest{
		PlayerID:        playerID,
		Source:          apitypes.RewardSourceIdleClaim,
		GoldDelta:       preview.GoldTotal,
		EquipmentGrants: grants,
	}
	result, err := s.rewards.Grant(ctx, applyReq)
	if err != nil {
		return nil, err
	}

	claimNow := now
	result.State.LastClaimAt = &claimNow
	if err := s.players.SaveProgress(ctx, result.State); err != nil {
		return nil, err
	}

	profile := player.ToProfile(result.State, s.cfg)
	freshPreview := ComputePreview(result.State, s.cfg, now)
	idleState := BuildIdleState(freshPreview, PreviewBundle(freshPreview), profile)

	return &apitypes.ClaimIdleRewardsRes{
		Success:       true,
		ClaimedReward: result.AppliedBundle,
		IdleState:     idleState,
	}, nil
}
