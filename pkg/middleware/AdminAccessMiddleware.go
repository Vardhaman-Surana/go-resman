package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
)

func AdminAccessOnly(c *gin.Context) {
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	if userAuth.Role != Admin && userAuth.Role != SuperAdmin {
		err := "only for  admin or superAdmin access"
		reqId := c.Request.Context().Value("reqId")
		logger.LogError(reqId.(string), c.Request.URL.String(), err, http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		c.Abort()
	}
	c.Next()
}

func SuperAdminAccessOnly(c *gin.Context) {
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	if userAuth.Role != SuperAdmin {
		err := "only for superAdmin access"
		reqId := c.Request.Context().Value("reqId")
		logger.LogError(reqId.(string), c.Request.URL.String(), err, http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err,
		})
		c.Abort()
	}
	c.Next()
}

func SetResponseHeader(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "applicatoin/json")
	c.Next()
}
func AllowOptions(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Token")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	if c.Request.Method == http.MethodOptions {
		c.Writer.WriteHeader(http.StatusOK)
		c.Abort()
	}
	c.Next()
}
