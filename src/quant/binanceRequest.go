package quant

import (
	"context"
	"encoding/json"
	"tinyquant/src/logger"
	"tinyquant/src/mod"
	"tinyquant/src/util"

	"go.uber.org/zap"
)

type Bid struct {
	Price    string
	Quantity string
}

type Ask struct {
	Price    string
	Quantity string
}

type DepthMessage struct {
	LastUpdateID int64 `json:"lastUpdateId"`
	Bids         []Bid `json:"bids"`
	Asks         []Ask `json:"asks"`
}

func (b *Binance) GetDepthMessage(ctx context.Context) (*DepthMessage, error) {
	r := &mod.ReqParam{
		Method: "GET",
		Url:    "/api/v3/depth",
	}
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Get Depth Failed", zap.Error(err))
		return nil, err
	}
	p := new(DepthMessage)
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Logger.Error("Binance Service Get Depth json Unmarshal Failed", zap.Error(err))
		return nil, err
	}
	return p, nil
}
