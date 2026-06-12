package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/delicious/delicious/internal/config"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

// Auth 个人使用：优先 JWT（Bearer），否则使用默认 user_id
func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := cfg.DefaultUID
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// TODO: 完整 JWT 校验；当前解析简单 payload 或 fallback default
			token := strings.TrimPrefix(auth, "Bearer ")
			if parsed, err := strconv.ParseUint(token, 10, 64); err == nil && parsed > 0 {
				uid = parsed
			}
		}
		c.Set(UserIDKey, uid)
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

func Forbidden(c *gin.Context) {
	JSONError(c, http.StatusForbidden, "无权访问")
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	c.Abort()
}
