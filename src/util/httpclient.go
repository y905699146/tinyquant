package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
	"tinyquant/src/logger"
	"tinyquant/src/mod"

	"go.uber.org/zap"
)

var (
	client *http.Client
)

func init() {
	if client == nil {
		client = initHttpClient()
	}
}

func initHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			MaxIdleConns:        30,               //最大空闲连接数
			MaxIdleConnsPerHost: 60,               //最大与服务器的连接数  默认是2
			IdleConnTimeout:     30 * time.Second, //空闲连接保持时间
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:7890")
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
		},
	}
	return client
}

func HttpRequest(ctx context.Context, req *mod.ReqParam) ([]byte, error) {

	urlx := fmt.Sprintf("%s%s", BaseURL, req.URL)

	queryString := req.Query.Encode()
	if queryString != "" {
		urlx = fmt.Sprintf("%s?%s", urlx, queryString)
	}
	r, err := http.NewRequest(req.Method, urlx, nil)
	r = r.WithContext(ctx)
	r.Header = req.Header
	if err != nil {
		logger.Logger.Error("http request failed ", zap.Error(err))
		return []byte{}, err
	}
	res, err := client.Do(r)
	if err != nil {
		logger.Logger.Error("http Do failed ", zap.Error(err))
		return []byte{}, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Logger.Error("io read failed ", zap.Error(err))
		return []byte{}, err
	}
	fmt.Println(string(body))
	return body, err
}
