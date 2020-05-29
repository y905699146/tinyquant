package main

import (
	"context"
	"fmt"
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
	fmt.Println(util.ApiKey)
	fmt.Println(util.SecretKey)
	fmt.Println(util.BaseURL)
	binance := binance.NewBinance(util.ApiKey, util.SecretKey, util.BaseURL)
	binance.GetServiceTime(context.Background())

	binance.GetDepthMessage(context.Background(), "LTCBTC", 10)

	binance.GetAccountInfo(context.Background())
}
