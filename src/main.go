package main

import (
	"forward-service/domain"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"strings"
	"time"
)

var sLimiter = rate.NewLimiter(1, 1)
var tLimiter = rate.NewLimiter(1, 1)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "tiki") {
			for tLimiter.Allow() == false {
				time.Sleep(10 * time.Millisecond)
			}
		} else {
			for sLimiter.Allow() == false {
				time.Sleep(10 * time.Millisecond)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func main()  {
	logger := domain.NewLogger()
	caller := domain.NewCaller(logger)
	shopee := domain.NewShopeeForwarder(caller)
	tiki := domain.NewTikiForwarder(caller, logger)
	lazada := domain.NewLazadaForwarder(caller, logger)

	r := mux.NewRouter()
	r.NotFoundHandler = caller.NotFoundHandler()
	r.MethodNotAllowedHandler = caller.MethodNotAllowedHandler()

	shopeeRouter := r.PathPrefix("/shopee").Subrouter()

	// category
	shopeeRouter.HandleFunc("/category", shopee.GetMainCatInfo).Queries("category_id", "{category_id:[0-9]+}")
	// products
	shopeeRouter.HandleFunc("/product/detail", shopee.GetProductInfo).Queries("product_id", "{product_id:[0-9]+}", "shop_id", "{shop_id:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("category_id", "{category_id}", "keyword", "{keyword}").Queries("from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("category_id", "{category_id}").Queries("from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("keyword", "{keyword}}").Queries("from", "{from:[0-9]+}")
	// shops
	shopeeRouter.HandleFunc("/shop/detail", shopee.GetShopDetail).Queries("shop_id", "{shop_id:[0-9]+}")
	shopeeRouter.HandleFunc("/shop/detail", shopee.GetShopDetail).Queries("username", "{username}")
	shopeeRouter.HandleFunc("/shop/collections", shopee.GetShopCollections).Queries("shop_id", "{shop_id:[0-9]+}", "from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/shop/products", shopee.GetShopProducts).Queries("shop_id", "{shop_id:[0-9]+}", "from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/shop/malls", shopee.GetMalls)

	tikiRouter := r.PathPrefix("/tiki").Subrouter()
	tikiRouter.HandleFunc("/shop/products", tiki.GetShopProducts).Queries("username", "{username}", "page", "{page:[0-9]}")

	lazadaRouter := r.PathPrefix("/lazada").Subrouter()
	lazadaRouter.HandleFunc("/shop/id", lazada.GetShopId).Queries("shop_url", "{shop_url}")

	log.Print("Listening on port 9090")
	log.Fatal(http.ListenAndServe(":9090", limit(r)))
}
