package main

import (
	"fmt"
	"tinyquant/src/config"
	"tinyquant/src/logger"

	"github.com/spf13/viper"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	fmt.Println(viper.GetString("system.ApiKey"))
	logger.Logger.Info("start server")
}
