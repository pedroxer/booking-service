package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

var rpcMetricCounter prometheus.Counter

var bookingMetricGauge *prometheus.GaugeVec

func MetricsInit() error {
	bookingMetricGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "bookings_by_type",
		Subsystem: "booking",
		Help:      "Number of bookings",
	}, []string{"service"})

	if err := prometheus.Register(bookingMetricGauge); err != nil {
		return fmt.Errorf("couldn't register rpcMetricCounter: %v", err)
	}

	rpcMetricCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "rpc_counter",
		Subsystem: "booking",
		Help:      "Number of rpc calls",
	})

	if err := prometheus.Register(rpcMetricCounter); err != nil {
		return fmt.Errorf("couldn't register rpcMetricCounter: %v", err)
	}
	return nil
}

func IncrementBookingCounter(service string) {
	bookingMetricGauge.WithLabelValues(service).Inc()
}
func RpcMetricCounterInc() {
	rpcMetricCounter.Inc()
}
