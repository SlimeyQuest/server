package entity

// Hero class identifiers.
const (
	HeroClassWarrior = "WARRIOR"
)

// SkillInfo is an MVP skill entry.
type SkillInfo struct {
	SkillID int32  `json:"skillId"`
	Name    string `json:"name"`
	Quality int32  `json:"quality"`
}

// CompanionInfo is an MVP companion entry.
type CompanionInfo struct {
	CompanionID int32  `json:"companionId"`
	Name        string `json:"name"`
	Quality     int32  `json:"quality"`
}

// PlayerProfile is the player snapshot returned by login and gameplay APIs.
type PlayerProfile struct {
	PlayerID            int64           `json:"playerId"`
	DisplayName         string          `json:"displayName"`
	Gold                int64           `json:"gold"`
	Gems                int64           `json:"gems"`
	CombatPower         int64           `json:"combatPower"`
	AdventureID         int32           `json:"adventureId"`
	StageIndex          int32           `json:"stageIndex"`
	HighestStageCleared int32           `json:"highestStageCleared"`
	EquippedSlots       []EquippedSlot  `json:"equippedSlots,omitempty"`
	CreatedAtMs         int64           `json:"createdAtMs"`
	LastLoginAtMs       int64           `json:"lastLoginAtMs"`
	HeroClass           string          `json:"heroClass"`
	HeroLevel           int32           `json:"heroLevel"`
	ZoneID              int32           `json:"zoneId"`
	ProfessionSkill     *SkillInfo      `json:"professionSkill,omitempty"`
	EquippedSkills      []SkillInfo     `json:"equippedSkills,omitempty"`
	Companions          []CompanionInfo `json:"companions,omitempty"`
	ChestLevel          int32           `json:"chestLevel"`
	BoxCount            int32           `json:"boxCount"`
}

// CreateRoleReq updates the display name.
type CreateRoleReq struct {
	DisplayName string `json:"displayName"`
}

// CreateRoleRes returns the updated profile.
type CreateRoleRes struct {
	Error   *ErrorInfo     `json:"error,omitempty"`
	Profile *PlayerProfile `json:"profile,omitempty"`
}

// DrawSkillReq requests skill shop draws.
type DrawSkillReq struct {
	DrawCount int32 `json:"drawCount"`
}

// DrawSkillRes returns drawn skills.
type DrawSkillRes struct {
	Error     *ErrorInfo  `json:"error,omitempty"`
	Rewards   []SkillInfo `json:"rewards,omitempty"`
	ShopLevel int32       `json:"shopLevel"`
}

// DrawCompanionReq requests companion shop draws.
type DrawCompanionReq struct {
	DrawCount int32 `json:"drawCount"`
}

// DrawCompanionRes returns drawn companions.
type DrawCompanionRes struct {
	Error     *ErrorInfo      `json:"error,omitempty"`
	Rewards   []CompanionInfo `json:"rewards,omitempty"`
	ShopLevel int32           `json:"shopLevel"`
}
