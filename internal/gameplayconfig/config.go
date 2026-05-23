package gameplayconfig

import (
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed data/idle/idle_globals.json
//go:embed data/idle/idle_gold_per_sec.csv
//go:embed data/stages/stages.csv
//go:embed data/economy/drop_table_idle.csv
//go:embed data/economy/chest_opener_config.yaml
var embeddedFS embed.FS

// Config holds gameplay balance tables loaded at startup.
type Config struct {
	Globals           Globals
	ClosedLoop        ClosedLoopConfig
	GoldPerSecByStage map[int32]float64
	Stages            map[int32]StageRow
	IdleDrops         []DropRow
	StarterWeapon     DropRow
}

// Globals are idle and combat tuning parameters.
type Globals struct {
	OfflineCapHours       float64 `json:"offline_cap_hours"`
	BaseOfflineRate       float64 `json:"base_offline_rate"`
	EquipRollIntervalSec  int64   `json:"equip_roll_interval_sec"`
	MaxEquipRollsPerClaim int64   `json:"max_equip_rolls_per_claim"`
	OnlineMultiplier      float64 `json:"online_multiplier"`
	ClearThreshold        float64 `json:"clear_threshold"`
	KArmor                float64 `json:"k_armor"`
	IdleScalePerStage     float64 `json:"idle_scale_per_stage"`
}

// ClosedLoopConfig contains fixed values for boxes, opener upgrades and MVP draws.
type ClosedLoopConfig struct {
	StageBoxMin                int32   `yaml:"-"`
	StageBoxMax                int32   `yaml:"-"`
	OpenerUpgradeBaseGold      int64   `yaml:"-"`
	OpenerUpgradeGrowthPct     int32   `yaml:"-"`
	OpenerMaxLevel             int32   `yaml:"-"`
	EquipmentAttackPerLevel    int64   `yaml:"-"`
	EquipmentHPPerLevel        int64   `yaml:"-"`
	RarityBoostEveryLevels     int32   `yaml:"-"`
	ChestRarityWeights         []int32 `yaml:"-"`
	ChestRarityUnlockLevels    []int32 `yaml:"-"`
	ShopRarityWeights          []int32 `yaml:"-"`
	ShopRarityUnlockLevels     []int32 `yaml:"-"`
	DecomposeBaseGold          int64   `yaml:"-"`
	DecomposeLevelGold         int64   `yaml:"-"`
	SkillShopDrawsPerLevel     int32   `yaml:"-"`
	CompanionShopDrawsPerLevel int32   `yaml:"-"`

	Stage   ClosedLoopStageConfig   `yaml:"stage"`
	Chest   ClosedLoopChestConfig   `yaml:"chest"`
	Shop    ClosedLoopShopConfig    `yaml:"shop"`
	Economy ClosedLoopEconomyConfig `yaml:"economy"`
}

type ClosedLoopStageConfig struct {
	BoxMin           int32 `yaml:"box_min"`
	BoxMax           int32 `yaml:"box_max"`
	OpenerBaseGold   int64 `yaml:"opener_base_gold"`
	OpenerGrowthPct  int32 `yaml:"opener_growth_pct"`
	MaxLevel         int32 `yaml:"max_level"`
	AttackPerLevel   int64 `yaml:"attack_per_level"`
	HPPerLevel       int64 `yaml:"hp_per_level"`
	RarityBoostEvery int32 `yaml:"rarity_boost_every"`
}

type ClosedLoopChestConfig struct {
	Rarities          []int32 `yaml:"rarities"`
	InitialRarities   int32   `yaml:"initial_rarities"`
	MaxActiveRarities int32   `yaml:"max_active_rarities"`
	UnlockInterval    int32   `yaml:"unlock_interval"`
	LowestBaseWeight  int32   `yaml:"lowest_base_weight"`
	AdjacentGapPct    int32   `yaml:"adjacent_gap_pct"`
	ProgressShiftPct  int32   `yaml:"progress_shift_pct"`
	TopWeightCapPct   int32   `yaml:"top_weight_cap_pct"`
}

type ClosedLoopShopConfig struct {
	Rarities          []int32 `yaml:"rarities"`
	InitialRarities   int32   `yaml:"initial_rarities"`
	MaxActiveRarities int32   `yaml:"max_active_rarities"`
	UnlockInterval    int32   `yaml:"unlock_interval"`
	LowestBaseWeight  int32   `yaml:"lowest_base_weight"`
	AdjacentGapPct    int32   `yaml:"adjacent_gap_pct"`
	ProgressShiftPct  int32   `yaml:"progress_shift_pct"`
	TopWeightCapPct   int32   `yaml:"top_weight_cap_pct"`
}

type ClosedLoopEconomyConfig struct {
	DecomposeBaseGold      int64 `yaml:"decompose_base_gold"`
	DecomposeLevelGold     int64 `yaml:"decompose_level_gold"`
	SkillDrawsPerLevel     int32 `yaml:"skill_draws_per_level"`
	CompanionDrawsPerLevel int32 `yaml:"companion_draws_per_level"`
}

func (c ClosedLoopConfig) StageBoxMinValue() int32             { return c.StageBoxMin }
func (c ClosedLoopConfig) StageBoxMaxValue() int32             { return c.StageBoxMax }
func (c ClosedLoopConfig) OpenerUpgradeBaseGoldValue() int64   { return c.OpenerUpgradeBaseGold }
func (c ClosedLoopConfig) OpenerUpgradeGrowthPctValue() int32  { return c.OpenerUpgradeGrowthPct }
func (c ClosedLoopConfig) OpenerMaxLevelValue() int32          { return c.OpenerMaxLevel }
func (c ClosedLoopConfig) EquipmentAttackPerLevelValue() int64 { return c.EquipmentAttackPerLevel }
func (c ClosedLoopConfig) EquipmentHPPerLevelValue() int64     { return c.EquipmentHPPerLevel }
func (c ClosedLoopConfig) RarityBoostEveryLevelsValue() int32  { return c.RarityBoostEveryLevels }

// StageRow is one flat stage definition.
type StageRow struct {
	FlatStage        int32
	AdventureID      int32
	StageIndex       int32
	RecommendedPower int64
	FirstClearGold   int64
	IsBoss           bool
	BossPowerMult    float64
	MilestoneGold    int64
}

// DropRow is one weighted equipment drop entry.
type DropRow struct {
	ConfigID       int32
	Rarity         int32
	Weight         int
	Slot           int32
	Attack         int64
	HP             int64
	BonusAttackPct int32
}

// Load parses embedded gameplay config files.
func Load() (*Config, error) {
	globals, err := loadGlobals()
	if err != nil {
		return nil, err
	}
	goldPerSec, err := loadGoldPerSec()
	if err != nil {
		return nil, err
	}
	stages, err := loadStages()
	if err != nil {
		return nil, err
	}
	drops, err := loadIdleDrops()
	if err != nil {
		return nil, err
	}
	closedLoop, err := loadClosedLoop()
	if err != nil {
		return nil, err
	}
	if len(drops) == 0 {
		return nil, fmt.Errorf("idle drop table is empty")
	}
	return &Config{
		Globals:           globals,
		ClosedLoop:        closedLoop,
		GoldPerSecByStage: goldPerSec,
		Stages:            stages,
		IdleDrops:         drops,
		StarterWeapon:     drops[0],
	}, nil
}

func loadGlobals() (Globals, error) {
	raw, err := embeddedFS.ReadFile("data/idle/idle_globals.json")
	if err != nil {
		return Globals{}, err
	}
	var g Globals
	if err := json.Unmarshal(raw, &g); err != nil {
		return Globals{}, err
	}
	return g, nil
}

func loadClosedLoop() (ClosedLoopConfig, error) {
	raw, err := embeddedFS.ReadFile("data/economy/chest_opener_config.yaml")
	if err != nil {
		return ClosedLoopConfig{}, err
	}
	var file struct {
		ClosedLoop struct {
			Stage struct {
				BoxMin           int32 `yaml:"box_min"`
				BoxMax           int32 `yaml:"box_max"`
				OpenerBaseGold   int64 `yaml:"opener_base_gold"`
				OpenerGrowthPct  int32 `yaml:"opener_growth_pct"`
				MaxLevel         int32 `yaml:"max_level"`
				AttackPerLevel   int64 `yaml:"attack_per_level"`
				HPPerLevel       int64 `yaml:"hp_per_level"`
				RarityBoostEvery int32 `yaml:"rarity_boost_every"`
			} `yaml:"stage"`
			Chest   ClosedLoopChestConfig `yaml:"chest"`
			Shop    ClosedLoopShopConfig  `yaml:"shop"`
			Economy struct {
				DecomposeBaseGold      int64 `yaml:"decompose_base_gold"`
				DecomposeLevelGold     int64 `yaml:"decompose_level_gold"`
				SkillDrawsPerLevel     int32 `yaml:"skill_draws_per_level"`
				CompanionDrawsPerLevel int32 `yaml:"companion_draws_per_level"`
			} `yaml:"economy"`
		} `yaml:"closed_loop"`
	}
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return ClosedLoopConfig{}, err
	}
	cfg := ClosedLoopConfig{
		StageBoxMin:                file.ClosedLoop.Stage.BoxMin,
		StageBoxMax:                file.ClosedLoop.Stage.BoxMax,
		OpenerUpgradeBaseGold:      file.ClosedLoop.Stage.OpenerBaseGold,
		OpenerUpgradeGrowthPct:     file.ClosedLoop.Stage.OpenerGrowthPct,
		OpenerMaxLevel:             file.ClosedLoop.Stage.MaxLevel,
		EquipmentAttackPerLevel:    file.ClosedLoop.Stage.AttackPerLevel,
		EquipmentHPPerLevel:        file.ClosedLoop.Stage.HPPerLevel,
		RarityBoostEveryLevels:     file.ClosedLoop.Stage.RarityBoostEvery,
		ChestRarityWeights:         nil,
		ChestRarityUnlockLevels:    nil,
		ShopRarityWeights:          nil,
		ShopRarityUnlockLevels:     nil,
		DecomposeBaseGold:          file.ClosedLoop.Economy.DecomposeBaseGold,
		DecomposeLevelGold:         file.ClosedLoop.Economy.DecomposeLevelGold,
		SkillShopDrawsPerLevel:     file.ClosedLoop.Economy.SkillDrawsPerLevel,
		CompanionShopDrawsPerLevel: file.ClosedLoop.Economy.CompanionDrawsPerLevel,
		Stage:                      file.ClosedLoop.Stage,
		Chest:                      file.ClosedLoop.Chest,
		Shop:                       file.ClosedLoop.Shop,
		Economy:                    file.ClosedLoop.Economy,
	}
	return cfg, nil
}

