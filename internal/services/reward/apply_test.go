package reward_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/player"
	"github.com/slimeyquest/server/internal/services/reward"
)

func TestApplyGold(t *testing.T) {
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	applier := reward.NewApplier(log, cfg)
	state := &player.ProgressState{PlayerID: 1, Gold: 10}
	result, err := applier.Apply(context.Background(), state, reward.ApplyRequest{
		PlayerID:  1,
		Source:    apitypes.RewardSourceIdleClaim,
		GoldDelta: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Gold != 35 {
		t.Fatalf("expected gold 35, got %d", result.State.Gold)
	}
	if len(result.AppliedBundle.Items) != 1 {
		t.Fatalf("expected one reward item, got %d", len(result.AppliedBundle.Items))
	}
}
