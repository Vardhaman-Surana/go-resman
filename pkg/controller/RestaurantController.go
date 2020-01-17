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
	"strconv"
	"strings"
)

type RestaurantController struct {
	database.Database
}

func NewRestaurantController(db database.Database) *RestaurantController {
	resController := new(RestaurantController)
	resController.Database = db
	return resController
}

func (r *RestaurantController) GetNearBy(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	logger.LogDebug(reqId, reqUrl, "parsing request body")
	var location models.Location
	err := c.ShouldBindJSON(&location)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var jsonData []models.RestaurantOutput
	logger.LogDebug(reqId, reqUrl, "getting nearby restaurants")
	stringData, err := r.ShowNearBy(c.Request.Context(), &location)
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "no restaurants to show", http.StatusOK)
		c.JSON(http.StatusOK, []models.RestaurantOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), &jsonData)
	logger.LogInfo(reqId, reqUrl, "retrieved near by restaurants successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (r *RestaurantController) GetRestaurants(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	jsonData := &[]models.RestaurantOutput{}
	var stringData string
	var err error
	logger.LogDebug(reqId, reqUrl, "retriving restaurants from db")
	stringData, err = r.ShowRestaurants(c.Request.Context(), userAuth)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in retreving restaurants:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "no restaurants to show", http.StatusOK)
		c.JSON(http.StatusOK, []models.RestaurantOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), jsonData)
	logger.LogInfo(reqId, reqUrl, "restaurants retrieved successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (r *RestaurantController) AddRestaurant(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	var restaurant models.Restaurant
	logger.LogDebug(reqId, reqUrl, "parsing request body")
	err := c.ShouldBindJSON(&restaurant)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	restaurant.CreatorID = userAuth.ID
	logger.LogDebug(reqId, reqUrl, "adding restaurant")
	restaurantAdded, err := r.InsertRestaurant(c.Request.Context(), &restaurant)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in adding restaurant :%v", err), http.StatusInternalServerError)
		c.Status(http.StatusInternalServerError)
		return
	}
	logger.LogInfo(reqId, reqUrl, "restaurant added successfully", http.StatusOK)
	c.JSON(http.StatusOK, restaurantAdded)
}

func (r *RestaurantController) EditRestaurant(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	logger.LogDebug(reqId, reqUrl, "verifying user role")
	if userAuth.Role != middleware.Admin && userAuth.Role != middleware.SuperAdmin {
		logger.LogError(reqId, reqUrl, "invalid user role", http.StatusUnauthorized)
		c.Status(http.StatusUnauthorized)
		return
	}
	logger.LogDebug(reqId, reqUrl, "getting restaurant id and parsing request body")
	res, _ := c.Get("restaurantID")
	resID := res.(int)
	var restaurant models.RestaurantOutput
	restaurant.ID = resID
	err := c.ShouldBindJSON(&restaurant)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "updating restaurant")
	restaurantUpdated, err := r.UpdateRestaurant(c.Request.Context(), &restaurant)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in updating restaurant:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "restaurant updated successfully", http.StatusOK)
	c.JSON(http.StatusOK, restaurantUpdated)
}

func (r *RestaurantController) DeleteRestaurants(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	logger.LogDebug(reqId, reqUrl, "getting query parameter from request")
	multipleIdString := c.Request.URL.Query().Get("id")
	if multipleIdString == "" {
		logger.LogError(reqId, reqUrl, "empty query parameter", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty id parameter",
		})
		return
	}
	idArrString := strings.Split(multipleIdString, ",")
	var idArrInt []int
	for _, id := range idArrString {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid restaurant id:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid restaurant id",
			})
			return
		}
		idArrInt = append(idArrInt, idInt)
	}
	//var resID struct {
	//	IDArr []int	`json:"idArr" binding:"required"`
	//}
	//err:=c.ShouldBindJSON(&resID)
	//if err!=nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": err.Error(),
	//	})
	//	return
	//}
	logger.LogDebug(reqId, reqUrl, "deleting restaurants")
	err := r.RemoveRestaurants(c.Request.Context(), userAuth, idArrInt...)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in deleting restaurants:%v", err), http.StatusBadRequest)
		if err != database.ErrInternal {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	logger.LogInfo(reqId, reqUrl, "restaurant deleted successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Restaurants deleted Successfully",
	})
}
func (r *RestaurantController) GetOwnerRestaurants(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	ownerID := c.Param("ownerID")
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	var err error
	if userAuth.Role == middleware.Admin {
		logger.LogDebug(reqId, reqUrl, "checking owner creator")
		err = r.CheckOwnerCreator(c.Request.Context(), userAuth.ID, ownerID)
		if err != nil {
			if err != database.ErrInternal {
				logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid creator:%v", err), http.StatusUnauthorized)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in checking owner creator:%v", err), http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
	}
	jsonData := &[]models.RestaurantOutput{}
	ownerAuth := models.UserAuth{
		ID:   ownerID,
		Role: middleware.Owner,
	}
	var stringData string
	logger.LogDebug(reqId, reqUrl, "retrieving owner restaurants")
	stringData, err = r.ShowRestaurants(c.Request.Context(), &ownerAuth)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in getting owner restaurants:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "no owner restaurants to show", http.StatusOK)
		c.JSON(http.StatusOK, []models.RestaurantOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), jsonData)
	logger.LogInfo(reqId, reqUrl, "owner restaurants retrieved successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (r *RestaurantController) GetAvailableRestaurants(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	jsonData := &[]models.RestaurantOutput{}
	logger.LogDebug(reqId, reqUrl, "retrieving available restaurants")
	stringData, err := r.ShowAvailableRestaurants(c.Request.Context(), userAuth)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in getting available restaurants:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "no available restaurants to show", http.StatusOK)
		c.JSON(http.StatusOK, []models.RestaurantOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), jsonData)
	logger.LogInfo(reqId, reqUrl, "available restaurants retrieved successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (r *RestaurantController) AddOwnerForRestaurants(c *gin.Context) {
	 reqId,reqUrl := logger.GetRequestFieldsFromContext(c.Request.Context())

	ownerID := c.Param("ownerID")
	value, _ := c.Get("userAuth")
	userAuth := value.(*models.UserAuth)
	var err error
	if userAuth.Role == middleware.Admin {
		logger.LogDebug(reqId, reqUrl, "checking owner creator")
		err = r.CheckOwnerCreator(c.Request.Context(), userAuth.ID, ownerID)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid creator:%v", err), http.StatusUnauthorized)
			if err != database.ErrInternal {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in checking owner creator:%v", err), http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
	}
	var resID struct {
		Assign   []int `json:"assign"`
		DeAssign []int `json:"deAssign"`
	}
	logger.LogDebug(reqId, reqUrl, "getting restaurant id and parsing request body")
	err = c.ShouldBindJSON(&resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var thrownError string
	logger.LogDebug(reqId, reqUrl, "adding the restaurants to owner")
	err = r.InsertOwnerForRestaurants(c.Request.Context(), userAuth, ownerID, resID.Assign...)
	if err != nil {
		if err != database.ErrInternal {
			thrownError = err.Error()
		} else {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in adding owner for restaurants:%v", err), http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	logger.LogDebug(reqId, reqUrl, "removing owner restaurants")
	err = r.RemoveOwnerForRestaurants(c.Request.Context(), userAuth, ownerID, resID.DeAssign...)
	if err != nil {
		if err != database.ErrInternal {
			thrownError = thrownError + err.Error()
		} else {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in removing owner for restaurants:%v", err), http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	if thrownError != "" {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in updating the list:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": thrownError,
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "list updated successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg": "List Updated Successfully",
	})
}
