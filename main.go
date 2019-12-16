package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type Params struct {
	RPS       int64    `json:"request_per_second"`
	TotalTime int      `json:"total_time"`
	URL       string   `json:"url"`
	Headers   []Header `json:"headers"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func test(p *Params) {
	fmt.Printf("testing [%s] for %d seconds with %d req/s\n", p.URL, p.TotalTime, p.RPS)

	tick := time.Tick(time.Duration(time.Second.Nanoseconds() / p.RPS))
	timeout := time.After(time.Duration(p.TotalTime) * time.Second)

	for {
		select {
		case <-tick:
			go func(url string, headers []Header) {
				client := http.DefaultClient

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					panic(err)
				}

				for _, h := range headers {
					req.Header.Set(h.Name, h.Value)
				}

				start := time.Now()
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				fmt.Printf("got [%s] in %dms\n", resp.Status, time.Since(start).Milliseconds())
			}(p.URL, p.Headers)
		case <-timeout:
			return
		}
	}
}

func main() {
	port := flag.Int("port", 2000, "Port to listen on")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var p Params
		err := decoder.Decode(&p)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		go test(&p)

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("listening on", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
