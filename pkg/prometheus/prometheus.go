package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
)

var(
	promMetrics *Metrics
)

func init(){
	promMetrics = New()
}

func New() *Metrics {
	return &Metrics{
		counterVecs:    make(map[string]*prometheus.CounterVec),
		histVecs: make(map[string]*prometheus.HistogramVec),
	}
}


type Metrics struct{
	counterVecs   map[string]*prometheus.CounterVec
	histVecs map[string]*prometheus.HistogramVec
}

type GlobalMetrics struct{}

func Global() *GlobalMetrics{
	return &GlobalMetrics{}
}

func NewHandler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) RegisterCounterVectors(opts []CounterVecOpts) {
	for name, count := range toCounterVecMap(opts) {
		err := prometheus.Register(count)
		if err != nil {
			log.Error("[Prometheus] ", err.Error())
		}
		m.counterVecs[name] = count
	}
}

func (m *Metrics) RegisterHistogramVectors(opts []HistogramVecOpts) {
	for name, hist := range toHistVecMap(opts) {
		err := prometheus.Register(hist)
		if err != nil {
			log.Error("[Prometheus] ", err.Error())
		}
		m.histVecs[name] = hist
	}
}

func (m *Metrics)GetHistogramVec(name string) *prometheus.HistogramVec{
	return m.histVecs[name]
}

func(*GlobalMetrics)RegisterHistogramVectors(opts []HistogramVecOpts){
	promMetrics.RegisterHistogramVectors(opts)
}

func(*GlobalMetrics)RegisterCounterVectors(opts []CounterVecOpts){
	promMetrics.RegisterCounterVectors(opts)
}
func(*GlobalMetrics)GetHistogramVec(name string) *prometheus.HistogramVec{
	return promMetrics.GetHistogramVec(name)
}

