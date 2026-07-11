package turbo

import (
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/broker"
	"github.com/kakkky/hotwire-go/internal/elm"
	"github.com/kakkky/hotwire-go/internal/tag"
)

// Attrs is the return type of every attribute helper in this package (for
// example AttrConfirm, AttrMethodDelete). It carries a collection of HTML
// attributes in an engine-neutral form so the same helper works with both
// html/template — via the HTMLAttr method registered through
// TemplateFuncMap — and a-h/templ (https://github.com/a-h/templ) via the
// spread-attributes syntax `{ turbo.AttrX(...)... }`.
type Attrs = attrs.Attrs

// Tag is the return type of every element helper in this package (for
// example ScriptImport, MetaVisitControlReload). It carries a
// pre-rendered HTML fragment in an engine-neutral form so the same helper
// works with both html/template — via the HTMLTag method registered
// through TemplateFuncMap — and a-h/templ
// (https://github.com/a-h/templ) via the component-call syntax
// `@turbo.TagX()`.
type Tag = tag.Tag

// Elm is the return type of every element helper in this package that
// wraps children (for example Frame). It carries a structured tag
// name and its attributes in an engine-neutral form so the same helper
// works with both html/template — where TemplateFuncMap registers
// separate opening and closing funcmap entries (for example
// turboFrame and turboFrameEnd) that template markup is written
// between — and a-h/templ (https://github.com/a-h/templ) via the
// component-call syntax `@turbo.TurboX(...) { ... }`.
type Elm = elm.Elm

// StreamBroker is a pub/sub hub for Turbo Streams bytes: Broadcast
// (and any custom publish path the application writes) sends bytes at
// a named stream, and every current subscriber of that stream receives
// them. Construct one with NewStreamBroker and share the returned
// value across handlers via dependency injection.
type StreamBroker = broker.Broker

// StreamBrokerConfig customizes how NewStreamBroker builds the
// underlying pub/sub backend — for example, switching from the default
// in-process implementation to a Redis PUB/SUB-backed one for
// horizontally scaled deployments. Values are produced by
// backend-specific factory helpers rather than constructed directly.
type StreamBrokerConfig = broker.Config

// WithRedisStreamBroker returns a StreamBrokerConfig that routes NewStreamBroker's
// Publish and Subscribe through Redis PUB/SUB instead of the default
// in-process channel implementation. The client is used as-is — pass an
// already-configured *redis.Client (or any implementation of
// redis.UniversalClient) so its connection pool and options can be
// shared with the rest of the application. NewStreamBroker only calls
// Publish and Subscribe on it, so a single-node client, a ClusterClient,
// and a FailoverClient are all acceptable inputs.
var WithRedisStreamBroker = broker.WithRedis
