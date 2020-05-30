package binance

import (
	"context"
	. "tinyquant/src/logger"

	"github.com/binance-exchange/go-binance"
)

var b binance.Binance

func NewBinance(accKey, secKey string) {

	hmacSigner := &binance.HmacSigner{
		Key: []byte(secKey),
	}
	ctx := context.Background()

	binanceService := binance.NewAPIService(
		"https://www.binance.com",
		accKey,
		hmacSigner,
		Logger,
		ctx,
	)
	b = binance.NewBinance(binanceService)
}
