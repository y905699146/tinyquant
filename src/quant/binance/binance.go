package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
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
	timestamp  int64
}

func NewBinance(accKey, secKey, baseurl string) *Binance {
	b := &Binance{
		accessKey: accKey,
		secretKey: secKey,
		baseUrl:   baseurl,
	}

	b.timestamp = b.LocolTimeSubServerTime(context.Background())
	return b
}

/*
	SHA256生成签名，SECRETKEY为密钥，body为参数
*/
func (b *Binance) ParamsSigned(req *mod.ReqParam) error {
	req.SetParam("recvWindow", "60000")
	tonce := strconv.FormatInt(util.GetCurrentUnixNano()+b.timestamp, 10)[0:13]
	req.SetParam("timestamp", tonce)
	postMsg := req.Query.Encode()
	mac := hmac.New(sha256.New, []byte(b.secretKey))
	_, err := mac.Write([]byte(postMsg))
	if err != nil {
		return err
	}
	sign := hex.EncodeToString(mac.Sum(nil))
	req.SetParam("signature", sign)
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
	//下单接口
	SYMBOL : 交易对
	orderSide : 买 / 卖
	orderType : 订单类型
*/
func (b *Binance) PlaceOrder(ctx context.Context, amount, price string, symbol, orderType, orderSide string) (*Order, error) {
	r := &mod.ReqParam{
		Method: "POST",
		URL:    "/api/v3/order/test",
		APIKEY: b.accessKey,
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
	b.ParamsSigned(r)
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Place Order Failed", zap.Error(err))
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

func (b *Binance) LimitBuy(ctx context.Context, amount, price string, symbol string) (*Order, error) {
	return b.PlaceOrder(ctx, amount, price, symbol, "LIMIT", "BUY")
}

func (b *Binance) LimitSell(ctx context.Context, amount, price string, symbol string) (*Order, error) {
	return b.PlaceOrder(ctx, amount, price, symbol, "LIMIT", "SELL")
}

func (b *Binance) MarketBuy(ctx context.Context, amount, price string, symbol string) (*Order, error) {
	return b.PlaceOrder(ctx, amount, price, symbol, "MARKET", "BUY")
}

func (b *Binance) MarketSell(ctx context.Context, amount, price string, symbol string) (*Order, error) {
	return b.PlaceOrder(ctx, amount, price, symbol, "MARKET", "SELL")
}

type OrderStatus struct {
	Symbol              string  `json:"symbol"`
	OrderID             int     `json:"orderId"`
	ClientOrderID       string  `json:"clientOrderId"`
	Price               float64 `json:"price"`
	OrigQty             float64 `json:"origQty"`
	ExecutedQty         float64 `json:"executedQty"`
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	StopPrice           float64 `json:"stopPrice"`
	IceBergQty          float64 `json:"icebergQty"`
	Time                int64   `json:"time"`
	UpdateTime          int64   `json:"updateTime"`
	IsWorking           bool    `json:"isWorking"`
}

/*
	查询订单状态
*/

func GetOrderStatus(symbol string) (*OrderStatus, error) {
	return nil, nil
}

type CancelOrder struct {
	Symbol              string  `json:"symbol"`
	OrderID             int     `json:"orderId"`
	OrigClientOrderID   string  `json:"origClientOrderId"`
	ClientOrderID       string  `json:"clientOrderId"`
	TransactionTime     int64   `json:"transactTime"`
	Price               float64 `json:"price"`
	OrigQty             float64 `json:"origQty"`
	ExecutedQty         float64 `json:"executedQty"`
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
}

/*
	撤销订单

*/

type Balances struct {
	Asset  string
	Free   string
	Locked string
}

type AccountInfo struct {
	MakerCommission  int
	TakerCommission  int
	BuyerCommission  int
	SellerCommission int
	CanTrade         bool
	CanWithDraw      bool
	CanDeposit       bool
	UpdateTime       int64
	AccountType      string
	BalanceList      []*Balances
	PerMissions      []string
}

/*
	//获取账户信息
*/

func (b *Binance) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    "/api/v3/account",
		APIKEY: b.accessKey,
	}
	b.ParamsSigned(r)
	data, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Get Account Info Failed", zap.Error(err))
		return nil, err
	}
	acc := &AccountInfo{
		MakerCommission:  util.ToInt(data["makerCommission"]),
		TakerCommission:  util.ToInt(data["takerCommission"]),
		BuyerCommission:  util.ToInt(data["buyerCommission"]),
		SellerCommission: util.ToInt(data["sellerCommission"]),
		CanTrade:         util.ToBool(data["canTrade"]),
		CanWithDraw:      util.ToBool(data["canWithdraw"]),
		CanDeposit:       util.ToBool(data["canDeposit"]),
		UpdateTime:       util.ToInt64(data["updateTime"]),
		//	AccountType:      data["accountType"].(string),
	}
	if data["accountType"] != nil {
		acc.AccountType = data["accountType"].(string)
	}
	if data["balances"] != nil {
		for _, bs := range data["balances"].([]interface{}) {
			_bs := bs.(map[string]interface{})
			if _bs["asset"].(string) == "BTC" || _bs["asset"].(string) == "USDT" {
				acc.BalanceList = append(acc.BalanceList, &Balances{
					Asset:  _bs["asset"].(string),
					Free:   _bs["free"].(string),
					Locked: _bs["locked"].(string),
				})
			}
		}
	}
	if data["permissions"] != nil {
		for _, v := range data["permissions"].([]interface{}) {
			acc.PerMissions = append(acc.PerMissions, v.(string))
		}
	}
	return acc, nil
}

type TradeHistoryList []*TradeHistory

type TradeHistory struct {
	Symbol          string
	ID              int
	OrderID         int
	OrderListID     int
	Price           string
	Qty             string
	QuoteQty        string
	Commission      string
	CommissionAsset string
	Time            int64
	IsBuyer         bool
	IsMaker         bool
	IsBestMatch     bool
}

func (b *Binance) GetMyTradeshistory(ctx context.Context, symbol string) (TradeHistoryList, error) {
	r := &mod.ReqParam{
		Method: "GET",
		URL:    "/api/v3/myTrades",
		APIKEY: b.accessKey,
	}
	r.SetParam("symbol", symbol)
	b.ParamsSigned(r)
	_, err := util.HttpRequest(ctx, r)
	if err != nil {
		logger.Logger.Error("Binance Service Get Account history Trade Failed", zap.Error(err))
		return nil, err
	}
	th := make([]*TradeHistory, 0)
	return th, nil
}
