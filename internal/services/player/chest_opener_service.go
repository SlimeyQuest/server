package player

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/slimeyquest/server/internal/apitypes"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

const (
	TestZoneID             int32 = 1
	DefaultChestLevel      int32 = 1
	SkillShopLevelBase     int32 = 1
	CompanionShopLevelBase int32 = 1

	maxChestLevel int32 = 28
	maxShopLevel  int32 = 20
	maxChestTiers int32 = 4
	maxShopTiers  int32 = 5
)

// ChestOpenerService implements the MVP chest/equipment/skill/companion loop.
type ChestOpenerService struct {
	players Repository
}

// NewChestOpenerService creates a gameplay loop service.
func NewChestOpenerService(players Repository) *ChestOpenerService {
	return &ChestOpenerService{players: players}
}

// ClosedLoopService is kept as a compatibility alias for older wiring code.
type ClosedLoopService = ChestOpenerService

// NewClosedLoopService creates a gameplay loop service.
func NewClosedLoopService(players Repository) *ClosedLoopService {
	return NewChestOpenerService(players)
}

// CreateRole updates the display name and returns the current role profile.
func (s *ChestOpenerService) CreateRole(ctx context.Context, playerID int64, displayName string) (*apitypes.CreateRoleRes, error) {
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
	return &apitypes.CreateRoleRes{Profile: ToProfile(state, s.players.Cfg())}, nil
}

// OpenChest grants one equipment item from a chest-level-scaled drop table.
func (s *ClosedLoopService) OpenChest(ctx context.Context, playerID int64, count int32) (*apitypes.ChestOpenRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		count = 1
	}
	if state.BoxCount() < count {
		return &apitypes.ChestOpenRes{Error: apitypes.Err(apitypes.ErrorCodeInvalidRequest, "not enough boxes"), RemainingBoxCount: state.BoxCount()}, nil
	}
	equipment := make([]apitypes.EquipmentInfo, 0, count)
	for i := int32(0); i < count; i++ {
		seed := playerID + int64(i) + time.Now().UnixNano()
		row := s.pickChestDrop(state, seed)
		inst := state.Equipment.AddInstance(row)
		equipment = append(equipment, inst.ToAPI())
	}
	state.SetBoxCount(state.BoxCount() - count)
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &apitypes.ChestOpenRes{Equipment: equipment, RemainingBoxCount: state.BoxCount()}, nil
}

// EquipItem equips an owned item into its compatible slot.
func (s *ClosedLoopService) EquipItem(ctx context.Context, playerID int64, equipmentUID int64, requestedSlot int32) (*apitypes.EquipItemRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	inst, ok := state.Equipment.Instances[equipmentUID]
	if !ok {
		return &apitypes.EquipItemRes{Error: apitypes.Err(apitypes.ErrorCodeNotFound, "equipment not found")}, nil
	}
	slot := inst.Slot
	if requestedSlot != 0 && requestedSlot != slot {
		return &apitypes.EquipItemRes{Error: apitypes.Err(apitypes.ErrorCodeInvalidRequest, "equipment slot mismatch")}, nil
	}
	if state.Equipment.Equipped == nil {
		state.Equipment.Equipped = make(map[int32]int64)
	}
	state.Equipment.Equipped[slot] = equipmentUID
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &apitypes.EquipItemRes{
		EquippedSlots: state.Equipment.EquippedSlotsAPI(),
		CombatPower:   ComputeCombatPower(state, s.players.Cfg()),
	}, nil
}

// DecomposeEquipment removes one unequipped equipment and grants gold.
func (s *ClosedLoopService) DecomposeEquipment(ctx context.Context, playerID int64, equipmentUID int64) (*apitypes.DecomposeEquipmentRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	inst, ok := state.Equipment.Instances[equipmentUID]
	if !ok {
		return &apitypes.DecomposeEquipmentRes{Error: apitypes.Err(apitypes.ErrorCodeNotFound, "equipment not found")}, nil
	}
	for _, equippedUID := range state.Equipment.Equipped {
		if equippedUID == equipmentUID {
			return &apitypes.DecomposeEquipmentRes{Error: apitypes.Err(apitypes.ErrorCodeInvalidRequest, "equipped item cannot be decomposed")}, nil
		}
	}
	delete(state.Equipment.Instances, equipmentUID)
	gained := s.decomposeGold(inst)
	state.Gold += gained
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &apitypes.DecomposeEquipmentRes{GainedGold: gained, TotalGold: state.Gold}, nil
}

