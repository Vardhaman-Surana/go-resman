package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vds/go-resman/pkg/logger"
	"net/http"
)

func GenerateRequestId(c *gin.Context) {
	reqId, err := uuid.NewUUID()
	reqUrl := c.Request.URL.String()
	logger.LogDebug("", reqUrl, "generating request id")
	if err != nil {
		logger.LogError("", reqUrl, fmt.Sprintf("can not generate request id: %v", err), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		c.Abort()
	}
	ctx := context.WithValue(c.Request.Context(), "reqId", reqId.String())
	ctx1 := context.WithValue(ctx, "reqUrl", reqUrl)
	r := c.Request.WithContext(ctx1)
	c.Request = r
	c.Set("reqId", reqId.String())
	c.Next()
}
