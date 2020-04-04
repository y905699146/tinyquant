package main

import (
	"context"
	"fmt"
	"tinyquant/src/config"
	"tinyquant/src/logger"
	"tinyquant/src/quant"
	"tinyquant/src/util"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	logger.Logger.Info("start server")

	util.InitSystemParams()

	binance := quant.NewBinance()
	t, err := binance.GetServiceTime(context.Background())
	fmt.Println(t, err)
}
