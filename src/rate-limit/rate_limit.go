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

func limit(next http.Handler, pq *domain.PriorityQueue) http.Handler {
	go func() {
		for {
			if pq.Len() > 0 {
				h := heap.Pop(pq).(*domain.RequestHandler)
				next.ServeHTTP(h.Writer, h.Request)
			}
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Direct") == "true" {
			w.Write([]byte("direct"))
			heap.Push(pq, &domain.RequestHandler{
				Handler: next,
				Writer: w,
				Request: r,
				Priority:   2,
			})
		} else {
			w.Write([]byte("crawl"))
			heap.Push(pq, &domain.RequestHandler{
				Handler: next,
				Writer: w,
				Request: r,
				Priority:   1,
			})
		}

		for limiter.Allow() == false {
			time.Sleep(10 * time.Millisecond)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("exec")
		log.Printf("%v", r)
		w.Write([]byte("hello"))
	})

	pq := make(domain.PriorityQueue, 0)
	heap.Init(&pq)

	log.Println("Listening on :4000...")
	http.ListenAndServe(":4000", limit(mux, &pq))
}
