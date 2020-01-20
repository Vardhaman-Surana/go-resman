package controller

import (
	"github.com/vds/go-resman/pkg/prometheus"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
)

const(
	logins         = "total_logins"
)

func init(){
	prometheus.Global().RegisterCounterVectors(counterVectors)
}

var(
	counterVectors = []prometheus.CounterVecOpts{
		{
			Opts: prometheus2.CounterOpts{
				Name: logins,
				Help: "total logins",
			},
			Labels: []string{"role"},
		},
	}
)
