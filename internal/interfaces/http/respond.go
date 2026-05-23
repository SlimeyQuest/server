package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/apitypes"
)

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": apitypes.Err(code, message)})
}

func writeBadRequest(c *gin.Context, message string) {
	writeError(c, http.StatusBadRequest, apitypes.ErrorCodeInvalidRequest, message)
}

func writeUnauthorized(c *gin.Context) {
	writeError(c, http.StatusUnauthorized, apitypes.ErrorCodeUnauthorized, "unauthorized")
}

func writeInternal(c *gin.Context) {
	writeError(c, http.StatusInternalServerError, apitypes.ErrorCodeInternal, "internal error")
}

func hasAPIError(err *apitypes.ErrorInfo) bool {
	return apitypes.HasError(err)
}
