package reward

import (
	"context"
	"fmt"
	"log/slog"

	rewardv1 "github.com/slimeyquest/proto/gen/go/reward"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/player"
)

// Service grants rewards and persists player state.
type Service struct {
	log      *slog.Logger
	cfg      *gameplayconfig.Config
	players  *player.Repository
	applier  *Applier
}

// NewService creates a reward service.
func NewService(log *slog.Logger, cfg *gameplayconfig.Config, players *player.Repository) *Service {
	return &Service{
		log:     log,
		cfg:     cfg,
		players: players,
		applier: NewApplier(log, cfg),
	}
}

// Grant applies an explicit apply request and saves progress.
func (s *Service) Grant(ctx context.Context, req ApplyRequest) (*ApplyResult, error) {
	state, err := s.players.LoadProgress(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}
	result, err := s.applier.Apply(ctx, state, req)
	if err != nil {
		return nil, err
	}
	if err := s.players.SaveProgress(ctx, result.State); err != nil {
		return nil, err
	}
	return result, nil
}

// GrantInMemory applies rewards to an in-memory state without saving.
func (s *Service) GrantInMemory(ctx context.Context, state *player.ProgressState, req ApplyRequest) (*ApplyResult, error) {
	return s.applier.Apply(ctx, state, req)
}

// GrantBundle maps a RewardBundle proto into ApplyRequest and grants it.
func (s *Service) GrantBundle(ctx context.Context, playerID int64, bundle *rewardv1.RewardBundle) (*ApplyResult, error) {
	if bundle == nil {
		return nil, fmt.Errorf("grant bundle: nil bundle")
	}
	req := ApplyRequest{
		PlayerID: playerID,
		Source:   bundle.GetSource(),
	}
	for _, item := range bundle.GetItems() {
		switch item.GetType() {
		case rewardv1.RewardType_REWARD_TYPE_GOLD:
			req.GoldDelta += item.GetGold()
		case rewardv1.RewardType_REWARD_TYPE_EQUIPMENT:
			eq := item.GetEquipment()
			if eq == nil {
				continue
			}
			stats := eq.GetStats()
			grant := EquipmentGrant{
				ConfigID: eq.GetConfigId(),
				Slot:     int32(eq.GetSlot()),
				Rarity:   int32(eq.GetRarity()),
			}
			if stats != nil {
				grant.Attack = stats.GetAttack()
				grant.HP = stats.GetHp()
				grant.BonusAttackPct = stats.GetBonusAttackPct()
			}
			req.EquipmentGrants = append(req.EquipmentGrants, grant)
		}
	}
	return s.Grant(ctx, req)
}