func loadGoldPerSec() (map[int32]float64, error) {
	rows, err := readCSV("data/idle/idle_gold_per_sec.csv")
	if err != nil {
		return nil, err
	}
	out := make(map[int32]float64, len(rows))
	for _, row := range rows {
		flat, err := parseInt32(row[0])
		if err != nil {
			return nil, err
		}
		rate, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return nil, err
		}
		out[flat] = rate
	}
	return out, nil
}

func loadStages() (map[int32]StageRow, error) {
	rows, err := readCSV("data/stages/stages.csv")
	if err != nil {
		return nil, err
	}
	out := make(map[int32]StageRow, len(rows))
	for _, row := range rows {
		flat, err := parseInt32(row[0])
		if err != nil {
			return nil, err
		}
		isBoss, _ := strconv.ParseBool(strings.TrimSpace(row[5]))
		bossMult, err := strconv.ParseFloat(strings.TrimSpace(row[6]), 64)
		if err != nil {
			return nil, err
		}
		milestoneGold, err := parseInt64(row[7])
		if err != nil {
			return nil, err
		}
		recPower, err := parseInt64(row[3])
		if err != nil {
			return nil, err
		}
		firstGold, err := parseInt64(row[4])
		if err != nil {
			return nil, err
		}
		adv, err := parseInt32(row[1])
		if err != nil {
			return nil, err
		}
		idx, err := parseInt32(row[2])
		if err != nil {
			return nil, err
		}
		out[flat] = StageRow{
			FlatStage:        flat,
			AdventureID:      adv,
			StageIndex:       idx,
			RecommendedPower: recPower,
			FirstClearGold:   firstGold,
			IsBoss:           isBoss,
			BossPowerMult:    bossMult,
			MilestoneGold:    milestoneGold,
		}
	}
	for flat := int32(1); flat <= 30; flat++ {
		if _, ok := out[flat]; !ok {
			return nil, fmt.Errorf("missing stage row for flat_stage %d", flat)
		}
	}
	return out, nil
}

