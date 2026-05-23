package idle_test

import (
	"testing"
	"time"

	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/idle"
	"github.com/slimeyquest/server/internal/services/player"
)

func TestAccumulatedSecondsCap(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	now := start.Add(10 * time.Hour)
	got := idle.AccumulatedSeconds(start, now, 8)
	if got != 8*3600 {
		t.Fatalf("expected cap 8h, got %d", got)
	}
}

func TestComputePreviewGold(t *testing.T) {
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now().UTC().Add(-2 * time.Hour)
	state := &player.ProgressState{
		PlayerID:            1,
		HighestStageCleared: 1,
		CreatedAt:           start,
		LastClaimAt:         &start,
	}
	now := time.Now().UTC()
	preview := idle.ComputePreview(state, cfg, now)
	if preview.GoldTotal <= 0 {
		t.Fatalf("expected positive gold, got %d", preview.GoldTotal)
	}
	if preview.EquipmentRolls > cfg.Globals.MaxEquipRollsPerClaim {
		t.Fatalf("roll cap exceeded: %d", preview.EquipmentRolls)
	}
}
