package util

import "github.com/spf13/viper"

var (
	BaseURL   string
	ApiKey    string
	SecretKey string
	ProxyURL  string
)

func InitSystemParams() {
	BaseURL = viper.GetString("system.BaseURL")
	if BaseURL == "" {
		panic("Get Binance base url failed ")
	}
	ApiKey = viper.GetString("system.ApiKey")
	if ApiKey == "" {
		panic("Get ApiKey  failed ")
	}
	SecretKey = viper.GetString("system.SecretKey")
	if SecretKey == "" {
		panic("Get secretKey failed ")
	}
	ProxyURL = viper.GetString("system.ProxyURL")
	if ProxyURL == "" {
		panic("Get ProxyURL failed ")
	}
}

const (
	timestampKey  = "timestamp"
	signatureKey  = "signature"
	recvWindowKey = "recvWindow"
	SymbolKey     = "symbol"
	LimitKey      = "limit"
	FromIDKey     = "fromId"
)

// binance url
const (
	PingURL        = "/api/v3/ping"
	ServiceTimeURL = "/api/v3/time"
	DepthURL       = "/api/v3/depth"
	LatestTrades   = "/api/v3/trades"
	HistoryTrades  = "/api/v3/historicalTrades"
	LatestTradesA  = "/api/v3/aggTrades"
)

// 交易对
const (
	BTC_USDT = "BTCUSDT"
)
