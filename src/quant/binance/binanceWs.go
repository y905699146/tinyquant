package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func (bnWs *BinanceWs) parseDepthData(bids, asks [][]interface{}) *Depth {
	depth := new(Depth)
	for _, v := range bids {
		depth.BidList = append(depth.BidList, DepthRecord{util.ToFloat64(v[0]), util.ToFloat64(v[1])})
	}

	for _, v := range asks {
		depth.AskList = append(depth.AskList, DepthRecord{util.ToFloat64(v[0]), util.ToFloat64(v[1])})
	}
	return depth
}

func (bnWs *BinanceWs) SubscribeKline(symbol string, period int) error {
	if bnWs.klineCallback == nil {
		return errors.New("place set kline callback func")
	}
	periodS, isOk := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if isOk != true {
		periodS = "M1"
	}
	endpoint := fmt.Sprintf("%s/%s@kline_%s", bnWs.baseURL, strings.ToLower(pair.ToSymbol("")), periodS)

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
			kline := bnWs.parseKlineData(k)
			kline.Pair = pair
			bnWs.klineCallback(kline, period)
			return nil
		default:
			return errors.New("unknown message " + msgType)
		}
	}
	bnWs.subscribe(endpoint, handle)
	return nil
}
