package quant

import (
	"context"
	"encoding/json"
	"tinyquant/src/logger"
	"tinyquant/src/mod"
	"tinyquant/src/util"

	"go.uber.org/zap"
)

type Binance struct {
}

func NewBinance() *Binance {
	return &Binance{}
}

type Ping struct {
}

// ping service
func (b *Binance) Ping(ctx context.Context) (*Ping, error) {
	r := &mod.ReqParam{
		Method: "GET",
		Url:    "/api/v1/ping",
	}
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Ping Failed", zap.Error(err))
		return nil, err
	}
	p := new(Ping)
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Logger.Error("Binance Service json Unmarshal Failed", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

type SeviceTime struct {
	ServerTime int64 `json:"serverTime"`
}

func (b *Binance) GetServiceTime(ctx context.Context) (int64, error) {
	r := &mod.ReqParam{
		Method: "GET",
		Url:    "/api/v3/time",
	}
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Ping Failed", zap.Error(err))
		return 0, err
	}
	p := new(SeviceTime)
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Logger.Error("Binance Service json Unmarshal Failed", zap.Error(err))
		return 0, err
	}
	return p.ServerTime, nil
}