// UpgradeChest spends gold to raise chest level.
func (s *ClosedLoopService) UpgradeChest(ctx context.Context, playerID int64, targetLevel int32) (*apitypes.UpgradeChestRes, error) {
	state, err := s.players.LoadProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	current := state.ChestLevel()
	if current <= 0 {
		current = DefaultChestLevel
	}
	if targetLevel > s.players.Cfg().ClosedLoop.OpenerMaxLevelValue() {
		targetLevel = s.players.Cfg().ClosedLoop.OpenerMaxLevelValue()
	}
	if targetLevel <= current {
		return &apitypes.UpgradeChestRes{ChestLevel: current, TotalGold: state.Gold}, nil
	}
	cost := s.chestUpgradeCost(current, targetLevel)
	if state.Gold < cost {
		return &apitypes.UpgradeChestRes{Error: apitypes.Err(apitypes.ErrorCodeInvalidRequest, "not enough gold"), ChestLevel: current, TotalGold: state.Gold}, nil
	}
	state.Gold -= cost
	state.SetChestLevel(targetLevel)
	if err := s.players.SaveProgress(ctx, state); err != nil {
		return nil, err
	}
	return &apitypes.UpgradeChestRes{ChestLevel: targetLevel, TotalGold: state.Gold}, nil
}

// DrawSkill returns MVP test skills and a shop level derived from draw count.
func (s *ClosedLoopService) DrawSkill(_ context.Context, playerID int64, drawCount int32) (*apitypes.DrawSkillRes, error) {
	if drawCount <= 0 {
		drawCount = 1
	}
	cfg := s.players.Cfg().ClosedLoop
	shopLevel := min32(SkillShopLevelBase+drawCount/s.shopDrawsPerLevel(true), maxShopLevel)
	rewards := make([]apitypes.SkillInfo, 0, drawCount)
	for i := int32(0); i < drawCount; i++ {
		rewards = append(rewards, testSkill(playerID, i, shopLevel, cfg.Shop))
	}
	return &apitypes.DrawSkillRes{Rewards: rewards, ShopLevel: shopLevel}, nil
}

// DrawCompanion returns MVP test companions and a shop level derived from draw count.
func (s *ClosedLoopService) DrawCompanion(_ context.Context, playerID int64, drawCount int32) (*apitypes.DrawCompanionRes, error) {
	if drawCount <= 0 {
		drawCount = 1
	}
	cfg := s.players.Cfg().ClosedLoop
	shopLevel := min32(CompanionShopLevelBase+drawCount/s.shopDrawsPerLevel(false), maxShopLevel)
	rewards := make([]apitypes.CompanionInfo, 0, drawCount)
	for i := int32(0); i < drawCount; i++ {
		rewards = append(rewards, testCompanion(playerID, i, shopLevel, cfg.Shop))
	}
	return &apitypes.DrawCompanionRes{Rewards: rewards, ShopLevel: shopLevel}, nil
}

func (s *ClosedLoopService) pickChestDrop(state *ProgressState, seed int64) gameplayconfig.DropRow {
	cfg := s.players.Cfg().ClosedLoop
	level := min32(state.ChestLevel(), s.players.Cfg().ClosedLoop.OpenerMaxLevelValue())
	row := s.players.Cfg().PickIdleDrop(seed)
	row.Rarity = rarityForLevel(level, chestCurve(cfg.Chest), stablePick(seed, level))
	row.Attack += int64(level-1) * cfg.Stage.AttackPerLevel
	row.HP += int64(level-1) * cfg.Stage.HPPerLevel
	row.Slot = normalizeDropSlot(row.Slot, seed)
	return row
}

func normalizeDropSlot(slot int32, seed int64) int32 {
	for _, s := range apitypes.AllEquipmentSlots {
		if slot == s {
			return slot
		}
	}
	idx := int(seed % int64(len(apitypes.AllEquipmentSlots)))
	if idx < 0 {
		idx = -idx
	}
	return apitypes.AllEquipmentSlots[idx]
}

