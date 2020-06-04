package binance1

import (
	"context"
	"fmt"
	"os"

	"github.com/y905699146/binance"
)

func TestBinance() {
	hmacSigner := &binance.HmacSigner{
		Key: []byte(os.Getenv("BINANCE_SECRET")),
	}
	ctx := context.Background()
	// use second return value for cancelling request
	binanceService := binance.NewAPIService(
		"https://www.binance.com",
		os.Getenv("BINANCE_APIKEY"),
		hmacSigner,
		ctx,
	)
	b := binance.NewBinance(binanceService)

	kl, err := b.Klines(binance.KlinesRequest{
		Symbol:   "BNBETH",
		Interval: binance.Hour,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", kl)
}
