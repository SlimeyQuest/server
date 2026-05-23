package stage_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	"github.com/slimeyquest/ent/enttest"
	"github.com/slimeyquest/server/internal/data/playerrepo"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/reward"
	"github.com/slimeyquest/server/internal/services/stage"
)

func TestPushStageGrantsConfiguredBoxReward(t *testing.T) {
	cfg, err := config.LoadGameplay()
	if err != nil {
		t.Fatal(err)
	}
	client := enttest.Open(t, dialect.SQLite, "file:stage_box?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()
	repo := playerrepo.New(client, cfg)
	p, err := repo.CreatePlayerForPlatform(ctx, "test", "stage-box", "Tester")
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rewardSvc := reward.NewService(log, cfg, repo)
	stageSvc := stage.NewService(log, cfg, repo, rewardSvc)

	res, err := stageSvc.PushStage(ctx, int64(p.ID), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatal("expected first stage clear success")
	}
	boxReward := res.BoxReward
	if boxReward == nil {
		t.Fatal("expected box reward")
	}
	if boxReward.BoxCount < cfg.ClosedLoop.StageBoxMin || boxReward.BoxCount > cfg.ClosedLoop.StageBoxMax {
		t.Fatalf("box reward %d outside configured range [%d,%d]", boxReward.BoxCount, cfg.ClosedLoop.StageBoxMin, cfg.ClosedLoop.StageBoxMax)
	}
	state, err := repo.LoadProgress(ctx, int64(p.ID))
	if err != nil {
		t.Fatal(err)
	}
	if state.BoxCount() != boxReward.TotalBoxCount {
		t.Fatalf("expected stored boxes %d, got %d", boxReward.TotalBoxCount, state.BoxCount())
	}
}
