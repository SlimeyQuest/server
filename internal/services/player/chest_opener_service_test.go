package player_test

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	"github.com/slimeyquest/ent/enttest"
	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/data/playerrepo"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/services/player"
)

func newChestOpenerTestRepo(t *testing.T) (player.Repository, int64) {
	t.Helper()
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	client := enttest.Open(t, dialect.SQLite, "file:chest_opener?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	repo := playerrepo.New(client, cfg)
	p, err := repo.CreatePlayerForPlatform(context.Background(), "test", t.Name(), "Tester")
	if err != nil {
		t.Fatal(err)
	}
	return repo, int64(p.ID)
}

func TestChestOpenConsumesStoredBoxesAndCreatesEquipment(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)
	ctx := context.Background()

	state, err := repo.LoadProgress(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	state.SetBoxCount(2)
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	res, err := svc.OpenChest(ctx, playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if apitypes.HasError(res.Error) {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if len(res.Equipment) != 2 {
		t.Fatalf("expected 2 equipment items, got %d", len(res.Equipment))
	}
	if res.RemainingBoxCount != 0 {
		t.Fatalf("expected no boxes remaining, got %d", res.RemainingBoxCount)
	}
}

func TestChestOpenRejectsInsufficientBoxes(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)

	res, err := svc.OpenChest(context.Background(), playerID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == nil || res.Error.Code != apitypes.ErrorCodeInvalidRequest {
		t.Fatalf("expected invalid request, got %v", res.Error)
	}
}

func TestEquipItemAndDecomposeLifecycle(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)
	ctx := context.Background()

	state, err := repo.LoadProgress(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	inst := state.Equipment.AddInstance(gameplayconfig.DropRow{
		ConfigID: 9001,
		Rarity:   4,
		Slot:     apitypes.SlotHat,
		Attack:   10,
		HP:       5,
	})
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	equipRes, err := svc.EquipItem(ctx, playerID, inst.UID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if apitypes.HasError(equipRes.Error) {
		t.Fatalf("unexpected equip error: %v", equipRes.Error)
	}

	decomposeEquipped, err := svc.DecomposeEquipment(ctx, playerID, inst.UID)
	if err != nil {
		t.Fatal(err)
	}
	if decomposeEquipped.Error == nil || decomposeEquipped.Error.Code != apitypes.ErrorCodeInvalidRequest {
		t.Fatalf("expected equipped decompose rejection, got %v", decomposeEquipped.Error)
	}

	state, err = repo.LoadProgress(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	delete(state.Equipment.Equipped, apitypes.SlotHat)
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	decomposeRes, err := svc.DecomposeEquipment(ctx, playerID, inst.UID)
	if err != nil {
		t.Fatal(err)
	}
	if apitypes.HasError(decomposeRes.Error) {
		t.Fatalf("unexpected decompose error: %v", decomposeRes.Error)
	}
	if decomposeRes.GainedGold <= 0 || decomposeRes.TotalGold <= 0 {
		t.Fatal("expected gained and total gold")
	}
}

func TestUpgradeChestUsesConfiguredGoldCost(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)
	ctx := context.Background()

	state, err := repo.LoadProgress(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	state.Gold = repo.Cfg().ClosedLoop.OpenerUpgradeBaseGold
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	res, err := svc.UpgradeChest(ctx, playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if apitypes.HasError(res.Error) {
		t.Fatalf("unexpected upgrade error: %v", res.Error)
	}
	if res.ChestLevel != 2 {
		t.Fatalf("expected chest level 2, got %d", res.ChestLevel)
	}
}

func TestDrawSkillAndCompanion(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)

	skillRes, err := svc.DrawSkill(context.Background(), playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(skillRes.Rewards) != 2 || skillRes.ShopLevel < 1 {
		t.Fatalf("unexpected skill draw result: %+v", skillRes)
	}

	companionRes, err := svc.DrawCompanion(context.Background(), playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(companionRes.Rewards) != 2 || companionRes.ShopLevel < 1 {
		t.Fatalf("unexpected companion draw result: %+v", companionRes)
	}
}
