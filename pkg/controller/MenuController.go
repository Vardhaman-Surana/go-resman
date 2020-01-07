package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"strconv"
	"strings"
)

type MenuController struct {
	database.Database
}

func NewMenuController(db database.Database) *MenuController {
	menuController := new(MenuController)
	menuController.Database = db
	return menuController
}

func (m *MenuController) GetMenu(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "retrieving restaurant id from url and parsing the body")
	res, _ := c.Get("restaurantID")
	resID := res.(int)
	jsonData := &[]models.DishOutput{}
	var stringData string
	logger.LogDebug(reqId, reqUrl, "retrieving dishes from db")
	stringData, err := m.ShowMenu(c.Request.Context(), resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in getting restaurant dishes:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	if stringData == "" {
		logger.LogInfo(reqId, reqUrl, "no dishes for requested restaurant", http.StatusOK)
		c.JSON(http.StatusOK, []models.DishOutput{})
		return
	}
	_ = json.Unmarshal([]byte(stringData), jsonData)
	logger.LogInfo(reqId, reqUrl, "dishes retrieved successfully", http.StatusOK)
	c.JSON(http.StatusOK, jsonData)
}

func (m *MenuController) AddDishes(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "retrieving restaurant id from url and parsing the body")
	res, _ := c.Get("restaurantID")
	resID := res.(int)
	var dish models.Dish
	err := c.ShouldBindJSON(&dish)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "adding dish to the restaurant")
	dishAdded, err := m.InsertDishes(c.Request.Context(), dish, resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in adding the dish:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "dish added successfully", http.StatusOK)
	c.JSON(http.StatusOK, dishAdded)
	return
}

func (m *MenuController) EditDish(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "retrieving restaurant id from url and parsing the body")
	var dish models.DishOutput
	res, _ := c.Get("restaurantID")
	resID := res.(int)
	dishValue := c.Param("dishID")
	dishID, _ := strconv.Atoi(dishValue)
	dish.ID = dishID
	err := c.ShouldBindJSON(&dish)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in parsing request body:%v", err), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "checking that request dish update exist in the resturant")
	err = m.CheckRestaurantDish(c.Request.Context(), resID, dish.ID)
	if err != nil {
		if err != database.ErrInternal {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("requested dish does not exist:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in checking request dish:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}
	logger.LogDebug(reqId, reqUrl, "updating the dish")
	updatedDish, err := m.UpdateDish(c.Request.Context(), &dish)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in updating the dish:%v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}
	logger.LogInfo(reqId, reqUrl, "dish updated successfully", http.StatusOK)
	c.JSON(http.StatusOK, updatedDish)
}

func (m *MenuController) DeleteDishes(c *gin.Context) {
	reqIdVal := c.Request.Context().Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := c.Request.Context().Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "retrieving restaurant id and query parameters from url")
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
			logger.LogError(reqId, reqUrl, fmt.Sprintf("invalid dish id :%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid dish id",
			})
			return
		}
		idArrInt = append(idArrInt, idInt)
	}

	//var dishID struct {
	//	IDArr []int	`json:"idArr" binding:"required"`
	//}
	//err:=c.ShouldBindJSON(&dishID)
	//if err!=nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": err.Error(),
	//	})
	//	return
	//}
	logger.LogDebug(reqId, reqUrl, "deleting dishes...")
	err := m.RemoveDishes(c.Request.Context(), idArrInt...)
	if err != nil {
		if err != database.ErrInternal {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("can not delete dishes:%v", err), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	logger.LogInfo(reqId, reqUrl, "dish deleted successfully", http.StatusOK)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Dishes deleted successfully",
	})
}
