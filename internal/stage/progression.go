package stage

// CanClear reports whether combat power meets the recommended gate.
func CanClear(combatPower, recommended int64, threshold float64) bool {
	if threshold <= 0 {
		threshold = 1.0
	}
	required := int64(float64(recommended) * threshold)
	return combatPower >= required
}

// IsUnlocked reports whether the target flat stage can be attempted.
func IsUnlocked(highestStageCleared, targetFlat int32) bool {
	return targetFlat == highestStageCleared+1
}

// IsCurrentTarget reports whether target index matches the current stage challenge.
func IsCurrentTarget(adventureID, stageIndex, targetStageIndex int32) bool {
	return targetStageIndex == stageIndex && targetStageIndex >= 1
}
