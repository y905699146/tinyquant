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
	binance1.TestBinance()

}
