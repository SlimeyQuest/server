package stage_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	"github.com/slimeyquest/ent/enttest"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/player"
	"github.com/slimeyquest/server/internal/reward"
	"github.com/slimeyquest/server/internal/stage"
)

func TestPushStageGrantsConfiguredBoxReward(t *testing.T) {
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	client := enttest.Open(t, dialect.SQLite, "file:stage_box?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()
	repo := player.NewRepository(client, cfg)
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
	if !res.GetSuccess() {
		t.Fatal("expected first stage clear success")
	}
	boxReward := res.GetBoxReward()
	if boxReward == nil {
		t.Fatal("expected box reward")
	}
	if boxReward.GetBoxCount() < cfg.ClosedLoop.StageBoxMin || boxReward.GetBoxCount() > cfg.ClosedLoop.StageBoxMax {
		t.Fatalf("box reward %d outside configured range [%d,%d]", boxReward.GetBoxCount(), cfg.ClosedLoop.StageBoxMin, cfg.ClosedLoop.StageBoxMax)
	}
	state, err := repo.LoadProgress(ctx, int64(p.ID))
	if err != nil {
		t.Fatal(err)
	}
	if state.BoxCount() != boxReward.GetTotalBoxCount() {
		t.Fatalf("expected stored boxes %d, got %d", boxReward.GetTotalBoxCount(), state.BoxCount())
	}
}