func loadIdleDrops() ([]DropRow, error) {
	rows, err := readCSV("data/economy/drop_table_idle.csv")
	if err != nil {
		return nil, err
	}
	out := make([]DropRow, 0, len(rows))
	for _, row := range rows {
		configID, err := parseInt32(row[0])
		if err != nil {
			return nil, err
		}
		rarity, err := parseInt32(row[1])
		if err != nil {
			return nil, err
		}
		weight, err := strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil {
			return nil, err
		}
		slot, err := parseInt32(row[3])
		if err != nil {
			return nil, err
		}
		atk, err := parseInt64(row[4])
		if err != nil {
			return nil, err
		}
		hp, err := parseInt64(row[5])
		if err != nil {
			return nil, err
		}
		bonus, err := parseInt32(row[6])
		if err != nil {
			return nil, err
		}
		out = append(out, DropRow{
			ConfigID:       configID,
			Rarity:         rarity,
			Weight:         weight,
			Slot:           slot,
			Attack:         atk,
			HP:             hp,
			BonusAttackPct: bonus,
		})
	}
	return out, nil
}

func readCSV(path string) ([][]string, error) {
	raw, err := embeddedFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(raw)))
	r.TrimLeadingSpace = true
	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(all) < 2 {
		return nil, fmt.Errorf("%s: no data rows", path)
	}
	return all[1:], nil
}

