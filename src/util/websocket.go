package util

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
	. "tinyquant/src/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsConfig struct {
	WsUrl                 string //websocket url
	ReqHeaders            map[string][]string
	readDeadLineTime      time.Duration //读超时时间
	reconnectTime         time.Duration //重试时间
	HeartbeatIntervalTime time.Duration //心跳时间
	close                 chan bool
	subs                  [][]byte
	reConnectLock         *sync.Mutex
}

type WsConn struct {
	conn *websocket.Conn
	WsConfig
	writeBufferChan        chan []byte
	pingMessageBufferChan  chan []byte
	pongMessageBufferChan  chan []byte
	closeMessageBufferChan chan []byte
	close                  chan bool
	subs                   [][]byte
	reConnectLock          *sync.Mutex
}

func (w *WsConn) NewWebsocket() error {

	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  30 * time.Second,
		EnableCompression: true,
	}
	wsConn, resp, err := dialer.Dial(w.WsUrl, http.Header(w.ReqHeaders))
	if err != nil {
		Logger.Error(" [ws][%s] request failedd ", zap.Error(fmt.Errorf("%s", w.WsUrl)), zap.Error(err))
		if resp != nil {
			dumpData, _ := httputil.DumpResponse(resp, true)
			Logger.Error("[ws][%s] %s", zap.Error(fmt.Errorf("%s", w.WsUrl)), zap.Error(fmt.Errorf("%s", string(dumpData))))
		}
	}
	wsConn.SetReadDeadline(time.Now().Add(w.readDeadLineTime))

	w.conn = wsConn

	w.close = make(chan bool, 1)
	w.pingMessageBufferChan = make(chan []byte, 10)
	w.pongMessageBufferChan = make(chan []byte, 10)
	w.closeMessageBufferChan = make(chan []byte, 10)
	w.writeBufferChan = make(chan []byte, 10)
	w.reConnectLock = new(sync.Mutex)

	return nil
}

func (w *WsConn) writeRequest() {
	var (
		heartTimer *time.Timer
		err        error
	)

	if w.HeartbeatIntervalTime == 0 {
		heartTimer = time.NewTimer(time.Hour)
	} else {
		heartTimer = time.NewTimer(w.HeartbeatIntervalTime)
	}

	for {
		select {
		case <-w.close:
			Logger.Info("[ws][%s] close websocket , exiting write message goroutine.", zap.Error(fmt.Errorf("%s", w.WsUrl)))
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
			if w.HeartbeatIntervalTime > 0 {
				err = w.conn.WriteMessage(websocket.TextMessage, w.HeartbeatData())
				heartTimer.Reset(w.HeartbeatIntervalTime)
			}
		}

		if err != nil {
			Logger.Error("[ws][%s] write message failed ", zap.Error(fmt.Errorf("%s", w.WsUrl)), zap.Error(err))
		}
	}
}
