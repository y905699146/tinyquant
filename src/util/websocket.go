package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
	. "tinyquant/src/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsConfig struct {
	WsUrl                 string //websocket url
	ProxyUrl              string // 代理URL
	ReqHeaders            map[string][]string
	readDeadLineTime      time.Duration      //读超时时间
	reconnectInterval     time.Duration      //重试时间
	HeartbeatIntervalTime time.Duration      //心跳时间
	ProtoHandleFunc       func([]byte) error //协议处理函数
	close                 chan bool
	subs                  [][]byte
	reConnectLock         *sync.Mutex
	IsAutoReconnect       bool //自动重连
	IsDump                bool
}

type WsConn struct {
	conn                   *websocket.Conn
	wsconfig               *WsConfig
	writeBufferChan        chan []byte
	pingMessageBufferChan  chan []byte
	pongMessageBufferChan  chan []byte
	closeMessageBufferChan chan []byte
	close                  chan bool
	subs                   [][]byte
	reConnectLock          *sync.Mutex
}

var dialer = &websocket.Dialer{
	Proxy:             http.ProxyFromEnvironment,
	HandshakeTimeout:  30 * time.Second,
	EnableCompression: true,
}

func NewWsConn(wsURL string, proxyURL string, handle func(msg []byte) error) *WsConn {
	return &WsConn{
		wsconfig: &WsConfig{
			ReqHeaders:        make(map[string][]string, 1),
			reconnectInterval: time.Second * 10,
			IsDump:            true,
			IsAutoReconnect:   true,
			ProxyUrl:          proxyURL,
			WsUrl:             wsURL,
		}}
}

