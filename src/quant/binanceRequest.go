package quant

import (
	"context"
	"encoding/json"
	"net/http"
	"tinyquant/src/logger"
	"tinyquant/src/mod"
	"tinyquant/src/util"

	"go.uber.org/zap"
)

type DepthMessage struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

/*
	Get Depth Message
	symbol(必需) : 品种
	limit :  默认 100; 最大 1000. 可选值:[5, 10, 20, 50, 100, 500, 1000, 5000]
*/
func (b *Binance) GetDepthMessage(ctx context.Context, symbol string, limit int32) (*DepthMessage, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    util.DepthURL,
	}
	r.SetParam(util.SymbolKey, symbol)
	if limit != 0 {
		r.SetParam(util.LimitKey, limit)
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

type LatestTrades struct {
	ID           int64  `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	Time         int64  `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

type LatestTradesList []*LatestTrades

/*
	Get latest Trade
	symbol(必需) : 品种
	limit :  默认 500; 最大 1000.
*/
func (b *Binance) GetLatestTrade(ctx context.Context, symbol string, limit int32) (*LatestTradesList, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    util.LatestTrades,
	}
	r.SetParam(util.SymbolKey, symbol)
	if limit != 0 {
		r.SetParam(util.LimitKey, limit)
	}

	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Get Latest trade Failed", zap.Error(err))
		return nil, err
	}
	p := new(LatestTradesList)
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Logger.Error("Binance Service Get Latest trade json Unmarshal Failed", zap.Error(err))
		return nil, err
	}
	return p, nil
}

type HistoryTrades struct {
	ID           int64  `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	QuoteQty     string `json:"quoteQty"`
	Time         int64  `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

type HistoryTradesList []*HistoryTrades

/*
	Get Hostory Trades
	symbol(必需) : 品种
	limit :  默认 100; 最大 1000. 可选值:[5, 10, 20, 50, 100, 500, 1000, 5000]
	fromId : 从哪一条成交id开始返回. 缺省返回最近的成交记录。
*/
func (b *Binance) GetHostoryTrades(ctx context.Context, symbol string, limit int32, fromID int64) (*HistoryTradesList, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    util.DepthURL,
	}
	r.SetParam(util.SymbolKey, symbol)
	if limit != 0 {
		r.SetParam(util.LimitKey, limit)
	}
	if fromID != 0 {
		r.SetParam(util.FromIDKey, fromID)
	}

	r.Header = http.Header{}
	r.Header.Set("X-MBX-APIKEY", util.ApiKey)

	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Get Hostory Trades Failed", zap.Error(err))
		return nil, err
	}
	p := new(HistoryTradesList)
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Logger.Error("Binance Service Get Hostory Trades json Unmarshal Failed", zap.Error(err))
		return nil, err
	}
	return p, nil
}

/*
	近期成交（归集） 归集交易与逐笔交易的区别在于，同一价格、同一方向、同一时间的trade会被聚合为一条
	symbol	STRING		YES
	fromId	LONG		NO	从包含fromId的成交id开始返回结果
	startTime	LONG	NO	从该时刻之后的成交记录开始返回结果
	endTime	LONG		NO	返回该时刻为止的成交记录
	limit	INT			NO	默认 500; 最大 1000.
*/
