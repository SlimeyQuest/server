package reward

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/player"
)

// Service grants rewards and persists player state.
type Service struct {
	log     *slog.Logger
	cfg     *config.GameplayConfig
	players player.Repository
	applier *Applier
}

// NewService creates a reward service.
func NewService(log *slog.Logger, cfg *config.GameplayConfig, players player.Repository) *Service {
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

// GrantBundle maps a reward bundle into ApplyRequest and grants it.
func (s *Service) GrantBundle(ctx context.Context, playerID int64, bundle *entity.RewardBundle) (*ApplyResult, error) {
	if bundle == nil {
		return nil, fmt.Errorf("grant bundle: nil bundle")
	}
	req := ApplyRequest{
		PlayerID: playerID,
		Source:   bundle.Source,
	}
	for _, item := range bundle.Items {
		switch item.Type {
		case entity.RewardTypeGold:
			req.GoldDelta += item.Gold
		case entity.RewardTypeEquipment:
			if item.Equipment == nil {
				continue
			}
			eq := item.Equipment
			stats := eq.Stats
			grant := EquipmentGrant{
				ConfigID: eq.ConfigID,
				Rarity:   eq.Rarity,
			}
			slot, ok := entity.ParseEquipmentSlot(eq.Slot)
			if ok {
				grant.Slot = slot
			}
			if stats != nil {
				grant.Attack = stats.Attack
				grant.HP = stats.HP
				grant.BonusAttackPct = stats.BonusAttackPct
			}
			req.EquipmentGrants = append(req.EquipmentGrants, grant)
		}
	}
	return s.Grant(ctx, req)
}
