package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/prometheus"
	"strconv"
	"strings"
	"time"
)

func InstrumentPrometheus(pathMap *map[string]string) gin.HandlerFunc{
	return func(c *gin.Context) {
		start := time.Now()
		if c.Request.URL.String() == "/metrics" {
			c.Next()
			return
		}
		path := (*pathMap)[c.HandlerName()]
		c.Next()
		status := strconv.Itoa(c.Writer.Status())
		requestTime := float64(time.Since(start)/ time.Millisecond)
		funcName := strings.Split(c.HandlerName(), "/")
		fname := funcName[len(funcName)-1]
		prometheus.Global().GetHistogramVec(requestDuration).WithLabelValues(
				c.Request.Method, path, fname, status).Observe(requestTime)
	}
}
