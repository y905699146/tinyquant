package binance

const (
	TICKER_URI             = "ticker/24hr?symbol=%s"
	TICKERS_URI            = "ticker/allBookTickers"
	DEPTH_URI              = "depth?symbol=%s&limit=%d"
	ACCOUNT_URI            = "account?"
	ORDER_URI              = "order"
	UNFINISHED_ORDERS_INFO = "openOrders?"
	KLINE_URI              = "klines"
	SERVER_TIME_URL        = "time"
)

type TradeSide int

func (ts TradeSide) String() string {
	switch ts {
	case 1:
		return "BUY"
	case 2:
		return "SELL"
	case 3:
		return "BUY_MARKET"
	case 4:
		return "SELL_MARKET"
	default:
		return "UNKNOWN"
	}
}

const (
	BUY TradeSide = 1 + iota
	SELL
	BUY_MARKET
	SELL_MARKET
)

type TradeStatus int

const (
	ORDER_NEW              TradeStatus = iota //新建订单
	ORDER_PARTIALLY_FILLED                    //部分成交
	ORDER_FILLED                              //全部成交
	ORDER_CANCELED                            // 已撤销
	ORDER_PENDING_CANCEL                      //撤销中
	ORDER_REJECT                              //订单被拒绝
	ORDER_EXPIRED                             //订单过期
)
