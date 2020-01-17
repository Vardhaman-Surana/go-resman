package prometheus

import "github.com/prometheus/client_golang/prometheus"

type CounterVecOpts struct{
	Opts prometheus.CounterOpts
	Labels []string
}

type HistogramVecOpts struct{
	Opts prometheus.HistogramOpts
	Labels []string
}


func toCounterVecMap(gaugeOpts []CounterVecOpts) map[string]*prometheus.CounterVec {
	countMap := make(map[string]*prometheus.CounterVec)
	for _, opt := range gaugeOpts {
		countMap[opt.Opts.Name] = prometheus.NewCounterVec(opt.Opts, opt.Labels)
	}
	return countMap
}

func toHistVecMap(histOpts []HistogramVecOpts) map[string]*prometheus.HistogramVec {
	histMap := make(map[string]*prometheus.HistogramVec)
	for _, opt := range histOpts {
		histMap[opt.Opts.Name] = prometheus.NewHistogramVec(opt.Opts, opt.Labels)
	}
	return histMap
}
