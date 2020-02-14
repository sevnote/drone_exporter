package drone_exporter

import (
	"drone_exporter/src/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var Reg = prometheus.NewPedanticRegistry()

func init() {
	Reg.MustRegister(metrics.NewMetrics("192.168.16.77"))
}
