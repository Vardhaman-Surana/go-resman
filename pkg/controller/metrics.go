package controller

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
)

const(
	logins         = "total_logins"
)

func init(){
}

var(
	counters = []prometheus2.CounterOpts{
		{
			Name: logins,
			Help: "total logins",
		},
	}
)
