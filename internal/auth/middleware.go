package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

//middleware requires a valid "Authorization: Bearer <token>"header
//and stores the user id in the request context for handlers to read

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		userID, err := ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

//THIS SECTION PROTECTS ROUTES MEANT FOR YOUR OWN MEDIA SERVER NOT FOR USERS

func InternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("INTERNAL_SECRET")
		if secret == "" || c.GetHeader("X-Internal-Secret") != secret {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
