package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	. "tinyquant/src/logger"
	"tinyquant/src/util"

	"go.uber.org/zap"
)

type BinanceWs struct {
	baseURL        string
	proxyUrl       string
	tickerCallback func(*Ticker)
	depthCallback  func(*Depth)
	tradeCallback  func(*Trade)
	klineCallback  func(*Kline, int)
	wsConns        []*util.WsConn
}

func NewBinanceWS(baseURL, ProxyURL string) *BinanceWs {
	return &BinanceWs{
		baseURL:  baseURL,
		proxyUrl: ProxyURL,
	}
}

type Ticker struct {
	symbol string  `json:"omitempty"`
	Last   float64 `json:"last,string"`
	Buy    float64 `json:"buy,string"`
	Sell   float64 `json:"sell,string"`
	High   float64 `json:"high,string"`
	Low    float64 `json:"low,string"`
	Vol    float64 `json:"vol,string"`
	Date   uint64  `json:"date"` // 单位:ms
}

type Trade struct {
	Tid    int64     `json:"tid"`
	Type   TradeSide `json:"type"`
	Amount float64   `json:"amount,string"`
	Price  float64   `json:"price,string"`
	Date   int64     `json:"date_ms"`
	symbol string    `json:"omitempty"`
}

type Depth struct {
	//ContractType string //for future
	symbol  string
	UTime   time.Time
	AskList DepthRecords // Descending order
	BidList DepthRecords // Descending order
}

type DepthRecord struct {
	Price  float64
	Amount float64
}

type DepthRecords []DepthRecord

type Kline struct {
	symbol    string
	Timestamp int64
	Open      float64
	Close     float64
	High      float64
	Low       float64
	Vol       float64
}

var _INERNAL_KLINE_PERIOD_REVERTER = map[string]int{
	"1m":  KLINE_PERIOD_1MIN,
	"3m":  KLINE_PERIOD_3MIN,
	"5m":  KLINE_PERIOD_5MIN,
	"15m": KLINE_PERIOD_15MIN,
	"30m": KLINE_PERIOD_30MIN,
	"1h":  KLINE_PERIOD_60MIN,
	"2h":  KLINE_PERIOD_2H,
	"4h":  KLINE_PERIOD_4H,
	"6h":  KLINE_PERIOD_6H,
	"8h":  KLINE_PERIOD_8H,
	"12h": KLINE_PERIOD_12H,
	"1d":  KLINE_PERIOD_1DAY,
	"3d":  KLINE_PERIOD_3DAY,
	"1w":  KLINE_PERIOD_1WEEK,
	"1M":  KLINE_PERIOD_1MONTH,
}

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:   "1m",
	KLINE_PERIOD_3MIN:   "3m",
	KLINE_PERIOD_5MIN:   "5m",
	KLINE_PERIOD_15MIN:  "15m",
	KLINE_PERIOD_30MIN:  "30m",
	KLINE_PERIOD_60MIN:  "1h",
	KLINE_PERIOD_1H:     "1h",
	KLINE_PERIOD_2H:     "2h",
	KLINE_PERIOD_4H:     "4h",
	KLINE_PERIOD_6H:     "6h",
	KLINE_PERIOD_8H:     "8h",
	KLINE_PERIOD_12H:    "12h",
	KLINE_PERIOD_1DAY:   "1d",
	KLINE_PERIOD_3DAY:   "3d",
	KLINE_PERIOD_1WEEK:  "1w",
	KLINE_PERIOD_1MONTH: "1M",
}

/*
	订阅深度信息
*/
func (bw *BinanceWs) SubscribeDepth(symbol string, size int) error {

	endpoint := fmt.Sprintf("%s/%s@depth%d@100ms", bw.baseURL, symbol, size)

	handle := func(msg []byte) error {
		rawDepth := struct {
			LastUpdateID int64           `json:"lastUpdateId"`
			Bids         [][]interface{} `json:"bids"`
			Asks         [][]interface{} `json:"asks"`
		}{}

		err := json.Unmarshal(msg, &rawDepth)
		if err != nil {
			Logger.Error("json unmarshal error for ", zap.Error(err))
			return err
		}
		depth := bw.parseDepthData(rawDepth.Bids, rawDepth.Asks)
		depth.symbol = symbol
		depth.UTime = time.Now()
		bw.depthCallback(depth)
		return nil
	}
	err := util.NewWsConn(endpoint, util.ProxyURL, handle).NewWebsocket()
	if err != nil {
		Logger.Error("[ws] SubscribeDepth failed ", zap.Error(err))
	}

	return nil
}

func (bw *BinanceWs) parseDepthData(bids, asks [][]interface{}) *Depth {
	depth := new(Depth)
	for _, v := range bids {
		depth.BidList = append(depth.BidList, DepthRecord{util.ToFloat64(v[0]), util.ToFloat64(v[1])})
	}

	for _, v := range asks {
		depth.AskList = append(depth.AskList, DepthRecord{util.ToFloat64(v[0]), util.ToFloat64(v[1])})
	}
	return depth
}

func (bw *BinanceWs) SubscribeKline(symbol string, period int) error {
	periodS, isOk := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if isOk != true {
		periodS = "M1"
	}
	endpoint := fmt.Sprintf("%s/%s@kline_%s", bw.baseURL, symbol, periodS)

	handle := func(msg []byte) error {
		datamap := make(map[string]interface{})
		err := json.Unmarshal(msg, &datamap)
		if err != nil {
			fmt.Println("json unmarshal error for ", string(msg))
			return err
		}

		msgType, isOk := datamap["e"].(string)
		if !isOk {
			return errors.New("no message type")
		}

		switch msgType {
		case "kline":
			k := datamap["k"].(map[string]interface{})
			period := _INERNAL_KLINE_PERIOD_REVERTER[k["i"].(string)]
			kline := bw.parseKlineData(k)
			kline.symbol = symbol
			bw.klineCallback(kline, period)
			return nil
		default:
			return errors.New("unknown message " + msgType)
		}
	}
	err := util.NewWsConn(endpoint, util.ProxyURL, handle).NewWebsocket()
	if err != nil {
		Logger.Error("[ws] SubscribeDepth failed ", zap.Error(err))
	}
	return nil
}

func (bnWs *BinanceWs) parseKlineData(k map[string]interface{}) *Kline {
	kline := &Kline{
		Timestamp: int64(util.ToInt(k["t"])) / 1000,
		Open:      util.ToFloat64(k["o"]),
		Close:     util.ToFloat64(k["c"]),
		High:      util.ToFloat64(k["h"]),
		Low:       util.ToFloat64(k["l"]),
		Vol:       util.ToFloat64(k["v"]),
	}
	return kline
}

func (bw *BinanceWs) parseTickerData(tickmap map[string]interface{}) *Ticker {
	t := new(Ticker)
	t.Date = util.ToUint64(tickmap["E"])
	t.Last = util.ToFloat64(tickmap["c"])
	t.Vol = util.ToFloat64(tickmap["v"])
	t.Low = util.ToFloat64(tickmap["l"])
	t.High = util.ToFloat64(tickmap["h"])
	t.Buy = util.ToFloat64(tickmap["b"])
	t.Sell = util.ToFloat64(tickmap["a"])

	return t
}
