package Prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusEngine struct {
	CRM_Counter_Prometheus_JSON prometheus.Counter
	CRM_Counter_Prometheus_XML  prometheus.Counter
	CRM_Counter_Gauge           prometheus.Gauge
}

func (PrometheusEngine *PrometheusEngine) InitPrometheus() {
	PrometheusEngine.CRM_Counter_Prometheus_JSON = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "CRM_Counter_JSON",
		})
	prometheus.MustRegister(PrometheusEngine.CRM_Counter_Prometheus_JSON)

	PrometheusEngine.CRM_Counter_Prometheus_XML = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "CRM_Counter_XML",
		})
	prometheus.MustRegister(PrometheusEngine.CRM_Counter_Prometheus_XML)

	PrometheusEngine.CRM_Counter_Gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "CRM_Gauge",
		})
	prometheus.MustRegister(PrometheusEngine.CRM_Counter_Gauge)
}
