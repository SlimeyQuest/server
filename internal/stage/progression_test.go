package stage_test

import (
	"testing"

	"github.com/slimeyquest/server/internal/stage"
)

func TestFlatStageMapping(t *testing.T) {
	if got := stage.FlatStage(2, 3); got != 13 {
		t.Fatalf("expected flat 13, got %d", got)
	}
	adv, idx := stage.FromFlatStage(13)
	if adv != 2 || idx != 3 {
		t.Fatalf("unexpected reverse mapping: %d %d", adv, idx)
	}
}

func TestCanClear(t *testing.T) {
	if !stage.CanClear(100, 100, 1.0) {
		t.Fatal("expected clear at equal power")
	}
	if stage.CanClear(90, 100, 1.0) {
		t.Fatal("expected fail below threshold")
	}
}

func TestIsUnlocked(t *testing.T) {
	if !stage.IsUnlocked(0, 1) {
		t.Fatal("expected flat 1 unlocked from zero")
	}
	if stage.IsUnlocked(0, 2) {
		t.Fatal("expected flat 2 locked from zero")
	}
}
