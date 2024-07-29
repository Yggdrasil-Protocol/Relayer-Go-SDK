package relayergosdk

import (
	"encoding/json"
)

type DataFeed struct {
	Event  string `json:"event"`
	Price  string `json:"p"`
	FeedID string `json:"feedID"`
	T      int64  `json:"t"`
}

type SubscriptionFeed struct {
	FeedID string `json:"feedID"`
	Type   string `json:"type"`
}

type SubscriptionMsg struct {
	Event   string             `json:"event"`
	Success []SubscriptionFeed `json:"success"`
	Error   []SubscriptionFeed `json:"error"`
}

type event struct {
	Event string `json:"event"`
}

func sendEvent(data []byte, priceFeedChan chan<- DataFeed, subscriptionMsgChan chan<- SubscriptionMsg) error {
	res := event{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	switch res.Event {
	case "price":
		var priceFeed DataFeed
		err := json.Unmarshal(data, &priceFeed)
		if err != nil {
			return err
		}
		priceFeedChan <- priceFeed
	case "subscribe-status":
		var subscriptionMsg SubscriptionMsg
		err := json.Unmarshal(data, &subscriptionMsg)
		if err != nil {
			return err
		}
		subscriptionMsgChan <- subscriptionMsg
	case "subscribe-failed":
		return &SubscribeFailed{}
	default:
		return &InvalidEvent{}
	}

	return nil
}
