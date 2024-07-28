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

	if resp.StatusCode != 101 {
		ws.logger.Printf("Failed to upgrade connection to websocket: %v", resp.Status)
		return err
	}
	defer resp.Body.Close()

	conn.NetConn().SetDeadline(time.Now().Add(config.PongWait))

	done := make(chan struct{}, 1)

	go ws.read(conn, done)
	go ws.ping(ctx, conn, done)

	return nil
}

func (ws *WS) Consume() (<-chan DataFeed, <-chan SubscriptionMsg) {
	return ws.priceEventsChan, ws.infoEventsChan
}

func (ws *WS) read(conn *websocket.Conn, done chan struct{}) {
	defer ws.Close(conn, done)

	for {
		_, data, err := conn.ReadMessage()
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

func (ws *WS) ping(ctx context.Context, conn *websocket.Conn, done chan struct{}) {
	defer ws.Close(conn, done)

	ticker := time.NewTicker(config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.logger.Printf("Failed to write ping message: %v", err)
				return
			}
		}
	}
}

func (ws *WS) Close(conn *websocket.Conn, done chan struct{}) {
	_, ok := (<-done)
	if !ok {
		return
	}

	done <- struct{}{}

	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		ws.logger.Printf("Failed to write close message: %v", err)
	} else {
		ws.logger.Println("Connection closed")
	}

	conn.Close()
	close(ws.priceEventsChan)
	close(ws.infoEventsChan)
}
