package domain

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	ShopeeUrlGetMainCategoryInfo = "https://shopee.vn/api/v0/search/api/categorytags"
	ShopeeUrlGetProductInfo      = "https://shopee.vn/api/v2/item/get"
	ShopeeUrlSearchProducts      = "https://shopee.vn/api/v2/search_items"
	ShopeeUrlGetShopDetail       = "https://shopee.vn/api/v4/shop/get_shop_detail"
	ShopeeUrlGetShopCategories   = "https://shopee.vn/api/v2/shop/get_categories"
	ShopeeUrlGetMalls            = "https://shopee.vn/api/v2/brand_lists/get"
	ShopeeUrlSimilarProducts = "https://shopee.vn/api/v4/recommend/recommend"
)

const (
	Limit = 50
)

type ShopeeForwarder struct {
	Caller Caller
}

func NewShopeeForwarder(caller Caller) ShopeeForwarder {
	return ShopeeForwarder{Caller: caller}
}

// category
func (h ShopeeForwarder) GetMainCatInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cateId := vars["category_id"]
	url := fmt.Sprintf("%v?main_catid=%v&page_type=search", ShopeeUrlGetMainCategoryInfo, cateId)
	h.Caller.Forward(url, w)
}

// product
func (h ShopeeForwarder) GetProductInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId := vars["product_id"]
	shopId := vars["shop_id"]

	url := fmt.Sprintf("%v?itemid=%v&shopid=%v", ShopeeUrlGetProductInfo, productId, shopId)
	h.Caller.Forward(url, w)
}

func (h ShopeeForwarder) SearchProducts(w http.ResponseWriter, r *http.Request) {
	by := "relevancy"
	limit := Limit
	order := "desc"
	pageType := "search"
	version := 2

	vars := mux.Vars(r)
	matchId := vars["category_id"]
	keyword := vars["keyword"]
	newest := vars["from"]

	h.Caller.Logger.Info("category: %v, keyword: %v, from: %v", matchId, keyword, newest)

	url := fmt.Sprintf("%v?by=%v&limit=%v&newest=%v&order=%v&page_type=%v&version=%v", ShopeeUrlSearchProducts, by, limit, newest, order, pageType, version)
	if matchId != "" {
		url = fmt.Sprintf("%v&match_id=%v", url, matchId)
	}
	if keyword != "" {
		url = fmt.Sprintf("%v&keyword=%v", url, keyword)
	}
	h.Caller.Forward(url, w)
}

// shop
func (h ShopeeForwarder) GetShopDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopId := vars["shop_id"]
	username := vars["username"]
	if shopId != "" {
		url := fmt.Sprintf("%v?shopid=%v", ShopeeUrlGetShopDetail, shopId)
		h.Caller.Forward(url, w)
	} else if username != "" {
		url := fmt.Sprintf("%v?username=%v", ShopeeUrlGetShopDetail, username)
		h.Caller.Forward(url, w)
	}
}

func (h ShopeeForwarder) GetShopCollections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopId := vars["shop_id"]
	from := vars["from"]
	url := fmt.Sprintf("%v?limit=%v&offset=%v&shopid=%v", ShopeeUrlGetShopCategories, Limit, from, shopId)
	h.Caller.Forward(url, w)
}

func (h ShopeeForwarder) GetShopProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopId := vars["shop_id"]
	from := vars["from"]
	h.Caller.ForwardWithHeaderParams(w, ShopeeUrlSearchProducts, map[string]interface{}{
		"by":        "pop",
		"limit":     Limit,
		"match_id":  shopId,
		"newest":    from,
		"order":     "desc",
		"page_type": "shop",
		"version":   2,
	}, map[string]interface{}{
		"Referer": fmt.Sprintf("https://shopee.vn/shop/%v/search", shopId),
	})
}

func (h ShopeeForwarder) GetMalls(w http.ResponseWriter, r *http.Request) {
	h.Caller.Forward(ShopeeUrlGetMalls, w)
}

func (h ShopeeForwarder) GetSimilarProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catid := vars["catid"]
	itemid := vars["itemid"]
	shopid := vars["shopid"]
	h.Caller.ForwardWithHeaderParams(w, ShopeeUrlSimilarProducts, map[string]interface{}{
		"catid": catid,
		"item_card": 2,
		"itemid": itemid,
		"limit": Limit,
		"offset": 0,
		"section": "similar_product",
		"shopid": shopid,
		"bundle": "product_detail_page",
	}, nil)
}