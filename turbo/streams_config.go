package turbo

import "time"

// StreamConfig customizes StreamSSEHandler by mutating its internal
// configuration struct. Values are produced by option helpers such as
// WithHeartbeatInterval rather than constructed directly, mirroring the
// StreamBrokerConfig / WithRedisStreamBroker pattern used elsewhere in
// this package.
type StreamConfig func(*streamConfigs)

type streamConfigs struct {
	heartbeatInterval time.Duration
}

// WithHeartbeatInterval returns a StreamConfig that sets how often
// StreamSSEHandler emits an SSE comment frame (":\n\n") on an otherwise
// idle connection. Reverse proxies and load balancers (nginx, ELB, and
// friends) will close a stream they see no bytes on for their idle
// timeout window; the heartbeat keeps a byte flowing so the connection
// stays open even when no Turbo Streams payload has been broadcast for
// a while.
//
// The default is 15 seconds, which sits comfortably under the common
// 30–60 second proxy defaults. Shortening it costs a few extra bytes
// per client per interval; lengthening it past the proxy's idle window
// will silently drop long-lived subscriptions. Tests that want to
// suppress heartbeat noise on the wire can pass a value larger than the
// test timeout (for example, one hour) rather than trying to disable
// them.
func WithHeartbeatInterval(d time.Duration) StreamConfig {
	return func(sc *streamConfigs) {
		sc.heartbeatInterval = d
	}
}
