package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"strconv"
)

func ValidateRestaurantAndCreator(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId, _ := c.Get("reqId")
		reqUrl := c.Request.URL.String()
		value, _ := c.Get("userAuth")
		userAuth := value.(*models.UserAuth)
		res := c.Param("resID")
		resID, _ := strconv.Atoi(res)
		if userAuth.Role == Admin {
			logger.LogDebug(reqId.(string), reqUrl, "checking for creator")
			err := db.CheckRestaurantCreator(c.Request.Context(), userAuth.ID, resID)
			if err != nil {
				if err != database.ErrInternal {
					logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("invalid creator: %v", err), http.StatusUnauthorized)
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": err.Error(),
					})
					c.Abort()
					return
				}
				logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("internal server error: %v", err), http.StatusUnauthorized)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				c.Abort()
				return
			}
		} else if userAuth.Role == Owner {
			logger.LogDebug(reqId.(string), reqUrl, "checking for owner")
			err := db.CheckRestaurantOwner(c.Request.Context(), userAuth.ID, resID)
			if err != nil {
				if err != database.ErrInternal {
					logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("invalid owner: %v", err), http.StatusUnauthorized)
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": err.Error(),
					})
					c.Abort()
					return
				}
				logger.LogError(reqId.(string), reqUrl, fmt.Sprintf("internal server error: %v", err), http.StatusUnauthorized)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				c.Abort()
				return
			}
		}
		c.Set("restaurantID", resID)
		c.Next()
	}
}
