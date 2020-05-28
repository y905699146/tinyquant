package main

import (
	"context"
	"tinyquant/src/config"
	"tinyquant/src/logger"
	"tinyquant/src/quant/binance"
	"tinyquant/src/util"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	logger.Logger.Info("start server")

	util.InitSystemParams()

	binance := binance.NewBinance()
	binance.GetServiceTime(context.Background())

	binance.GetDepthMessage(context.Background(), "LTCBTC", 10)

}
