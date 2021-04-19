package domain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	Direct = "direct"
	Proxy  = "proxy"
)

type Caller struct {
	Logger    Logger
	ProxyPool *ProxyPool
	Kind      string
	client    *http.Client
}

func NewCaller(logger Logger, kind string, pool *ProxyPool) Caller {
	return Caller{
		Logger:    logger,
		Kind:      kind,
		ProxyPool: pool,
		client:    &http.Client{},
	}
}

func (c *Caller) Forward(url string, w http.ResponseWriter) {
	res, err := c.Get(url, nil, nil)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if res != nil {
			statusCode = res.StatusCode
		}
		c.SendError(w, statusCode, err.Error())
		return
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for key, vals := range res.Header {
		w.Header().Set(key, strings.Join(vals, "; "))
	}
	w.Write(bytes)
}

func (c *Caller) ForwardWithHeaderParams(w http.ResponseWriter, url string, params map[string]interface{}, headers map[string]interface{}) {
	res, err := c.Get(url, params, headers)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if res != nil {
			statusCode = res.StatusCode
		}
		c.SendError(w, statusCode, err.Error())
		return
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for key, vals := range res.Header {
		w.Header().Set(key, strings.Join(vals, "; "))
	}
	w.Write(bytes)
}

func (c *Caller) Get(url string, params map[string]interface{}, headers map[string]interface{}) (res *http.Response, err error) {
	client := http.DefaultClient
	if c.Kind == Proxy {
		proxy, err := c.ProxyPool.GetOne()
		if err != nil {
			c.Logger.Error("Error: %v", err)
		} else {
			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxy),
				},
				Timeout: 15 * time.Second,
			}
		}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error("Error creating request: %v", err)
		return
	}

	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, fmt.Sprintf("%v", val))
		}
	}

	req.URL.RawQuery = q.Encode()

	if headers != nil {
		for key, val := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", val))
		}
	}

	res, err = client.Do(req)
	if err != nil {
		c.Logger.Error("Error when calling api. Url: %v, error: %v", url, err)
		return
	}
	return
}

func (c *Caller) SendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func (c *Caller) SendError(w http.ResponseWriter, statusCode int, err string) {
	c.SendResponse(w, statusCode, map[string]string{"error": err})
}

func (c *Caller) NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.SendError(w, http.StatusNotFound, "route not found or not provide necessary params")
	})
}

func (c *Caller) MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
	})
}
