package config

import "time"

const (
	// WS URL Constants
	BaseWSUrl   = "ws://localhost:4000"
	EndpointUrl = "/v1/ws"

	// SDK Constants
	PingInterval  = 30 * time.Second
	PongWait      = 60 * time.Second
	EventChanSize = 1024
)
