package stage

// FlatStage converts adventure and stage index to a flat stage number (1-30).
func FlatStage(adventureID, stageIndex int32) int32 {
	if adventureID < 1 {
		adventureID = 1
	}
	if stageIndex < 1 {
		stageIndex = 1
	}
	return (adventureID-1)*10 + stageIndex
}

// FromFlatStage converts flat stage back to adventure and index.
func FromFlatStage(flat int32) (adventureID, stageIndex int32) {
	if flat < 1 {
		return 1, 1
	}
	adventureID = (flat-1)/10 + 1
	stageIndex = (flat-1)%10 + 1
	return adventureID, stageIndex
}

// AdvanceStage moves current coordinates to the next stage after a clear.
func AdvanceStage(adventureID, stageIndex int32) (int32, int32) {
	if stageIndex < 10 {
		return adventureID, stageIndex + 1
	}
	return adventureID + 1, 1
}

// ChallengeFlat returns the flat stage currently being challenged.
func ChallengeFlat(highestStageCleared int32) int32 {
	if highestStageCleared >= 30 {
		return 30
	}
	return highestStageCleared + 1
}
