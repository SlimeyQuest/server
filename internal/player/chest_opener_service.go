package player

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	equipmentv1 "github.com/slimeyquest/proto/gen/go/equipment"
	playerv1 "github.com/slimeyquest/proto/gen/go/player"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

const (
	TestZoneID              int32 = 1
	DefaultChestLevel       int32 = 1
	SkillShopLevelBase      int32 = 1
	CompanionShopLevelBase  int32 = 1
)

// ChestOpenerService implements the MVP chest/equipment/skill/companion loop.
type ChestOpenerService struct {
	players *Repository
}

// NewChestOpenerService creates a gameplay loop service.
func NewChestOpenerService(players *Repository) *ChestOpenerService {
	return &ChestOpenerService{players: players}
}

// ClosedLoopService is kept as a compatibility alias for older wiring code.
type ClosedLoopService = ChestOpenerService

// NewClosedLoopService creates a gameplay loop service.
func NewClosedLoopService(players *Repository) *ClosedLoopService {
	return NewChestOpenerService(players)
}

// CreateRole updates the display name and returns the current role profile.
func (s *ChestOpenerService) CreateRole(ctx context.Context, playerID int64, displayName string) (*playerv1.CreateRoleRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if displayName != "" {
		state.DisplayName = displayName
	}
	if err := s.players.SaveRole(ctx, state); err != nil {
		return nil, err
	}
	return &playerv1.CreateRoleRes{Profile: ToProfile(state, s.players.Cfg())}, nil
}

// OpenChest grants one equipment item from a chest-level-scaled drop table.
func (s *ClosedLoopService) OpenChest(ctx context.Context, playerID int64, count int32) (*equipmentv1.ChestOpenRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		count = 1
	}
	if state.BoxCount() < count {
		return &equipmentv1.ChestOpenRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "not enough boxes"), RemainingBoxCount: state.BoxCount()}, nil
	}
	equipment := make([]*equipmentv1.EquipmentInfo, 0, count)
	for i := int32(0); i < count; i++ {
		seed := playerID + int64(i) + time.Now().UnixNano()
		row := s.pickChestDrop(state, seed)
		inst := state.Equipment.AddInstance(row)
		equipment = append(equipment, inst.ToProto())
	}
	state.SetBoxCount(state.BoxCount() - count)
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &equipmentv1.ChestOpenRes{Equipment: equipment, RemainingBoxCount: state.BoxCount()}, nil
}

// EquipItem equips an owned item into its compatible slot.
func (s *ClosedLoopService) EquipItem(ctx context.Context, playerID int64, equipmentUID int64, requestedSlot equipmentv1.EquipmentSlot) (*equipmentv1.EquipItemRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	inst, ok := state.Equipment.Instances[equipmentUID]
	if !ok {
		return &equipmentv1.EquipItemRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_NOT_FOUND, "equipment not found")}, nil
	}
	slot := equipmentv1.EquipmentSlot(inst.Slot)
	if requestedSlot != equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_UNSPECIFIED && requestedSlot != slot {
		return &equipmentv1.EquipItemRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "equipment slot mismatch")}, nil
	}
	if state.Equipment.Equipped == nil {
		state.Equipment.Equipped = make(map[int32]int64)
	}
	state.Equipment.Equipped[int32(slot)] = equipmentUID
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &equipmentv1.EquipItemRes{
		EquippedSlots: state.Equipment.EquippedSlots(),
		CombatPower:   ComputeCombatPower(state, s.players.Cfg()),
	}, nil
}

// DecomposeEquipment removes one unequipped equipment and grants gold.
func (s *ClosedLoopService) DecomposeEquipment(ctx context.Context, playerID int64, equipmentUID int64) (*equipmentv1.DecomposeEquipmentRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	inst, ok := state.Equipment.Instances[equipmentUID]
	if !ok {
		return &equipmentv1.DecomposeEquipmentRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_NOT_FOUND, "equipment not found")}, nil
	}
	for _, equippedUID := range state.Equipment.Equipped {
		if equippedUID == equipmentUID {
			return &equipmentv1.DecomposeEquipmentRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "equipped item cannot be decomposed")}, nil
		}
	}
	delete(state.Equipment.Instances, equipmentUID)
	gained := s.decomposeGold(inst)
	state.Gold += gained
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &equipmentv1.DecomposeEquipmentRes{GainedGold: gained, TotalGold: state.Gold}, nil
}

// UpgradeChest spends gold to raise chest level.
func (s *ClosedLoopService) UpgradeChest(ctx context.Context, playerID int64, targetLevel int32) (*equipmentv1.UpgradeChestRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	current := state.ChestLevel()
	if targetLevel <= current {
		return &equipmentv1.UpgradeChestRes{ChestLevel: current, TotalGold: state.Gold}, nil
	}
	cost := s.chestUpgradeCost(current, targetLevel)
	if state.Gold < cost {
		return &equipmentv1.UpgradeChestRes{Error: loopError(commonv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "not enough gold"), ChestLevel: current, TotalGold: state.Gold}, nil
	}
	state.Gold -= cost
	state.SetChestLevel(targetLevel)
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &equipmentv1.UpgradeChestRes{ChestLevel: targetLevel, TotalGold: state.Gold}, nil
}

