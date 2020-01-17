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

type RegisterController struct {
	database.Database
}

func NewRegisterController(db database.Database) *RegisterController {
	regController := new(RegisterController)
	regController.Database = db
	return regController
}

func (r *RegisterController) Register(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	var user models.UserReg
	logger.LogDebug(reqId, reqUrl, "parsing request body")
	err := c.ShouldBindJSON(&user)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": Fail,
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "validating user role")
	if user.Role != middleware.Admin && user.Role != middleware.SuperAdmin {
		logger.LogError(reqId, reqUrl, "invalid role ", http.StatusNotFound)
		c.Status(http.StatusNotFound)
		return
	}
	logger.LogDebug(reqId, reqUrl, "creating user")
	userId, err := r.CreateUser(c.Request.Context(), &user)
	if err != nil {
		if err == database.ErrDupEmail {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in creating user:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": Fail,
			})
			return
		}
		logger.LogError(reqId, reqUrl, fmt.Sprintf("can not create user: %v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	claims := &models.Claims{
		ID:   userId,
		Role: user.Role,
	}
	logger.LogDebug(reqId, reqUrl, "creating owner")
	token, err := encryption.CreateToken(c.Request.Context(), claims)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.LogInfo(reqId, reqUrl, "user registration successful", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"role":   user.Role,
		"token":  token,
		"msg":    "Registration Successful",
		"status": Success,
	})
}
