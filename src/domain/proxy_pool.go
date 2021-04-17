package domain

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	MaxGoroutine = 200
	ProxyUrl = "http://api.insight.smartecommerce.tech/proxy"
)

type ProxyResponse struct {
	Data []string `json:"data"`
}

type ProxyPool struct {
	logger Logger
	Proxies []string
	mu sync.Mutex
	currentIndex int
}

func NewProxyPool(logger Logger) *ProxyPool{
	return &ProxyPool{logger: logger}
}

func (p *ProxyPool) UpdatePool() (err error){
	res, err := http.Get(ProxyUrl)
	if err != nil {
		p.logger.Error("%v", err)
		return
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		p.logger.Error("%v", err)
		return
	}
	var proxyResp ProxyResponse
	err = json.Unmarshal(bytes, &proxyResp)
	if err != nil {
		p.logger.Error("%v", err)
		return
	}
	if len(proxyResp.Data) == 0 {
		return errors.New("empty proxy list")
	}
	p.mu.Lock()
	p.Proxies = proxyResp.Data
	p.mu.Unlock()
	return
}

func (p *ProxyPool) FilterWorkingProxies() {
	start := time.Now()
	for len(p.Proxies) == 0 {
		p.logger.Debug("wait for proxies")
		time.Sleep(5 * time.Second)
	}
	workingProxies := make([]string, 0)
	client := http.Client{Timeout: 15 * time.Second}
	proxies := make(chan string)
	guard := make(chan struct{}, MaxGoroutine)

	var wg sync.WaitGroup
	wg.Add(len(p.Proxies))

	go func() {
		wg.Wait()
		close(proxies)
	}()

	for _, proxyStr := range p.Proxies {
		guard <- struct{}{}
		go func() {
			defer wg.Done()
			if ok := p.IsAvailable(client, proxyStr); ok {
				proxies <- proxyStr
			}
			<- guard
		}()
	}

	for proxyStr := range proxies {
		workingProxies = append(workingProxies, proxyStr)
	}

	end := time.Now()
	p.logger.Info("Filter proxy done. %v working on %v. Time: %v", len(workingProxies), len(p.Proxies), end.Sub(start))
	if len(workingProxies) > 0 {
		p.mu.Lock()
		p.Proxies = workingProxies
		p.mu.Unlock()
	}
}

func (p *ProxyPool) IsAvailable(client http.Client, proxyStr string) (ok bool){
	proxy, err := url.Parse(proxyStr)
	if err != nil {
		p.logger.Error("Parse proxy fail: %v. Proxy: %v", err, proxyStr)
		return false
	}

	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	res, err := client.Get("https://shopee.vn/api/v2/category_list/get")
	if err != nil {
		p.logger.Error("Send request fail: %v. Proxy: %v", err, proxy)
		return false
	}
	if res.StatusCode != 200 {
		p.logger.Error("Request with code: %v, proxy: %v", res.StatusCode, proxyStr)
		return false
	}
	p.logger.Info("Proxy %v ok", proxyStr)
	return true
}

func (p *ProxyPool) GetOne() (proxyUrl *url.URL, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Proxies) == 0 {
		return proxyUrl, errors.New("no proxy in pool")
	}
	proxyString := p.Proxies[p.currentIndex % len(p.Proxies)]
	proxyUrl, err = url.Parse(proxyString)
	if err != nil {
		p.logger.Error("%v", err)
	}

	p.currentIndex++
	return
}

