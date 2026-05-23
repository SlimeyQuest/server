package player

import (
	"context"

	"github.com/slimeyquest/ent"
	"github.com/slimeyquest/server/internal/config"
)

// Repository defines the player persistence operations required by gameplay services.
type Repository interface {
	Cfg() *config.GameplayConfig
	FindByExternalID(ctx context.Context, platform, externalID string) (*ent.Player, error)
	GetByID(ctx context.Context, playerID int64) (*ent.Player, error)
	LoadProgress(ctx context.Context, playerID int64) (*ProgressState, error)
	CreatePlayerForPlatform(ctx context.Context, platform, externalID, nickname string) (*ent.Player, error)
	RecordLogin(ctx context.Context, playerID int) (*ent.Player, error)
	SaveRole(ctx context.Context, state *ProgressState) error
	SaveProgress(ctx context.Context, state *ProgressState) error
}
