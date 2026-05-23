package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimeyquest/server/internal/entity"
)

func WriteError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": entity.Err(code, message)})
}

func WriteBadRequest(c *gin.Context, message string) {
	WriteError(c, http.StatusBadRequest, entity.ErrorCodeInvalidRequest, message)
}

func WriteUnauthorized(c *gin.Context) {
	WriteError(c, http.StatusUnauthorized, entity.ErrorCodeUnauthorized, "unauthorized")
}

func WriteInternal(c *gin.Context) {
	WriteError(c, http.StatusInternalServerError, entity.ErrorCodeInternal, "internal error")
}

func HasAPIError(err *entity.ErrorInfo) bool {
	return entity.HasError(err)
}
