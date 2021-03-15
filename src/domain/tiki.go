package domain

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	TikiUrlGetShopProductsByName = "https://api.tiki.vn/v2/seller/stores/%v/products?page=%v"
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
