package config

import "time"

const (
	// WS URL Constants
	BaseWSUrl   = "feeds.yggdrasilprotocol.io"
	EndpointUrl = "/v1/ws"

	// SDK Constants
	PingInterval  = 30 * time.Second
	PingTimeout   = 60 * time.Second
	EventChanSize = 1024
)