func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	return int32(v), err
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// GoldPerSec returns base gold per second for a flat stage key.
func (c *Config) GoldPerSec(highestStageCleared int32) float64 {
	key := highestStageCleared
	if key < 0 {
		key = 0
	}
	if key > 30 {
		key = 30
	}
	if rate, ok := c.GoldPerSecByStage[key]; ok {
		return rate
	}
	return c.GoldPerSecByStage[0]
}

// Stage returns the row for a flat stage (1-30).
func (c *Config) Stage(flat int32) (StageRow, bool) {
	row, ok := c.Stages[flat]
	return row, ok
}

// RecommendedPower returns gate power for a flat stage including boss multiplier.
func (c *Config) RecommendedPower(flat int32) int64 {
	row, ok := c.Stages[flat]
	if !ok {
		return 0
	}
	power := float64(row.RecommendedPower)
	if row.IsBoss && row.BossPowerMult > 0 {
		power *= row.BossPowerMult
	}
	return int64(power)
}

// StageIdleScale returns idle efficiency multiplier from highest stage cleared.
func (c *Config) StageIdleScale(highestStageCleared int32) float64 {
	if highestStageCleared <= 0 {
		return 1.0
	}
	return 1 + float64(highestStageCleared-1)*c.Globals.IdleScalePerStage
}

// ClearThreshold returns the combat power ratio required to clear a stage.
func (c *Config) ClearThreshold() float64 {
	if c.Globals.ClearThreshold <= 0 {
		return 1.0
	}
	return c.Globals.ClearThreshold
}

// MilestoneFlats are boss milestone flat stages.
var MilestoneFlats = []int32{5, 10, 15, 20, 25, 30}

// IsMilestone reports whether a flat stage is a milestone boss.
func IsMilestone(flat int32) bool {
	for _, m := range MilestoneFlats {
		if m == flat {
			return true
		}
	}
	return false
}

// PickIdleDrop selects a drop row by weighted roll (deterministic from seed).
func (c *Config) PickIdleDrop(seed int64) DropRow {
	total := 0
	for _, row := range c.IdleDrops {
		total += row.Weight
	}
	if total == 0 {
		return c.IdleDrops[0]
	}
	pick := int(seed % int64(total))
	if pick < 0 {
		pick = -pick
	}
	cursor := 0
	for _, row := range c.IdleDrops {
		cursor += row.Weight
		if pick < cursor {
			return row
		}
	}
	return c.IdleDrops[len(c.IdleDrops)-1]
}
