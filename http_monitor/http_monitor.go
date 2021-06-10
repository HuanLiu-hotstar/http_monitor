package http_monitor

import (
	"fmt"
	"net/http"
	"strings"
	// "strconv"
	"sync"
	"time"

	// "github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto" // auto register metrics with init
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// count,http_status,remote_ip,path,duration monitor

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
	reqLabels      = []string{"status", "endpoint", "path"}
	// http_status, remote_ip ,method ,uri,duration
	one     sync.Once
	destroy bool
	// monitor metrics list
	monitorList []prometheus.Collector
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
	monitorList = []prometheus.Collector{totalRequests, responseStatus, httpDuration}
}

func registerMetric() {
	for _, c := range monitorList {
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
		destroy = true
	})
	return PrometheusFunc(PrometheusMiddleware)
	// return PrometheusMiddleware
}
func unregisterMetrics() {
	for _, c := range monitorList {
		prometheus.Unregister(c)
	}
}

func Destory() {
	if destroy {
		unregisterMetrics()
	}
}
func GetIP(s string) string {
	if ip := strings.Split(s, ":"); len(ip) > 1 && len(ip[0]) >= 7 { // 1.1.1.1
		return ip[0]
	}
	return ""
}

type PrometheusFunc func(next http.Handler) http.Handler

func (p PrometheusFunc) WrapFunc(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p(http.Handler(http.HandlerFunc(f))).ServeHTTP(w, r)
	}
}

func WrapFunc(middle PrometheusFunc, f http.HandlerFunc) http.HandlerFunc {
	return middle.WrapFunc(f)
}

// PrometheusMiddleware return the wrap for prometheus monitor
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, start, path := GetIP(r.RemoteAddr), time.Now(), r.URL.Path

		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := fmt.Sprintf("%d", rw.statusCode)
		labels := []string{statusCode, ip, path}
		//duration
		httpDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
		// resp status
		responseStatus.WithLabelValues(labels...).Inc()
		// total request
		totalRequests.WithLabelValues(labels...).Inc()

	})
}

// UseMiddleHandlerFunc wrap for HandlerFunc with Prometheus middleware
func UseMiddleHandlerFunc(f func(w http.ResponseWriter, r *http.Request), middleware PrometheusFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware(http.Handler(http.HandlerFunc(f))).ServeHTTP(w, r)
	}
}

// GetPromotheusHandler Prometheus http.HandlerFunc
func GetPrometheusHandler(middleware PrometheusFunc) http.HandlerFunc {
	return middleware(promhttp.Handler()).ServeHTTP
}

// route := mux.CurrentRoute(r)
// path, _ := route.GetPathTemplate()
// timer := prometheus.NewTimer(httpDuration.WithLabelValues(status, ip, path))
// responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
// timer.ObserveDuration()
// "status", "endpoint", "method"
// count,http_status, remote_ip ,method ,uri,duration
