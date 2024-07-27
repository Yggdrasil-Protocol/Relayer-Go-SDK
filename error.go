package relayergosdk

type SubscribeFailed struct{}

func (s *SubscribeFailed) Error() string {
	return "SubscribeFailed: Due to either no or invalid feedIDs passed."
}

type InvalidEvent struct{}

func (s *InvalidEvent) Error() string {
	return "InvalidEvent: Due to either no or invalid event passed."
}
