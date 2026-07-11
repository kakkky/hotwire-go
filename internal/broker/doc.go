// Package broker provides a small pub/sub hub keyed by named streams,
// used by the sibling turbo package to fan Turbo Streams payloads out
// to connected clients. The [Broker] interface abstracts the backend
// so a single-process default and a Redis PUB/SUB implementation can
// be swapped without touching call sites.
//
// This package lives under internal/ and is not part of the public
// API; callers should reach it through the aliases exported from the
// turbo package (StreamBroker, StreamBrokerConfig, NewStreamBroker,
// WithRedisStreamBroker).
package broker
