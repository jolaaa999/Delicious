package middleware

import (
	"net/http"

	"github.com/delicious/delicious/internal/config"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

// InjectOwner 个人使用：固定 owner，无需登录
func InjectOwner(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(UserIDKey, cfg.DefaultUID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(UserIDKey); ok {
		if uid, ok := v.(uint64); ok {
			return uid
		}
	}
	return 1
}

func JSONError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"message": msg})
	c.Abort()
}

func NotFound(c *gin.Context, msg string) {
	JSONError(c, http.StatusNotFound, msg)
}

func BadRequest(c *gin.Context, msg string) {
	JSONError(c, http.StatusBadRequest, msg)
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	c.Abort()
}
