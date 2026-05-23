package stage

import (
	"context"
	"log/slog"

	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
)

// Service handles stage progression.
type Service struct {
	log     *slog.Logger
	cfg     *config.GameplayConfig
	players player.Repository
	rewards *reward.Service
}

// NewService creates a stage service.
func NewService(log *slog.Logger, cfg *config.GameplayConfig, players player.Repository, rewards *reward.Service) *Service {
	return &Service{log: log, cfg: cfg, players: players, rewards: rewards}
}

// BuildStageState returns the current stage snapshot for a player.
func (s *Service) BuildStageState(state *player.ProgressState) *entity.StageState {
	challengeFlat := ChallengeFlat(state.HighestStageCleared)
	rec := s.cfg.RecommendedPower(challengeFlat)
	isCleared := challengeFlat <= state.HighestStageCleared

	milestones := make([]*entity.RewardBundle, 0)
	for _, flat := range config.MilestoneFlats {
		if flat <= state.HighestStageCleared {
			continue
		}
		row, ok := s.cfg.Stage(flat)
		if !ok || row.MilestoneGold <= 0 {
			continue
		}
		if state.HasClearedMilestone(flat) {
			continue
		}
		milestones = append(milestones, reward.BundleFromGrants(
			entity.RewardSourceStageMilestone,
			row.MilestoneGold,
			nil,
		))
	}

	return &entity.StageState{
		AdventureID:            state.AdventureID,
		StageIndex:             state.StageIndex,
		HighestStageCleared:    state.HighestStageCleared,
		RecommendedCombatPower: rec,
		IsCleared:              isCleared,
		MilestoneRewards:       milestones,
	}
}

// PushStage attempts to clear the current stage target.
func (s *Service) PushStage(ctx context.Context, playerID int64, targetStageIndex int32) (*entity.PushStageRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}

	if !IsCurrentTarget(state.AdventureID, state.StageIndex, targetStageIndex) {
		return &entity.PushStageRes{
			Success:    false,
			StageState: s.BuildStageState(state),
		}, nil
	}

	challengeFlat := FlatStage(state.AdventureID, state.StageIndex)
	if !IsUnlocked(state.HighestStageCleared, challengeFlat) {
		return &entity.PushStageRes{
			Success:    false,
			StageState: s.BuildStageState(state),
		}, nil
	}

	combatPower := player.ComputeCombatPower(state, s.cfg)
	recommended := s.cfg.RecommendedPower(challengeFlat)
	if !CanClear(combatPower, recommended, s.cfg.ClearThreshold()) {
		s.log.InfoContext(ctx, "stage_push_failed",
			"player_id", playerID,
			"flat_stage", challengeFlat,
			"combat_power", combatPower,
			"recommended", recommended,
		)
		return &entity.PushStageRes{
			Success:    false,
			StageState: s.BuildStageState(state),
		}, nil
	}

	row, ok := s.cfg.Stage(challengeFlat)
	if !ok {
		return &entity.PushStageRes{
			Success:    false,
			StageState: s.BuildStageState(state),
		}, nil
	}

	firstClear := challengeFlat > state.HighestStageCleared
	var milestoneBundle *entity.RewardBundle
	var boxReward *entity.StageBoxReward
	if firstClear {
		goldReward := row.FirstClearGold
		if goldReward > 0 {
			if _, err := s.rewards.GrantInMemory(ctx, state, reward.ApplyRequest{
				PlayerID:  playerID,
				Source:    entity.RewardSourceStageClear,
				GoldDelta: goldReward,
			}); err != nil {
				return nil, err
			}
		}

		state.HighestStageCleared = challengeFlat
		nextFlat := challengeFlat + 1
		if nextFlat <= 30 {
			state.AdventureID, state.StageIndex = FromFlatStage(nextFlat)
		} else {
			state.AdventureID, state.StageIndex = FromFlatStage(30)
		}

		boxCount := s.stageBoxCount(playerID, challengeFlat)
		state.AddBoxes(boxCount)
		boxReward = &entity.StageBoxReward{
			BoxCount:      boxCount,
			TotalBoxCount: state.BoxCount(),
		}

		if config.IsMilestone(challengeFlat) && row.MilestoneGold > 0 && !state.HasClearedMilestone(challengeFlat) {
			state.MarkMilestone(challengeFlat)
			milestoneBundle = reward.BundleFromGrants(
				entity.RewardSourceStageMilestone,
				row.MilestoneGold,
				nil,
			)
			if _, err := s.rewards.GrantInMemory(ctx, state, reward.ApplyRequest{
				PlayerID:  playerID,
				Source:    entity.RewardSourceStageMilestone,
				GoldDelta: row.MilestoneGold,
			}); err != nil {
				return nil, err
			}
		}
	}

	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}

	s.log.InfoContext(ctx, "stage_push_success",
		"player_id", playerID,
		"flat_stage", challengeFlat,
	)

	return &entity.PushStageRes{
		Success:         true,
		StageState:      s.BuildStageState(state),
		MilestoneReward: milestoneBundle,
		BoxReward:       boxReward,
	}, nil
}

func (s *Service) stageBoxCount(playerID int64, flat int32) int32 {
	cfg := s.cfg.ClosedLoop
	min := cfg.StageBoxMinValue()
	max := cfg.StageBoxMaxValue()
	if min <= 0 {
		min = 1
	}
	if max < min {
		max = min
	}
	rangeSize := max - min + 1
	return min + int32((playerID+int64(flat))%int64(rangeSize))
}
