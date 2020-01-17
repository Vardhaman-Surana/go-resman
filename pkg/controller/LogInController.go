package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/encryption"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
)

const (
	Success = "Success"
	Fail    = "Fail"
)

type LogInController struct {
	database.Database
}

func NewLogInController(db database.Database) *LogInController {
	lc := new(LogInController)
	lc.Database = db
	return lc
}
func (l *LogInController) LogIn(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	var cred models.Credentials
	logger.LogDebug(reqId, reqUrl, "parsing request body")
	err := c.ShouldBindJSON(&cred)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "checking for valid user type")
	isValid := middleware.IsValidUserType(cred.Role)
	if !isValid {
		logger.LogError(reqId, reqUrl, "invalid user type", http.StatusBadRequest)
		c.Status(http.StatusBadRequest)
		return
	}
	logger.LogDebug(reqId, reqUrl, "authenticating user from db")
	userID, err := l.LogInUser(c.Request.Context(), &cred)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("unauthenticated user: %v", err), http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  err.Error(),
			"status": Fail,
		})
		return
	}
	claims := &models.Claims{
		ID:   userID,
		Role: cred.Role,
	}
	logger.LogDebug(reqId, reqUrl, "creating token for user")
	token, err := encryption.CreateToken(c.Request.Context(), claims)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in token generation:%v", err), http.StatusInternalServerError)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.LogInfo(reqId, reqUrl, "User logged in successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"role":   cred.Role,
		"msg":    "Login Successful",
		"status": Success,
	})
}

func (l *LogInController) LogOut(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	logger.LogDebug(reqId, reqUrl, "getting token from header")
	tokenStr := c.Request.Header.Get("token")
	if tokenStr == "" {
		logger.LogError(reqId, reqUrl, "empty token in request header", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty token",
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "storing token in db")
	err := l.StoreToken(c.Request.Context(), tokenStr)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("can not store token in db:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		c.Status(http.StatusInternalServerError)
		return
	}
	logger.LogInfo(reqId, reqUrl, "user logged out successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg":    "Logged Out Successfully",
		"status": Success,
	})
}
