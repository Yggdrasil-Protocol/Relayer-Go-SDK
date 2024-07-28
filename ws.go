package relayergosdk

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Yggdrasil-Protocol/Relayer-Go-SDK/config"
	"github.com/fasthttp/websocket"
)

type WS struct {
	url             *url.URL
	priceEventsChan chan DataFeed
	infoEventsChan  chan SubscriptionMsg
	dialer          *websocket.Dialer
	logger          *log.Logger
	conn            *websocket.Conn
	done            chan struct{}
}

func NewWS(feedIDs []string, logger *log.Logger, dialer *websocket.Dialer) *WS {
	ws := &WS{
		url: &url.URL{
			Scheme:   "wss",
			Host:     config.BaseWSUrl,
			Path:     config.EndpointUrl,
			RawQuery: "feedIDs=" + strings.Join(feedIDs, ","),
		},
		priceEventsChan: make(chan DataFeed, config.EventChanSize),
		infoEventsChan:  make(chan SubscriptionMsg, config.EventChanSize),
		logger:          logger,
		conn:            nil,
	}

	if dialer == nil {
		ws.dialer = websocket.DefaultDialer
		ws.dialer.EnableCompression = true
	}

	return ws
}

func (ws *WS) Subscribe(ctx context.Context) error {
	conn, resp, err := ws.dialer.DialContext(ctx, ws.url.String(), nil)
	if err != nil {
		ws.logger.Printf("Failed to connect to %s: %v", ws.url.String(), err)
		return err
	}
	ws.conn = conn
	ws.done = make(chan struct{}, 1)

	if resp.StatusCode != 101 {
		ws.logger.Printf("Failed to upgrade connection to websocket: %v", resp.Status)
		return err
	}
	defer resp.Body.Close()

	conn.NetConn().SetDeadline(time.Now().Add(config.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.NetConn().SetDeadline(time.Now().Add(config.PongWait))
		return nil
	})

	go ws.read()
	go ws.ping(ctx)

	return nil
}

func (ws *WS) Consume() (<-chan DataFeed, <-chan SubscriptionMsg) {
	return ws.priceEventsChan, ws.infoEventsChan
}

func (ws *WS) read() {
	defer ws.Close()

	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			ws.logger.Printf("Failed to read message: %v", err)
			return
		}

		err = sendEvent(data, ws.priceEventsChan, ws.infoEventsChan)
		if err != nil {
			ws.logger.Printf("Failed to send event to chan: %v", err)
			continue
		}
	}
}

func (ws *WS) ping(ctx context.Context) {
	defer ws.Close()

	ticker := time.NewTicker(config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.done:
			return
		case <-ticker.C:
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.logger.Printf("Failed to write ping message: %v", err)
				return
			} else {
				ws.conn.NetConn().SetDeadline(time.Now().Add(config.PongWait))
			}
		}
	}
}

func (ws *WS) Close() {
	select {
	case _, ok := <-ws.done:
		if !ok {
			return
		}
	}

	ws.done <- struct{}{}

	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		ws.logger.Printf("Failed to write close message: %v", err)
	} else {
		ws.logger.Println("Connection closed")
	}

	ws.conn.Close()
	close(ws.priceEventsChan)
	close(ws.infoEventsChan)
}
