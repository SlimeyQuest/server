package apitypes

// GuestLoginReq authenticates a guest device.
type GuestLoginReq struct {
	DeviceID      string `json:"deviceId"`
	ClientVersion string `json:"clientVersion"`
}

// PhoneRegisterReq registers or resumes a phone account.
type PhoneRegisterReq struct {
	Phone         string `json:"phone"`
	VerifyCode    string `json:"verifyCode"`
	ClientVersion string `json:"clientVersion"`
}

// PhoneLoginReq logs in a phone account.
type PhoneLoginReq struct {
	Phone         string `json:"phone"`
	VerifyCode    string `json:"verifyCode"`
	ClientVersion string `json:"clientVersion"`
}

// AuthResponse is returned by all login endpoints on success.
type AuthResponse struct {
	Error        *ErrorInfo     `json:"error,omitempty"`
	SessionToken string         `json:"sessionToken,omitempty"`
	PlayerID     int64          `json:"playerId,omitempty"`
	Profile      *PlayerProfile `json:"profile,omitempty"`
	IdleState    *IdleState     `json:"idleState,omitempty"`
	StageState   *StageState    `json:"stageState,omitempty"`
}
