package player

import (
	"context"
	"fmt"
	"time"

	"github.com/slimeyquest/server/ent"
	"github.com/slimeyquest/server/ent/player"
)

// Repository persists and loads players.
type Repository struct {
	client *ent.Client
}

// NewRepository creates a player repository.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
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

// CreatePlayerForPlatform inserts a new player for a platform account.
func (r *Repository) CreatePlayerForPlatform(ctx context.Context, platform, externalID, nickname string) (*ent.Player, error) {
	now := time.Now()
	p, err := r.client.Player.
		Create().
		SetPlatform(platform).
		SetExternalID(externalID).
		SetNickname(nickname).
		SetLastLoginAt(now).
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