// DrawSkill returns MVP test skills and a shop level derived from draw count.
func (s *ClosedLoopService) DrawSkill(_ context.Context, playerID int64, drawCount int32) (*playerv1.DrawSkillRes, error) {
	if drawCount <= 0 {
		drawCount = 1
	}
	shopLevel := SkillShopLevelBase + drawCount/10
	rewards := make([]*playerv1.SkillInfo, 0, drawCount)
	for i := int32(0); i < drawCount; i++ {
		rewards = append(rewards, testSkill(playerID, i, shopLevel))
	}
	return &playerv1.DrawSkillRes{Rewards: rewards, ShopLevel: shopLevel}, nil
}

// DrawCompanion returns MVP test companions and a shop level derived from draw count.
func (s *ClosedLoopService) DrawCompanion(_ context.Context, playerID int64, drawCount int32) (*playerv1.DrawCompanionRes, error) {
	if drawCount <= 0 {
		drawCount = 1
	}
	shopLevel := CompanionShopLevelBase + drawCount/10
	rewards := make([]*playerv1.CompanionInfo, 0, drawCount)
	for i := int32(0); i < drawCount; i++ {
		rewards = append(rewards, testCompanion(playerID, i, shopLevel))
	}
	return &playerv1.DrawCompanionRes{Rewards: rewards, ShopLevel: shopLevel}, nil
}

func (s *ClosedLoopService) pickChestDrop(state *ProgressState, seed int64) gameplayconfig.DropRow {
	cfg := s.players.Cfg().ClosedLoop
	row := s.players.Cfg().PickIdleDrop(seed)
	level := state.ChestLevel()
	boostEvery := cfg.RarityBoostEveryLevels
	if boostEvery <= 0 {
		boostEvery = 5
	}
	if level > 1 && row.Rarity < int32(commonv1.EquipmentRarity_EQUIPMENT_RARITY_LEGENDARY) {
		row.Rarity += (level - 1) / boostEvery
		if row.Rarity > int32(commonv1.EquipmentRarity_EQUIPMENT_RARITY_LEGENDARY) {
			row.Rarity = int32(commonv1.EquipmentRarity_EQUIPMENT_RARITY_LEGENDARY)
		}
	}
	row.Attack += int64(level-1) * cfg.EquipmentAttackPerLevel
	row.HP += int64(level-1) * cfg.EquipmentHPPerLevel
	row.Slot = normalizeDropSlot(row.Slot, seed)
	return row
}

func normalizeDropSlot(slot int32, seed int64) int32 {
	slots := []equipmentv1.EquipmentSlot{
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_HAT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_SHOE_LEFT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_SHOE_RIGHT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_GLOVE_LEFT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_GLOVE_RIGHT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_CLOTH,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_PANTS,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_WEAPON,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_RING_LEFT,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_RING_RIGHT,
	}
	for _, s := range slots {
		if slot == int32(s) {
			return slot
		}
	}
	idx := int(seed % int64(len(slots)))
	if idx < 0 {
		idx = -idx
	}
	return int32(slots[idx])
}

func (s *ClosedLoopService) decomposeGold(inst EquipmentInstance) int64 {
	cfg := s.players.Cfg().ClosedLoop
	baseGold := cfg.DecomposeBaseGold
	if baseGold <= 0 {
		baseGold = 20
	}
	levelGold := cfg.DecomposeLevelGold
	if levelGold <= 0 {
		levelGold = 5
	}
	base := baseGold + int64(inst.Level)*levelGold
	if inst.Rarity > 0 {
		base *= int64(inst.Rarity)
	}
	return base
}

func (s *ClosedLoopService) chestUpgradeCost(current, target int32) int64 {
	cfg := s.players.Cfg().ClosedLoop
	base := cfg.OpenerUpgradeBaseGold
	if base <= 0 {
		base = 100
	}
	growthPct := cfg.OpenerUpgradeGrowthPct
	var total int64
	cost := base
	for lv := int32(2); lv <= target; lv++ {
		if lv > current {
			total += cost
		}
		cost += cost * int64(growthPct) / 100
	}
	return total
}

func testSkill(playerID int64, index int32, shopLevel int32) *playerv1.SkillInfo {
	quality := int32(1)
	if shopLevel >= 3 {
		quality = 2
	}
	return &playerv1.SkillInfo{SkillId: 1001, Name: fmt.Sprintf("Warrior Slash %d", stablePick(playerID, index)+1), Quality: quality}
}

func testCompanion(playerID int64, index int32, shopLevel int32) *playerv1.CompanionInfo {
	quality := int32(1)
	if shopLevel >= 3 {
		quality = 2
	}
	return &playerv1.CompanionInfo{CompanionId: 2001, Name: fmt.Sprintf("Tiny Slime %d", stablePick(playerID, index)+1), Quality: quality}
}

func stablePick(playerID int64, index int32) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d", playerID, index)))
	return h.Sum32() % 3
}

func loopError(code commonv1.ErrorCode, message string) *commonv1.ErrorInfo {
	return &commonv1.ErrorInfo{Code: code, Message: message}
}
