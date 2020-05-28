package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"tinyquant/src/logger"
	"tinyquant/src/mod"
	"tinyquant/src/util"

	"go.uber.org/zap"
)

type Binance struct {
	accessKey  string
	secretKey  string
	baseUrl    string
	httpClient *http.Client
}

func NewBinance() *Binance {
	return &Binance{}
}

/*
	SHA256生成签名，SECRETKEY为密钥，body为参数
*/
func (b *Binance) ParamsSigned(postForm *url.Values) error {
	postForm.Set("recvWindow", "60000")
	tonce := strconv.FormatInt(util.GetCurrentUnixNano(), 10)[0:13]
	postForm.Set("timestamp", tonce)
	postMsg := postForm.Encode()
	mac := hmac.New(sha256.New, []byte(b.secretKey))
	_, err := mac.Write([]byte(postMsg))
	if err != nil {
		return err
	}
	sign := hex.EncodeToString(mac.Sum(nil))
	postForm.Set("signature", sign)
	return nil
}

/*
	获取服务端与本地的时间差
*/

func (b *Binance) LocolTimeSubServerTime(ctx context.Context) int64 {
	serverTime, err := b.GetServiceTime(ctx)
	if err != nil {
		logger.Logger.Error("get server time failed ", zap.Error(err))
		return 0
	}
	st := time.Unix(serverTime/1000, 0)
	nt := time.Now()
	return st.Sub(nt).Nanoseconds()
}

type Ping struct {
}

// ping service
func (b *Binance) Ping(ctx context.Context) (*Ping, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    util.PingURL,
	}
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Ping Failed", zap.Error(err))
		return nil, err
	}
	if _, ok := data["code"]; ok {
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
		URL:    util.ServiceTimeURL,
	}
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Ping Failed", zap.Error(err))
		return 0, err
	}
	if _, ok := data["code"]; ok {
		return time.Now().Unix(), err
	}
	p := new(SeviceTime)
	p.ServerTime = util.ToInt64(data["serverTime"])

	return p.ServerTime, nil
}

type Order struct {
	Symbol     string
	OrderID    int
	OrderType  int //0:default,1:maker,2:fok,3:ioc
	Side       TradeSide
	AvgPrice   float64
	Type       string // limit / market
	Fee        float64
	Price      float64
	DealAmount float64
	Amount     float64
	Status     TradeStatus
	OrderTime  int
}

/*
	SYMBOL : 交易对
	orderSide : 买 / 卖
	orderType : 订单类型
*/
func (b *Binance) PlaceOrder(ctx context.Context, amount, price string, symbol, orderType, orderSide string) (*Order, error) {
	r := &mod.ReqParam{
		Method: "POST",
		URL:    "/api/v3/order/test",
	}
	r.SetParam("symbol", symbol)
	r.SetParam("side", orderSide)
	r.SetParam("type", orderType)
	r.SetParam("newOrderRespType", "ACK")
	r.SetParam("quantity", amount)
	switch orderType {
	case "LIMIT":
		r.SetParam("timeInForce", "GTC")
		r.SetParam("price", price)
	case "MARKET":
		r.SetParam("newOrderRespType", "RESULT")
	}
	b.ParamsSigned(&r.Query)
	r.Header.Set("X-MBX-APIKEY", b.accessKey)
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Ping Failed", zap.Error(err))
		return nil, err
	}
	orderID := util.ToInt(data["orderId"])
	if orderID < 0 {
		return nil, fmt.Errorf("orderid error")
	}
	dealAmount := util.ToFloat64(data["executedQty"])
	cummulativeQuoteQty := util.ToFloat64(data["cummulativeQuoteQty"])
	avgPrice := 0.0
	if cummulativeQuoteQty > 0 && dealAmount > 0 {
		avgPrice = cummulativeQuoteQty / dealAmount
	}
	side := BUY
	if orderSide == "SELL" {
		side = SELL
	}
	return &Order{
		Symbol:     symbol,
		OrderID:    orderID,
		Price:      util.ToFloat64(price),
		DealAmount: dealAmount,
		Amount:     util.ToFloat64(amount),
		Side:       TradeSide(side),
		AvgPrice:   avgPrice,
		Status:     ORDER_NEW,
		OrderTime:  util.ToInt(data["transactTime"]),
	}, nil

}
