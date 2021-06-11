package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/HuanLiu-hotstar/monitor/http_monitor"
)

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
func main() {
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
	middleware := http_monitor.Init("chat_infra", "router")
	http.HandleFunc("/world", http_monitor.WrapFunc(middleware, f))
	http.HandleFunc("/prometheus", http_monitor.GetPrometheusHandler(middleware))
	// http.Handle("/prometheus", middleware(f))

	fmt.Println("listen:9001")
	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// curl -X POST  127.0.0.1:9001/word  '{"num":1}'
