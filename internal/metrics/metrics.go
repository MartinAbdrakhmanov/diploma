package metrics

import (
	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	functionInvocations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "faas_function_invocations_total",
			Help: "Total number of function calls",
		},
		[]string{"function_id", "function_name", "user_id", "runtime", "status"},
	)
	functionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "faas_function_duration_ms",
			Help:    "Duration of function execution in milliseconds",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000}, // TODO change
		},
		[]string{"function_id", "function_name", "user_id", "runtime"},
	)
)

func Init() {
	prometheus.MustRegister(functionDuration, functionInvocations)
}

func FunctionInvocationInc(fn ds.Function, status string) {
	functionInvocations.WithLabelValues(
		fn.ID,
		fn.Name,
		fn.UserId,
		fn.Runtime,
		status,
	).Inc()
}

func FunctionDurationObserve(fn ds.Function, duration float64) {
	functionDuration.WithLabelValues(
		fn.ID,
		fn.Name,
		fn.UserId,
		fn.Runtime,
	).Observe(duration)
}
