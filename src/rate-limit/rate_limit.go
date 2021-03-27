package main

import (
	"container/heap"
	"forward-service/domain"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"time"
)

var limiter = rate.NewLimiter(0.1, 1)

func limit(next http.Handler, pq domain.PriorityQueue) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Direct") == "true" {
			log.Print("direct")
			for limiter.Allow() == false {
				time.Sleep(10 * time.Millisecond)
			}
			w.Write([]byte("direct"))
			heap.Push(&pq, &domain.RequestHandler{
				Handler: next,
				Writer: w,
				Request: r,
				Priority:   2,
			})
		} else {
			log.Print("crawl")
			for limiter.Allow() == false {
				time.Sleep(10 * time.Millisecond)
			}
			w.Write([]byte("crawl"))
			heap.Push(&pq, &domain.RequestHandler{
				Handler: next,
				Writer: w,
				Request: r,
				Priority:   1,
			})
		}
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("exec")
		w.Write([]byte("hello"))
		w.Write([]byte(time.Now().String()))
	})

	pq := make(domain.PriorityQueue, 0)
	heap.Init(&pq)

	go func() {
		for {
			if pq.Len() > 0 {
				h := heap.Pop(&pq).(*domain.RequestHandler)
				h.Handler.ServeHTTP(h.Writer, h.Request)
			}
		}
	}()


	log.Println("Listening on :4000...")
	http.ListenAndServe(":4000", limit(mux, pq))
}
