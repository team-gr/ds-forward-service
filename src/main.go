package main

import (
	"forward-service/domain"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"strings"
	"time"
)

var limiter = rate.NewLimiter(1, 1)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/direct") {
			for limiter.Allow() == false {
				time.Sleep(500 * time.Millisecond)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func main()  {
	logger := domain.NewLogger()
	pool := domain.NewProxyPool(logger)

	s := gocron.NewScheduler(time.UTC)
	s.Every(60).Minutes().Do(pool.UpdatePool)
	s.Every(25).Minutes().Do(pool.FilterWorkingProxies)
	s.StartAsync()

	directCaller := domain.NewCaller(logger, domain.Direct, nil)
	proxyCaller := domain.NewCaller(logger, domain.Proxy, pool)

	shopee := domain.NewShopeeForwarder(directCaller)
	tiki := domain.NewTikiForwarder(directCaller, logger)
	lazada := domain.NewLazadaForwarder(directCaller, logger)

	shopeeProxy := domain.NewShopeeForwarder(proxyCaller)
	tikiProxy := domain.NewTikiForwarder(proxyCaller, logger)
	lazadaProxy := domain.NewLazadaForwarder(proxyCaller, logger)

	r := mux.NewRouter()
	r.NotFoundHandler = directCaller.NotFoundHandler()
	r.MethodNotAllowedHandler = directCaller.MethodNotAllowedHandler()

	direct := r.PathPrefix("/direct").Subrouter()
	RegisterRoutes(direct, shopee, tiki, lazada)

	proxy := r.PathPrefix("/proxy").Subrouter()
	RegisterRoutes(proxy, shopeeProxy, tikiProxy, lazadaProxy)

	log.Print("Listening on port 9090")
	log.Fatal(http.ListenAndServe(":9090", limit(r)))
}

func RegisterRoutes(r *mux.Router, shopee domain.ShopeeForwarder, tiki domain.TikiForwarder, lazada domain.LazadaForwarder) {
	shopeeRouter := r.PathPrefix("/shopee").Subrouter()

	// category
	shopeeRouter.HandleFunc("/category", shopee.GetMainCatInfo).Queries("category_id", "{category_id:[0-9]+}")
	// products
	shopeeRouter.HandleFunc("/product/similar", shopee.GetSimilarProducts).Queries("shopid", "{shopid:[0-9]+}", "itemid", "{itemid:[0-9]+}")
	shopeeRouter.HandleFunc("/product/detail", shopee.GetProductInfo).Queries("product_id", "{product_id:[0-9]+}", "shop_id", "{shop_id:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("category_id", "{category_id}", "keyword", "{keyword}").Queries("from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("category_id", "{category_id}").Queries("from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/product/search", shopee.SearchProducts).Queries("keyword", "{keyword}").Queries("from", "{from:[0-9]+}")
	shopeeRouter.HandleFunc("/product/hints", shopee.SearchHints).Queries("keyword", "{keyword}")
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
}
