package http_monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	// "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// type Func func(w http.ResponseWriter, r *http.Request)

// func (f Func) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	f(w, r)
// }

// go test -run=TestHttpMonitor
// curl  127.0.0.1:9001/hello
func TestHttpMonitor(t *testing.T) {
	router := mux.NewRouter()
	middleware := Init("chat_infra", "router")
	Init("chat_infra", "router")
	router.Use(middleware)
	// Prometheus endpoint register prometheus handler
	router.Path("/prometheus").Handler(promhttp.Handler())

	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world http2 " + r.URL.Path))

	}
	// Serving static files
	router.PathPrefix("/hello").Handler(http.HandlerFunc(f))

	fmt.Println("Serving requests on port 9000")
	err := http.ListenAndServe(":9000", router)
	log.Fatal(err)
}
func add(i, j int) int {
	return i + j
}
func fib(n int) int {
	if n < 0 {
		return 0
	}
	if n <= 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

type Req struct {
	N int `json:"num"`
}
type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	N    int    `json:"num"`
}

func CommonWrite(w http.ResponseWriter, num, code int, msg string) {
	resp := Resp{
		Code: code,
		Msg:  msg,
		N:    num,
	}
	bye, _ := json.Marshal(resp)
	w.Write(bye)
}

// go test -run=TestGolangHttpMonitor
// curl  127.0.0.1:9001/world -d '{"num":10}'
func TestGolangHttpMonitor(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		res := 0
		bye, err := io.ReadAll(r.Body)
		if err != nil {
			CommonWrite(w, res, -1, fmt.Sprintf("err:%s", err))
			return
		}
		req := Req{}
		if err := json.Unmarshal(bye, &req); err != nil {
			CommonWrite(w, res, -2, fmt.Sprintf("err:%s", err))
			return
		}
		res = fib(req.N)
		log.Printf("N:%d res:%d\n", req.N, res)
		CommonWrite(w, res, 0, "success")
	}
	// f is type of  HandlerFunc
	middleware := Init("chat_infra", "router")
	http.HandleFunc("/world", WrapFunc(middleware, f))
	http.HandleFunc("/prometheus", GetPrometheusHandler(middleware))
	// http.Handle("/prometheus", middleware(f))

	fmt.Println("listen:9001")
	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// curl -X POST  127.0.0.1:9001/word  '{"num":1}'
