package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"strings"
)

type OwnerController struct {
	database.Database
}

func NewOwnerController(db database.Database) *OwnerController {
	ownerController := new(OwnerController)
	ownerController.Database = db
	return ownerController
}
func (o *OwnerController) GetOwners(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	jsonData := &[]models.UserOutput{}
	var stringData string
	var err error
	logger.LogDebug(reqId, reqUrl, "retrieving owners from db")
	stringData, err = o.ShowOwners(c.Request.Context(), userAuth)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("can not get owners: %v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "empty owner list", http.StatusOK)
		c.JSON(http.StatusOK, []models.UserOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), jsonData)
	logger.LogInfo(reqId, reqUrl, "owners retrieved successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (o *OwnerController) AddOwner(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	var owner models.OwnerReg
	logger.LogDebug(reqId, reqUrl, "parsing request body")
	err := c.ShouldBindJSON(&owner)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "creating owner")
	createdOwner, err := o.CreateOwner(c.Request.Context(), userAuth.ID, &owner)
	if err != nil {
		if err != database.ErrInternal {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in adding owner :%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in creating owner :%v", err), http.StatusInternalServerError)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "owner added successfully", http.StatusOK)
	c.JSON(http.StatusOK, createdOwner)
}

func (o *OwnerController) EditOwner(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	logger.LogDebug(reqId, reqUrl, "getting owner id from url and parsing request body")
	ownerID := c.Param("ownerID")
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	var owner models.UserOutput
	owner.ID = ownerID
	err := c.ShouldBindJSON(&owner)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if userAuth.Role == middleware.Admin {
		logger.LogDebug(reqId, reqUrl, "checking owner creator")
		err = o.CheckOwnerCreator(c.Request.Context(), userAuth.ID, owner.ID)
		if err != nil {
			if err != database.ErrInternal {
				logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid creator:%v", err), http.StatusUnauthorized)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid creator:%v", err), http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
	}
	logger.LogDebug(reqId, reqUrl, "updating owner")
	ownerString, err := o.UpdateOwner(c.Request.Context(), &owner)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in updating owner:%v", err), http.StatusBadRequest)
		if err != database.ErrInternal {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in updating owner:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	jsonData := &models.UserOutput{}
	_ = json.Unmarshal([]byte(ownerString), jsonData)
	logger.LogInfo(reqId, reqUrl, "owner updated successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (o *OwnerController) DeleteOwners(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	logger.LogDebug(reqId, reqUrl, "getting request query parameters")
	multipleIdString := c.Request.URL.Query().Get("id")
	if multipleIdString == "" {
		logger.LogError(reqId, reqUrl, "empty query parameter for id", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty id parameter",
		})
		return
	}
	idArr := strings.Split(multipleIdString, ",")
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	//var ownerID struct {
	//	IDArr []string	`json:"idArr" binding:"required"`
	//}
	//err:=c.ShouldBindJSON(&ownerID)
	//if err!=nil {
	//	fmt.Print(err)
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": err.Error(),
	//	})
	//	return
	//}
	logger.LogDebug(reqId, reqUrl, "deleting owners")
	err := o.RemoveOwners(c.Request.Context(), userAuth, idArr...)
	if err != nil {
		if err != database.ErrInternal {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in deleting owner:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "owner does not exist please refresh",
			})
			return
		}
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in deleting owner:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "owner deleted successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg": "owner deleted successfully",
	})
}
