module tinyquant

go 1.12

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20200331124033-c3d80250170d

require (
	github.com/binance-exchange/go-binance v0.0.0-20180518133450-1af034307da5
	github.com/gin-gonic/gin v1.6.2
	github.com/go-kit/kit v0.8.0
	github.com/gorilla/websocket v1.4.0
	github.com/spf13/viper v1.6.2
	go.uber.org/zap v1.10.0
)
