package player

import (
	"encoding/json"

	commonv1 "github.com/slimeyquest/proto/gen/go/common"
	equipmentv1 "github.com/slimeyquest/proto/gen/go/equipment"
	"github.com/slimeyquest/server/internal/gameplayconfig"
)

// EquipmentInstance is one owned equipment item.
type EquipmentInstance struct {
	UID            int64 `json:"uid"`
	ConfigID       int32 `json:"config_id"`
	Slot           int32 `json:"slot"`
	Rarity         int32 `json:"rarity"`
	Level          int32 `json:"level"`
	Attack         int64 `json:"attack"`
	HP             int64 `json:"hp"`
	BonusAttackPct int32 `json:"bonus_attack_pct"`
}

// EquipmentData stores owned items and equipped slot mapping.
type EquipmentData struct {
	NextUID   int64                        `json:"next_uid"`
	Instances map[int64]EquipmentInstance  `json:"instances"`
	Equipped  map[int32]int64              `json:"equipped"`
}

// DecodeEquipment parses the ent JSON blob into EquipmentData.
func DecodeEquipment(raw map[string]interface{}) EquipmentData {
	if raw == nil || len(raw) == 0 {
		return EquipmentData{
			Instances: make(map[int64]EquipmentInstance),
			Equipped:  make(map[int32]int64),
		}
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return EquipmentData{
			Instances: make(map[int64]EquipmentInstance),
			Equipped:  make(map[int32]int64),
		}
	}
	var data EquipmentData
	if err := json.Unmarshal(b, &data); err != nil {
		return EquipmentData{
			Instances: make(map[int64]EquipmentInstance),
			Equipped:  make(map[int32]int64),
		}
	}
	if data.Instances == nil {
		data.Instances = make(map[int64]EquipmentInstance)
	}
	if data.Equipped == nil {
		data.Equipped = make(map[int32]int64)
	}
	return data
}

// EncodeEquipment serializes EquipmentData for ent JSON storage.
func EncodeEquipment(data EquipmentData) map[string]interface{} {
	b, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

// StarterEquipment creates default loadout for a new player.
func StarterEquipment(cfg *gameplayconfig.Config) EquipmentData {
	row := cfg.StarterWeapon
	uid := int64(1)
	inst := EquipmentInstance{
		UID:            uid,
		ConfigID:       row.ConfigID,
		Slot:           row.Slot,
		Rarity:         row.Rarity,
		Level:          1,
		Attack:         100,
		HP:             0,
		BonusAttackPct: row.BonusAttackPct,
	}
	if inst.Attack == 0 {
		inst.Attack = row.Attack
	}
	return EquipmentData{
		NextUID:   uid + 1,
		Instances: map[int64]EquipmentInstance{uid: inst},
		Equipped: map[int32]int64{
			int32(equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_WEAPON): uid,
		},
	}
}

// AddInstance appends a new equipment instance and returns it.
func (d *EquipmentData) AddInstance(row gameplayconfig.DropRow) EquipmentInstance {
	if d.NextUID == 0 {
		d.NextUID = 1
	}
	uid := d.NextUID
	d.NextUID++
	inst := EquipmentInstance{
		UID:            uid,
		ConfigID:       row.ConfigID,
		Slot:           row.Slot,
		Rarity:         row.Rarity,
		Level:          1,
		Attack:         row.Attack,
		HP:             row.HP,
		BonusAttackPct: row.BonusAttackPct,
	}
	if d.Instances == nil {
		d.Instances = make(map[int64]EquipmentInstance)
	}
	d.Instances[uid] = inst
	return inst
}

// EquippedSlots returns protobuf equipped slot entries.
func (d *EquipmentData) EquippedSlots() []*equipmentv1.EquippedSlot {
	slots := []equipmentv1.EquipmentSlot{
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_WEAPON,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_ARMOR,
		equipmentv1.EquipmentSlot_EQUIPMENT_SLOT_RING,
	}
	out := make([]*equipmentv1.EquippedSlot, 0, len(slots))
	for _, slot := range slots {
		uid := d.Equipped[int32(slot)]
		out = append(out, &equipmentv1.EquippedSlot{
			Slot:         slot,
			EquipmentUid: uid,
		})
	}
	return out
}

// ToProto converts an instance to EquipmentInfo.
func (inst EquipmentInstance) ToProto() *equipmentv1.EquipmentInfo {
	return &equipmentv1.EquipmentInfo{
		EquipmentUid: inst.UID,
		ConfigId:     inst.ConfigID,
		Slot:         equipmentv1.EquipmentSlot(inst.Slot),
		Rarity:       commonv1.EquipmentRarity(inst.Rarity),
		Level:        inst.Level,
		Stats: &equipmentv1.EquipmentStats{
			Attack:         inst.Attack,
			Hp:             inst.HP,
			BonusAttackPct: inst.BonusAttackPct,
		},
	}
}
