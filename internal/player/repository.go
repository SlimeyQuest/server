package player

import (
	"context"
	"fmt"
	"time"

	"github.com/slimeyquest/server/ent"
	"github.com/slimeyquest/server/ent/player"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

// Repository persists and loads players.
type Repository struct {
	client *ent.Client
	cfg    *gameplayconfig.Config
}

// NewRepository creates a player repository.
func NewRepository(client *ent.Client, cfg *gameplayconfig.Config) *Repository {
	return &Repository{client: client, cfg: cfg}
}

// Cfg returns gameplay config used by the repository.
func (r *Repository) Cfg() *gameplayconfig.Config {
	return r.cfg
}

// FindByExternalID returns a player for the given platform account identity.
func (r *Repository) FindByExternalID(ctx context.Context, platform, externalID string) (*ent.Player, error) {
	p, err := r.client.Player.
		Query().
		Where(
			player.PlatformEQ(platform),
			player.ExternalIDEQ(externalID),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("find player by external id: %w", err)
	}
	return p, nil
}

// GetByID loads a player by primary key.
func (r *Repository) GetByID(ctx context.Context, playerID int64) (*ent.Player, error) {
	p, err := r.client.Player.Get(ctx, int(playerID))
	if err != nil {
		return nil, fmt.Errorf("get player by id: %w", err)
	}
	return p, nil
}

// LoadProgress loads gameplay state for a player id.
func (r *Repository) LoadProgress(ctx context.Context, playerID int64) (*ProgressState, error) {
	p, err := r.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return FromEntity(p), nil
}

// CreatePlayerForPlatform inserts a new player for a platform account.
func (r *Repository) CreatePlayerForPlatform(ctx context.Context, platform, externalID, nickname string) (*ent.Player, error) {
	now := time.Now()
	equip := StarterEquipment(r.cfg)
	p, err := r.client.Player.
		Create().
		SetPlatform(platform).
		SetExternalID(externalID).
		SetNickname(nickname).
		SetLastLoginAt(now).
		SetLastClaimAt(now).
		SetAdventureID(1).
		SetStageIndex(1).
		SetHighestStageCleared(0).
		SetEquipmentJSON(EncodeEquipment(equip)).
		SetClearedMilestones([]int32{}).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create player for platform: %w", err)
	}
	return p, nil
}

// RecordLogin updates last_login_at for a successful login.
func (r *Repository) RecordLogin(ctx context.Context, playerID int) (*ent.Player, error) {
	now := time.Now()
	p, err := r.client.Player.
		UpdateOneID(playerID).
		SetLastLoginAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("record player login: %w", err)
	}
	return p, nil
}

// SaveProgress persists gameplay fields from ProgressState.
func (r *Repository) SaveProgress(ctx context.Context, state *ProgressState) error {
	if state == nil {
		return fmt.Errorf("save progress: nil state")
	}
	upd := r.client.Player.UpdateOneID(int(state.PlayerID)).
		SetGold(state.Gold).
		SetGems(state.Gems).
		SetAdventureID(state.AdventureID).
		SetStageIndex(state.StageIndex).
		SetHighestStageCleared(state.HighestStageCleared).
		SetEquipmentJSON(EncodeEquipment(state.Equipment)).
		SetClearedMilestones(state.ClearedMilestones)
	if state.LastClaimAt != nil {
		upd = upd.SetLastClaimAt(*state.LastClaimAt)
	}
	if _, err := upd.Save(ctx); err != nil {
		return fmt.Errorf("save player progress: %w", err)
	}
	return nil
}
