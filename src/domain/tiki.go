package domain

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const (
	TikiUrlGetShopProductsByName = "https://api.tiki.vn/v2/seller/stores/%v/products?page=%v"
	TikiUrlGetProductDetail = "https://tiki.vn/api/v2/products/%v?platform=web&spid=%v&include=tag,images,stock_item,variants,product_links,discount_tag,ranks,breadcrumbs,top_features,cta_desktop"
	TikiUrlSearchProducts = "https://tiki.vn/api/v2/products"
)

type TikiForwarder struct {
	Caller Caller
	Logger Logger
}

func NewTikiForwarder(caller Caller, logger Logger) TikiForwarder{
	return TikiForwarder{Caller: caller, Logger: logger}
}

func (h TikiForwarder) GetShopProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	page := vars["page"]

	url := fmt.Sprintf(TikiUrlGetShopProductsByName, username, page)
	h.Logger.Info(url)
	h.Caller.Forward(url, w)
}

func (h TikiForwarder) GetProductDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId := vars["product_id"]
	spid := vars["spid"]

	url := fmt.Sprintf(TikiUrlGetProductDetail, productId, spid)
	h.Logger.Info(url)
	h.Caller.Forward(url, w)
}

func (h TikiForwarder) SearchProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyword := vars["keyword"]
	page := vars["page"]

	q := r.URL.Query()
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = Limit
	}

	url := fmt.Sprintf("%v?limit=%v&q=%v&page=%v", TikiUrlSearchProducts, limit, keyword, page)
	h.Logger.Info(url)
	h.Caller.Forward(url, w)
}
