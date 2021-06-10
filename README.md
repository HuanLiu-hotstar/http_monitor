# monitor

## http_monitor

- `request_count,http_status,remote_ip,path,duration` monitor
- totalRequest:(`requests_count`) number of get requests
- httpRespStatus:(`response_status`) Status of Http response
- httpDuration:(`response_time_seconds`) duration of Http request


## usage example

```golang
// Init func 
func Init(Namespace string, SubSystem string) func(next http.Handler) http.Handler
```

- for `mux router` from `github.com/gorilla/mux`


```go
func main(){
    router := mux.NewRouter()
     // init middleware with namespace = chat_infra ,subsystem= router 
    middleware := Init("chat_infra", "router")
    // user prometheus middleware 
    router.Use(middleware)  
    //register for prometheus data colloction
    router.Path("/prometheus").Handler(promhttp.Handler())  
    f := func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("hello world http2\n"))
    }
    // Serving for logic uri 
    router.PathPrefix("/hello").Handler(http.HandlerFunc(f))    
    fmt.Println("Serving requests on port 9000")
    err := http.ListenAndServe(":9000", router)
    log.Fatal(err)
}
```

- for `raw http` in golang

```go
func main() {
    f := func(w http.ResponseWriter, r *http.Request) {
        w.Write("hello world with prometheus")
    }
    // init middleware with namespace = chat_infra ,subsystem= router 
    middleware := Init("chat_infra", "router")


    //register with handler func for /world 
    http.HandleFunc("/world", WrapFunc(middleware, f))// wrap func f with middleware 

    //register for prometheus data colloction
    http.HandleFunc("/prometheus", GetPrometheusHandler(middleware)) // start prometheus uri 

    fmt.Println("listen:9001")
    err := http.ListenAndServe(":9001", nil)
    if err != nil {
        log.Fatal(err)
    }
}
```
