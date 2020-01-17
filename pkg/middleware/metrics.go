package middleware

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/vds/go-resman/pkg/prometheus"
)

const(
	requestDuration         = "request_duration"
)

func init(){
	prometheus.Global().RegisterHistogramVectors(histVecs)
}

var(
	histVecs = []prometheus.HistogramVecOpts{
		{
			Opts: prometheus2.HistogramOpts{
				Name: requestDuration,
				Help: "time in each request completion",
				Buckets: []float64{2,4,6,8,10,},
			},
			Labels: []string{"method","path","handler","status"},
		},
	}

)
