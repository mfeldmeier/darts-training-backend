package middleware

import (
	"net/http"
	"strings"
	"time"

	"darts-training-app/internal/models"
	"darts-training-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func CheckAuth(authManager *services.AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		token := strings.Split(authHeader, "Bearer ")

		if len(token) < 2 || token[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		jwks, err := authManager.GetJWKS()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to get JWKS",
				"code":  "JWKS_ERROR",
			})
			return
		}

		userToken := models.UserToken{}
		_, err = jwt.ParseWithClaims(token[1], &userToken, jwks.Keyfunc)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "JWKS_ERROR",
			})
			return
		}

		if !userToken.VerifyExpiresAt(time.Now(), true) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token has expired",
				"code":  "Auth Error",
			})
			return
		}
		c.Set("User", userToken)
		c.Next()
	}
}

//func OptionalAuthMiddleware(authManager *services.AuthManager) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// Get token from Authorization header
//		authHeader := c.GetHeader("Authorization")
//		if authHeader == "" {
//			c.Next()
//			return
//		}
//
//		// Check if the header has the Bearer prefix
//		if !strings.HasPrefix(authHeader, "Bearer ") {
//			c.Next()
//			return
//		}
//
//		// Extract the token
//		token := strings.TrimPrefix(authHeader, "Bearer ")
//		if token == "" {
//			c.Next()
//			return
//		}
//
//		// Parse and validate the token
//		userToken := &models.UserToken{}
//		_, err := jwt.ParseWithClaims(token, userToken, func(token *jwt.Token) (interface{}, error) {
//			// For optional auth, we'll just verify the structure
//			return []byte("dummy-key"), nil // We'll need proper JWKS for this
//		})
//		if err != nil {
//			c.Next()
//			return
//		}
//
//		// Set user information in the context
//		c.Set("user_id", userToken.Subject)
//		c.Set("user_email", userToken.Email)
//		c.Set("user_name", userToken.Name)
//		c.Set("user_nickname", userToken.Nickname)
//
//		c.Next()
//	}
//}
