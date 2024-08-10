package main

import (
	"context"
	"log"
	"os"
	"time"

	rel "github.com/Yggdrasil-Protocol/Relayer-Go-SDK"
)

func main() {
	logger := log.New(os.Stdout, "relayergosdk: ", log.LstdFlags)
	feedIDs := []string{"SPOT:BTC_USDT", "SPOT:ETH_USDT"}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(35*time.Second))
	defer cancel()

	ws := rel.NewWS(feedIDs, logger, nil)
	defer ws.Close()

	dataChan, infoChan := ws.Consume()

	err := ws.Subscribe(ctx)
	if err != nil {
		logger.Panicf("Failed to subscribe: %v", err)
	}

	for {
		select {
		case price, ok := <-dataChan:
			if !ok {
				return
			}
			logger.Printf("Received price feed event: %+v", price)
		case info, ok := <-infoChan:
			if !ok {
				return
			}
			logger.Printf("Received info event: %+v", info)
		}
	}
}
