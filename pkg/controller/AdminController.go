package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"strings"
)

type AdminController struct {
	database.Database
}

func NewAdminController(db database.Database) *AdminController {
	ac := new(AdminController)
	ac.Database = db
	return ac
}

func (a *AdminController) GetAdmins(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)

	jsonData := &[]models.UserOutput{}
	var stringData string
	var err error
	logger.LogDebug(reqId, reqUrl, "getting admins from db")
	stringData, err = a.ShowAdmins(c.Request.Context())
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in retrieving admins from db:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	if stringData != "" {
		_ = json.Unmarshal([]byte(stringData), jsonData)
		logger.LogInfo(reqId, reqUrl, "admins retrieved  successful", http.StatusOK)
		c.JSON(http.StatusOK, jsonData)

	} else {
		logger.LogInfo(reqId, reqUrl, "no admins to show", http.StatusOK)
		c.JSON(http.StatusOK, []models.UserOutput{})
	}
}

func (a *AdminController) EditAdmin(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	adminID := c.Param("adminID")
	var admin models.UserOutput
	admin.ID = adminID
	logger.LogDebug(reqId, reqUrl, "parsing request body")
	err := c.ShouldBindJSON(&admin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body: %v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "checking requested admin id")
	err = a.CheckAdmin(c.Request.Context(), admin.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("requested admin id does not exist: %v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Admin does not exist",
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "updating requested admin")
	adminString, err := a.UpdateAdmin(c.Request.Context(), &admin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("can not update admin: %v", err), http.StatusInternalServerError)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData := &models.UserOutput{}
	_ = json.Unmarshal([]byte(adminString), jsonData)
	logger.LogInfo(reqId, reqUrl, "admin updated successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (a *AdminController) DeleteAdmins(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "parsing query parameter")
	multipleIdString := c.Request.URL.Query().Get("id")
	if multipleIdString == "" {
		logger.LogError(reqId, reqUrl, "absent query parameter", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty id parameter",
		})
		return
	}
	idArr := strings.Split(multipleIdString, ",")

	//var adminID struct {
	//	IDArr []string	`json:"idArr" binding:"required"`
	//}
	//err:=c.ShouldBindJSON(&adminID)
	//if err!=nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": err.Error(),
	//	})
	//	return
	//}
	logger.LogDebug(reqId, reqUrl, "deleting requested admins")
	err := a.RemoveAdmins(c.Request.Context(), idArr...)
	if err != nil {
		if err != database.ErrInternal {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("admin does not exist:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Admin does not exist please refresh",
			})
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.LogInfo(reqId, reqUrl, "admin deleted successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Admins deleted successfully",
	})
}
