package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"strings"
)

const (
	Admin              = "admin"
	SuperAdmin         = "superAdmin"
	Owner              = "owner"
	tokenExpireMessage = "Token expired please login again"
	statusTokenInvalid = 498
)

func TokenValidator(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.Request.Context().Value("reqId")
		reqUrl := c.Request.URL.String()
		tokenStr := c.Request.Header.Get("token")
		logger.LogDebug(reqId.(string), reqUrl, "checking if the token is logged out validity")
		isValid := db.VerifyToken(c.Request.Context(), tokenStr)
		if !isValid {
			err := "invalid token"
			logger.LogError(reqId.(string), reqUrl, err, statusTokenInvalid)
			c.JSON(statusTokenInvalid, gin.H{
				"error": err,
			})
			c.Abort()
		}
		c.Next()
	}
}

func AuthMiddleware(c *gin.Context) {
	jwtKey := []byte("SecretKey")
	reqId, _ := c.Get("reqId")
	reqUrl := c.Request.URL.String()
	logger.LogDebug(reqId.(string), reqUrl, "checking token validity")
	tokenStr := c.Request.Header.Get("token")
	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("auth token signature not valid: %v", err), statusTokenInvalid)
			c.JSON(statusTokenInvalid, gin.H{
				"error": err.Error(),
			})
			c.Abort()
		}
		if strings.Contains(err.Error(), "expired") {
			logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("auth token expired: %v", err), statusTokenInvalid)
			c.JSON(statusTokenInvalid, gin.H{
				"error": tokenExpireMessage,
			})
			c.Abort()
		}
		logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("error in parsing auth token: %v", err), statusTokenInvalid)
		c.JSON(statusTokenInvalid, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	if !tkn.Valid {
		logger.LogError(reqId.(string), reqUrl, "Invalid Token", statusTokenInvalid)
		c.JSON(statusTokenInvalid, gin.H{
			"error": "Invalid Token",
		})
		c.Abort()
	}
	isValid := IsValidUserType(claims.Role)
	if !isValid {
		logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("Invalid role:%v", claims.Role), statusTokenInvalid)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid role",
		})
		c.Abort()
	}
	userAuth := &models.UserAuth{
		ID:   claims.ID,
		Role: claims.Role,
	}
	c.Set("userAuth", userAuth)
	c.Next()
}

func IsValidUserType(userType string) bool {
	if userType != Admin && userType != SuperAdmin && userType != Owner {
		return false
	}
	return true
}
