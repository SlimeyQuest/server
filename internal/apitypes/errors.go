package apitypes

// Error codes for HTTP JSON responses.
const (
	ErrorCodeOK              = "OK"
	ErrorCodeInvalidRequest  = "INVALID_REQUEST"
	ErrorCodeUnauthorized    = "UNAUTHORIZED"
	ErrorCodeNotFound        = "NOT_FOUND"
	ErrorCodeInternal        = "INTERNAL"
)

// ErrorInfo is the standard API error payload.
type ErrorInfo struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// HasError reports whether the error represents a failure.
func HasError(e *ErrorInfo) bool {
	if e == nil {
		return false
	}
	return e.Code != "" && e.Code != ErrorCodeOK
}

// Err returns a populated ErrorInfo.
func Err(code, message string) *ErrorInfo {
	return &ErrorInfo{Code: code, Message: message}
}
