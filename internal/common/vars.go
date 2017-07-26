package common

import (
	"time"
)

// TODO: move these vars to Context?

// NodeTXPayloadQueueTTL defines the TTL of the node TXPayload queue
var NodeTXPayloadQueueTTL = time.Minute * 55