func (s *ClosedLoopService) decomposeGold(inst EquipmentInstance) int64 {
	cfg := s.players.Cfg().ClosedLoop
	baseGold := cfg.Economy.DecomposeBaseGold
	if baseGold <= 0 {
		baseGold = 20
	}
	levelGold := cfg.Economy.DecomposeLevelGold
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
	base := cfg.OpenerUpgradeBaseGoldValue()
	if base <= 0 {
		base = 100
	}
	growthPct := cfg.OpenerUpgradeGrowthPctValue()
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

func testSkill(playerID int64, index int32, shopLevel int32, curve gameplayconfig.ClosedLoopShopConfig) apitypes.SkillInfo {
	quality := rarityForLevel(shopLevel, shopCurve(curve), stablePick(playerID, index))
	return apitypes.SkillInfo{SkillID: 1001, Name: fmt.Sprintf("Warrior Slash %d", stablePick(playerID, index)+1), Quality: quality}
}

func testCompanion(playerID int64, index int32, shopLevel int32, curve gameplayconfig.ClosedLoopShopConfig) apitypes.CompanionInfo {
	quality := rarityForLevel(shopLevel, shopCurve(curve), stablePick(playerID, index))
	return apitypes.CompanionInfo{CompanionID: 2001, Name: fmt.Sprintf("Tiny Slime %d", stablePick(playerID, index)+1), Quality: quality}
}

func stablePick(parts ...any) uint32 {
	h := fnv.New32a()
	_, _ = fmt.Fprint(h, parts...)
	return h.Sum32()
}

func rarityForLevel(level int32, curve probabilityCurve, seed uint32) int32 {
	weights, rarities := smoothRarityWeights(level, curve)
	total := int32(0)
	for _, w := range weights {
		total += w
	}
	if total <= 0 || len(rarities) == 0 {
		return 1
	}
	roll := int32(seed%uint32(total)) + 1
	cursor := int32(0)
	for i, rarity := range rarities {
		cursor += weights[i]
		if roll <= cursor {
			return rarity
		}
	}
	return rarities[len(rarities)-1]
}

func smoothRarityWeights(level int32, curve probabilityCurve) ([]int32, []int32) {
	if level <= 0 {
		level = 1
	}
	rarities := curve.rarities
	if len(rarities) == 0 {
		rarities = []int32{1}
	}
	initial := clamp32(curve.initialRarities, 1, int32(len(rarities)))
	maxActive := clamp32(curve.maxActiveRarities, 1, int32(len(rarities)))
	unlockInterval := curve.unlockInterval
	if unlockInterval <= 0 {
		unlockInterval = 4
	}
	unlocked := initial + (level-1)/unlockInterval
	unlocked = clamp32(unlocked, initial, int32(len(rarities)))
	start := int32(0)
	if unlocked > maxActive {
		start = unlocked - maxActive
	}
	activeRarities := append([]int32(nil), rarities[start:unlocked]...)
	base := geometricWeights(len(activeRarities), curve.lowestBaseWeight, curve.adjacentGapPct)
	step := (level - 1) % unlockInterval
	shift := step * clamp32(curve.progressShiftPct, 0, 10)
	weights := applySmoothShift(base, shift, curve.topWeightCapPct)
	return weights, activeRarities
}

type probabilityCurve struct {
	rarities          []int32
	initialRarities   int32
	maxActiveRarities int32
	unlockInterval    int32
	lowestBaseWeight  int32
	adjacentGapPct    int32
	progressShiftPct  int32
	topWeightCapPct   int32
}

func chestCurve(c gameplayconfig.ClosedLoopChestConfig) probabilityCurve {
	return probabilityCurve{c.Rarities, c.InitialRarities, c.MaxActiveRarities, c.UnlockInterval, c.LowestBaseWeight, c.AdjacentGapPct, c.ProgressShiftPct, c.TopWeightCapPct}
}

func shopCurve(c gameplayconfig.ClosedLoopShopConfig) probabilityCurve {
	return probabilityCurve{c.Rarities, c.InitialRarities, c.MaxActiveRarities, c.UnlockInterval, c.LowestBaseWeight, c.AdjacentGapPct, c.ProgressShiftPct, c.TopWeightCapPct}
}

func geometricWeights(count int, lowestBaseWeight int32, adjacentGapPct int32) []int32 {
	if count <= 0 {
		return nil
	}
	if lowestBaseWeight <= 0 {
		lowestBaseWeight = 8
	}
	gap := clamp32(adjacentGapPct, 40, 80)
	weights := make([]int32, count)
	weights[count-1] = lowestBaseWeight
	for i := count - 2; i >= 0; i-- {
		weights[i] = weights[i+1] * (100 + gap) / 100
		if weights[i] <= weights[i+1] {
			weights[i] = weights[i+1] + 1
		}
	}
	return normalizeTo100(weights)
}

func applySmoothShift(base []int32, shift int32, topCap int32) []int32 {
	weights := append([]int32(nil), base...)
	if len(weights) < 2 || shift <= 0 {
		return capTopWeight(normalizeTo100(weights), topCap)
	}
	highStart := len(weights) / 2
	lowCount := highStart
	highCount := len(weights) - highStart
	for i := 0; i < highStart; i++ {
		weights[i] -= shift / int32(lowCount)
	}
	for i := highStart; i < len(weights); i++ {
		weights[i] += shift / int32(highCount)
	}
	return capTopWeight(normalizeTo100(weights), topCap)
}

func normalizeTo100(weights []int32) []int32 {
	if len(weights) == 0 {
		return weights
	}
	total := int32(0)
	for i, w := range weights {
		if w < 0 {
			weights[i] = 0
		}
		total += weights[i]
	}
	if total <= 0 {
		weights[0] = 100
		return weights
	}
	remaining := int32(100)
	for i := range weights {
		if i == len(weights)-1 {
			weights[i] = remaining
			break
		}
		weights[i] = weights[i] * 100 / total
		remaining -= weights[i]
	}
	return weights
}

func capTopWeight(weights []int32, topCap int32) []int32 {
	if len(weights) == 0 || topCap <= 0 || weights[len(weights)-1] <= topCap {
		return weights
	}
	excess := weights[len(weights)-1] - topCap
	weights[len(weights)-1] = topCap
	weights[0] += excess
	return normalizeTo100(weights)
}

func clamp32(v, min, max int32) int32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (s *ClosedLoopService) shopDrawsPerLevel(skill bool) int32 {
	cfg := s.players.Cfg().ClosedLoop
	if skill {
		if cfg.Economy.SkillDrawsPerLevel > 0 {
			return cfg.Economy.SkillDrawsPerLevel
		}
		return 10
	}
	if cfg.Economy.CompanionDrawsPerLevel > 0 {
		return cfg.Economy.CompanionDrawsPerLevel
	}
	return 10
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

