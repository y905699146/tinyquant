package main

import (
	"tinyquant/src/config"
	"tinyquant/src/logger"
	"tinyquant/src/quant/binance1"
	"tinyquant/src/util"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	logger.Logger.Info("start server")

	util.InitSystemParams()

	//binance := binance.NewBinance()
	//binance.GetServiceTime(context.Background())

	//binance.GetDepthMessage(context.Background(), "LTCBTC", 10)
	//binanceWs := binance.NewBinanceWS(util.WebSocketURL, util.ProxyURL)
	//binanceWs.SubscribeDepth("LTCBTC", 5)
	binance1.TestBinance()

}