func (w *WsConn) NewWebsocket() error {

	if w.wsconfig.HeartbeatIntervalTime == 0 {
		w.wsconfig.readDeadLineTime = time.Minute
	} else {
		w.wsconfig.readDeadLineTime = w.wsconfig.HeartbeatIntervalTime * 2
	}

	if err := w.connect(); err != nil {
		Logger.Panic("[%s] %s", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
	}

	w.close = make(chan bool, 1)
	w.pingMessageBufferChan = make(chan []byte, 10)
	w.pongMessageBufferChan = make(chan []byte, 10)
	w.closeMessageBufferChan = make(chan []byte, 10)
	w.writeBufferChan = make(chan []byte, 10)
	w.reConnectLock = new(sync.Mutex)

	go w.writeRequest()
	go w.receiveMessage()
	go w.exitHandler()
	return nil
}

func (w *WsConn) writeRequest() {
	var (
		heartTimer *time.Timer
		err        error
	)

	if w.wsconfig.HeartbeatIntervalTime == 0 {
		heartTimer = time.NewTimer(time.Hour)
	} else {
		heartTimer = time.NewTimer(w.wsconfig.HeartbeatIntervalTime)
	}

	for {
		select {
		case <-w.close:
			Logger.Info("[w][%s] close websocket , exiting write message goroutine.", zap.Error(fmt.Errorf("%s", w.wsconfig.WsUrl)))
			return
		case d := <-w.writeBufferChan:
			err = w.conn.WriteMessage(websocket.TextMessage, d)
		case d := <-w.pingMessageBufferChan:
			err = w.conn.WriteMessage(websocket.PingMessage, d)
		case d := <-w.pongMessageBufferChan:
			err = w.conn.WriteMessage(websocket.PongMessage, d)
		case d := <-w.closeMessageBufferChan:
			err = w.conn.WriteMessage(websocket.CloseMessage, d)
		case <-heartTimer.C:
			if w.wsconfig.HeartbeatIntervalTime > 0 {
				heartTimer.Reset(w.wsconfig.HeartbeatIntervalTime)
			}
		}

		if err != nil {
			Logger.Error("[w][%s] write message failed ", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
		}
	}
}

func (w *WsConn) receiveMessage() {
	//exit
	w.conn.SetCloseHandler(func(code int, text string) error {
		Logger.Error("[w][%s] websocket exiting [code=%d , text=%s]", zap.String("url :", w.wsconfig.WsUrl), zap.Int("code : ", code), zap.String("text :", text))
		//w.CloseWs()
		return nil
	})

	w.conn.SetPongHandler(func(pong string) error {
		Logger.Error("[%s] received [pong] %s", zap.String("url :", w.wsconfig.WsUrl), zap.String("pong :", pong))
		w.conn.SetReadDeadline(time.Now().Add(w.wsconfig.readDeadLineTime))
		return nil
	})

	w.conn.SetPingHandler(func(ping string) error {
		Logger.Error("[%s] received [ping] %s", zap.String("url :", w.wsconfig.WsUrl), zap.String("ping :", ping))
		w.conn.SetReadDeadline(time.Now().Add(w.wsconfig.readDeadLineTime))
		return nil
	})

	for {
		select {
		case <-w.close:
			Logger.Info("[w][%s] close websocket , exiting receive message goroutine.", zap.String("%s", w.wsconfig.WsUrl))
			return
		default:
			t, msg, err := w.conn.ReadMessage()
			if err != nil {
				Logger.Error("[w][%s] %s", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
				if w.wsconfig.IsAutoReconnect {
					Logger.Info("[w][%s] Unexpected Closed , Begin Retry Connect.", zap.String("url :", w.wsconfig.WsUrl))
					w.reconnect()
					continue
				}

				return
			}
			w.conn.SetReadDeadline(time.Now().Add(w.wsconfig.readDeadLineTime))
			switch t {
			case websocket.TextMessage: //文本消息
				w.wsconfig.ProtoHandleFunc(msg)
			case websocket.BinaryMessage:
				fmt.Println(string(msg))
			default:
				Logger.Error("[w][%s] error websocket message type , content is :\n %s \n", zap.String("url :", w.wsconfig.WsUrl), zap.String("msg :", string(msg)))
			}
		}
	}
}

func (w *WsConn) connect() error {
	if w.wsconfig.ProxyUrl != "" { //代理URL： 127.0.0.1：7890
		proxy, err := url.Parse(w.wsconfig.ProxyUrl)
		if err == nil {
			Logger.Info("[w][%s] proxy url failed ", zap.String("url :", w.wsconfig.WsUrl))
			dialer.Proxy = http.ProxyURL(proxy)
		} else {
			Logger.Error("[w][%s]parse proxy url [%s] err %s  ", zap.String("url :", w.wsconfig.WsUrl), zap.String("proxy url :", w.wsconfig.ProxyUrl), zap.Error(err))
		}
	}

	wsConn, resp, err := dialer.Dial(w.wsconfig.WsUrl, http.Header(w.wsconfig.ReqHeaders))
	if err != nil {
		Logger.Error("[w][%s] %s", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
		if w.wsconfig.IsDump && resp != nil {
			dumpData, _ := httputil.DumpResponse(resp, true)
			Logger.Debug("[w][%s] %s", zap.String("url :", w.wsconfig.WsUrl), zap.String("dumpData :", string(dumpData)))
		}
		return err
	}

	wsConn.SetReadDeadline(time.Now().Add(w.wsconfig.readDeadLineTime))

	if w.wsconfig.IsDump {
		dumpData, _ := httputil.DumpResponse(resp, true)
		Logger.Debug("[w][%s] dumpData %s", zap.String("url :", w.wsconfig.WsUrl), zap.String("dumpData :", string(dumpData)))
	}
	Logger.Info("[w][%s] connected", zap.String("url :", w.wsconfig.WsUrl))
	w.conn = wsConn
	return nil
}

func (w *WsConn) reconnect() {
	w.reConnectLock.Lock()
	defer w.reConnectLock.Unlock()

	w.conn.Close() //主动关闭一次
	var err error
	for retry := 1; retry <= 100; retry++ {
		err = w.connect()
		if err != nil {
			Logger.Error("[w] [%s] websocket reconnect fail , %s", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
		} else {
			break
		}
		time.Sleep(w.wsconfig.reconnectInterval * time.Duration(retry))
	}

	if err != nil {
		Logger.Error("[w] [%s] retry connect 100 count fail , begin exiting. ", zap.String("url :", w.wsconfig.WsUrl))
		w.CloseWs()
	}
	/*else {
		//re subscribe
		if w.ConnectSuccessAfterSendMessage != nil {
			msg := w.ConnectSuccessAfterSendMessage()
			w.SendMessage(msg)
			Logger.Info("[w] [%s] execute the connect success after send message=%s", zap.String("url :", w.wsconfig.WsUrl), zap.String("msg :", string(msg)))
			time.Sleep(time.Second) //wait response
		}

		for _, sub := range w.subs {
			Logger.Info("[w] re subscribe: ", zap.String("sub :", string(sub)))
			w.SendMessage(sub)
		}
	}*/
}

func (w *WsConn) CloseWs() {
	//w.close <- true
	close(w.close)
	close(w.writeBufferChan)
	close(w.closeMessageBufferChan)
	close(w.pingMessageBufferChan)
	close(w.pongMessageBufferChan)

	err := w.conn.Close()
	if err != nil {
		Logger.Error("[w][] close websocket error ", zap.String("url :", w.wsconfig.WsUrl), zap.Error(err))
	}
}

func (w *WsConn) SendMessage(msg []byte) {
	w.writeBufferChan <- msg
}

func (w *WsConn) SendPingMessage(msg []byte) {
	w.pingMessageBufferChan <- msg
}

func (w *WsConn) SendPongMessage(msg []byte) {
	w.pongMessageBufferChan <- msg
}

func (w *WsConn) SendCloseMessage(msg []byte) {
	w.closeMessageBufferChan <- msg
}

func (w *WsConn) SendJsonMessage(m interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	w.writeBufferChan <- data
	return nil
}

func (w *WsConn) exitHandler() {
	pingTicker := time.NewTicker(10 * time.Minute)
	pongTicker := time.NewTicker(time.Second)
	defer pingTicker.Stop()
	defer pongTicker.Stop()
	defer w.CloseWs()

	for {
		select {
		case t := <-pingTicker.C:
			w.SendPingMessage([]byte(strconv.Itoa(int(t.UnixNano() / int64(time.Millisecond)))))
		case t := <-pongTicker.C:
			w.SendPongMessage([]byte(strconv.Itoa(int(t.UnixNano() / int64(time.Millisecond)))))
		}
	}
}
