package player_test

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	equipmentv1 "github.com/slimeyquest/proto/gen/go/equipment"
	"github.com/slimeyquest/server/ent/enttest"
	"github.com/slimeyquest/server/internal/gameplayconfig"
	"github.com/slimeyquest/server/internal/player"
)

func newChestOpenerTestRepo(t *testing.T) (*player.Repository, int64) {
	t.Helper()
	cfg, err := gameplayconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	client := enttest.Open(t, dialect.SQLite, "file:chest_opener?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	repo := player.NewRepository(client, cfg)
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
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("unexpected error: %v", res.GetError())
	}
	if len(res.GetEquipment()) != 2 {
		t.Fatalf("expected 2 equipment items, got %d", len(res.GetEquipment()))
	}
	if res.GetRemainingBoxCount() != 0 {
		t.Fatalf("expected no boxes remaining, got %d", res.GetRemainingBoxCount())
	}
}

func TestChestOpenRejectsInsufficientBoxes(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)

	res, err := svc.OpenChest(context.Background(), playerID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("expected invalid request, got %v", res.GetError().GetCode())
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
		Rarity:   int32(commonv1.EquipmentRarity_EQUIPMENT_RARITY_RARE),
		Slot:     int32(equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_HAT),
		Attack:   10,
		HP:       5,
	})
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	equipRes, err := svc.EquipItem(ctx, playerID, inst.UID, equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_UNSPECIFIED)
	if err != nil {
		t.Fatal(err)
	}
	if equipRes.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("unexpected equip error: %v", equipRes.GetError())
	}

	decomposeEquipped, err := svc.DecomposeEquipment(ctx, playerID, inst.UID)
	if err != nil {
		t.Fatal(err)
	}
	if decomposeEquipped.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST {
		t.Fatalf("expected equipped decompose rejection, got %v", decomposeEquipped.GetError().GetCode())
	}

	state, err = repo.LoadProgress(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	delete(state.Equipment.Equipped, int32(equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_HAT))
	if err := repo.SaveProgress(ctx, state); err != nil {
		t.Fatal(err)
	}

	decomposeRes, err := svc.DecomposeEquipment(ctx, playerID, inst.UID)
	if err != nil {
		t.Fatal(err)
	}
	if decomposeRes.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("unexpected decompose error: %v", decomposeRes.GetError())
	}
	if decomposeRes.GetGainedGold() <= 0 || decomposeRes.GetTotalGold() <= 0 {
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
	if res.GetError().GetCode() != commonv1.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("unexpected upgrade error: %v", res.GetError())
	}
	if res.GetChestLevel() != 2 {
		t.Fatalf("expected chest level 2, got %d", res.GetChestLevel())
	}
}

func TestDrawSkillAndCompanion(t *testing.T) {
	repo, playerID := newChestOpenerTestRepo(t)
	svc := player.NewChestOpenerService(repo)

	skillRes, err := svc.DrawSkill(context.Background(), playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(skillRes.GetRewards()) != 2 || skillRes.GetShopLevel() < 1 {
		t.Fatalf("unexpected skill draw result: %+v", skillRes)
	}

	companionRes, err := svc.DrawCompanion(context.Background(), playerID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(companionRes.GetRewards()) != 2 || companionRes.GetShopLevel() < 1 {
		t.Fatalf("unexpected companion draw result: %+v", companionRes)
	}
}
