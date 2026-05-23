package entity

import "strings"

// Equipment slot identifiers (numeric wire values preserved from legacy proto).
const (
	SlotHat        int32 = 1
	SlotShoeLeft   int32 = 2
	SlotShoeRight  int32 = 3
	SlotGloveLeft  int32 = 4
	SlotGloveRight int32 = 5
	SlotCloth      int32 = 6
	SlotPants      int32 = 7
	SlotWeapon     int32 = 8
	SlotRingLeft   int32 = 9
	SlotRingRight  int32 = 10
)

// AllEquipmentSlots lists every slot in display order.
var AllEquipmentSlots = []int32{
	SlotHat, SlotShoeLeft, SlotShoeRight, SlotGloveLeft, SlotGloveRight,
	SlotCloth, SlotPants, SlotWeapon, SlotRingLeft, SlotRingRight,
}

// SlotName returns the REST slot label for a slot id.
func SlotName(slot int32) string {
	switch slot {
	case SlotHat:
		return "HAT"
	case SlotShoeLeft:
		return "SHOE_LEFT"
	case SlotShoeRight:
		return "SHOE_RIGHT"
	case SlotGloveLeft:
		return "GLOVE_LEFT"
	case SlotGloveRight:
		return "GLOVE_RIGHT"
	case SlotCloth:
		return "CLOTH"
	case SlotPants:
		return "PANTS"
	case SlotWeapon:
		return "WEAPON"
	case SlotRingLeft:
		return "RING_LEFT"
	case SlotRingRight:
		return "RING_RIGHT"
	default:
		return "UNSPECIFIED"
	}
}

// ParseEquipmentSlot parses a REST slot name. Empty string means unspecified (0).
func ParseEquipmentSlot(name string) (int32, bool) {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "", "UNSPECIFIED":
		return 0, true
	case "HAT":
		return SlotHat, true
	case "SHOE_LEFT":
		return SlotShoeLeft, true
	case "SHOE_RIGHT":
		return SlotShoeRight, true
	case "GLOVE_LEFT":
		return SlotGloveLeft, true
	case "GLOVE_RIGHT":
		return SlotGloveRight, true
	case "CLOTH":
		return SlotCloth, true
	case "PANTS":
		return SlotPants, true
	case "WEAPON":
		return SlotWeapon, true
	case "RING_LEFT":
		return SlotRingLeft, true
	case "RING_RIGHT":
		return SlotRingRight, true
	default:
		return 0, false
	}
}

// EquippedSlot is one equipment slot on the player profile.
type EquippedSlot struct {
	Slot         string `json:"slot"`
	EquipmentUID int64  `json:"equipmentUid"`
}

// EquipmentStats holds combat stats for one item.
type EquipmentStats struct {
	Attack         int64 `json:"attack"`
	HP             int64 `json:"hp"`
	BonusAttackPct int32 `json:"bonusAttackPct"`
}

// EquipmentInfo describes one owned equipment instance.
type EquipmentInfo struct {
	EquipmentUID int64           `json:"equipmentUid"`
	ConfigID     int32           `json:"configId"`
	Slot         string          `json:"slot"`
	Rarity       int32           `json:"rarity"`
	Level        int32           `json:"level"`
	Stats        *EquipmentStats `json:"stats,omitempty"`
}

// StageBoxReward is granted when clearing a stage.
type StageBoxReward struct {
	BoxCount      int32 `json:"boxCount"`
	TotalBoxCount int32 `json:"totalBoxCount"`
}

// ChestOpenReq opens chests.
type ChestOpenReq struct {
	Count int32 `json:"count"`
}

// ChestOpenRes returns opened equipment.
type ChestOpenRes struct {
	Error             *ErrorInfo      `json:"error,omitempty"`
	Equipment         []EquipmentInfo `json:"equipment,omitempty"`
	RemainingBoxCount int32           `json:"remainingBoxCount"`
}

// EquipItemReq equips an item.
type EquipItemReq struct {
	EquipmentUID int64  `json:"equipmentUid"`
	Slot         string `json:"slot"`
}

// EquipItemRes returns updated loadout.
type EquipItemRes struct {
	Error         *ErrorInfo     `json:"error,omitempty"`
	EquippedSlots []EquippedSlot `json:"equippedSlots,omitempty"`
	CombatPower   int64          `json:"combatPower"`
}

// DecomposeEquipmentReq decomposes one item.
type DecomposeEquipmentReq struct {
	EquipmentUID int64 `json:"equipmentUid"`
}

// DecomposeEquipmentRes returns gold gained.
type DecomposeEquipmentRes struct {
	Error      *ErrorInfo `json:"error,omitempty"`
	GainedGold int64      `json:"gainedGold"`
	TotalGold  int64      `json:"totalGold"`
}

// UpgradeChestReq upgrades chest opener level.
type UpgradeChestReq struct {
	TargetLevel int32 `json:"targetLevel"`
}

// UpgradeChestRes returns chest level after upgrade.
type UpgradeChestRes struct {
	Error      *ErrorInfo `json:"error,omitempty"`
	ChestLevel int32      `json:"chestLevel"`
	TotalGold  int64      `json:"totalGold"`
}
