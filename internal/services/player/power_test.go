package player_test

import (
	"testing"

	"github.com/slimeyquest/server/internal/entity"
	"github.com/slimeyquest/server/internal/config"
	"github.com/slimeyquest/server/internal/services/player"
)

func TestComputeCombatPowerStarter(t *testing.T) {
	cfg, err := config.LoadGameplay()
	if err != nil {
		t.Fatal(err)
	}
	state := &player.ProgressState{
		Level: 1,
		Equipment: player.EquipmentData{
			Instances: map[int64]player.EquipmentInstance{
				1: {UID: 1, Attack: 100, Slot: entity.SlotWeapon},
			},
			Equipped: map[int32]int64{
				entity.SlotWeapon: 1,
			},
		},
	}
	power := player.ComputeCombatPower(state, cfg)
	if power < 100 {
		t.Fatalf("expected at least 100 power, got %d", power)
	}
}
