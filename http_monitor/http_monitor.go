package http_monitor

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto" // auto register metrics with init
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var (
	totalRequests  *prometheus.CounterVec
	responseStatus *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
	namespace      = "namespace"
	subsystem      = "subsystem"
	reqLabels      = []string{"status", "endpoint", "method"}
	one            sync.Once
)

func getCounterVecOpt(name, help string) prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
}
func newMetric() {
	totalRequests = prometheus.NewCounterVec(
		getCounterVecOpt("requests_count", "Number of get requests."),
		reqLabels,
	)

	responseStatus = prometheus.NewCounterVec(
		getCounterVecOpt("response_status", "Status of HTTP response."),
		reqLabels,
	)

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "response_time_seconds",
		Help:      "Duration of HTTP requests.",
	}, reqLabels)
}
func registerMetric() {
	list := []prometheus.Collector{totalRequests /*responseStatus,*/, httpDuration}
	for _, c := range list {
		if err := prometheus.Register(c); err != nil {
			panic(fmt.Sprintf("err:%s", err))
		}
	}
}

// Init Namespace: should be the project name
// SubSystem: should be the server name
// using alphabet and _ for string name
func Init(Namespace string, SubSystem string) func(next http.Handler) http.Handler {
	one.Do(func() {
		namespace, subsystem = Namespace, SubSystem
		newMetric()
		registerMetric()
	})
	return PrometheusMiddleware
}
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode
		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequests.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	})
}
