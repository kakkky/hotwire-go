# hotwire-go

Server-side helpers for building [Hotwire](https://hotwired.dev) Turbo apps in Go.

[![CI](https://github.com/kakkky/hotwire-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kakkky/hotwire-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/kakkky/hotwire-go.svg)](https://pkg.go.dev/github.com/kakkky/hotwire-go)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

## Overview

hotwire-go is a small toolkit that lets you build [Hotwire](https://hotwired.dev)
apps in Go. It targets the [Turbo](https://turbo.hotwired.dev) side of Hotwire
(Drive / Frames / Streams) and provides Go helpers that generate the exact HTML
tags, attributes, and meta elements Turbo expects — usable from both
`html/template` and [a-h/templ](https://github.com/a-h/templ) with the same
helper API.

- **`turbo` package** — Turbo helpers (Drive / Frames / Streams), an SSE
  handler, and a pluggable stream broker.
- **`view` package** — *optional* thin `html/template` renderer that follows
  the layout / partial / page convention. templ users do not need it.

hotwire-go does **not** ship the Turbo JavaScript runtime itself
(`turbo.ScriptImport()` emits a `<script>` tag that loads it from a CDN) and
is **not** a router or web framework. Stimulus helpers are tentatively in
scope (see Roadmap).

## Installation

Requires Go 1.25.1 or newer.

```bash
go get github.com/kakkky/hotwire-go
```

```go
import (
    "github.com/kakkky/hotwire-go/turbo"
    "github.com/kakkky/hotwire-go/view" // optional; html/template users only
)
```

## Features

### Turbo Drive

- CDN `<script>` tag helper (`ScriptImport`)
- Cache-control / view-transition / refresh-method / prefetch meta helpers
- `data-turbo-*` attribute helpers: confirm, submits-with, action, method,
  track, preload, permanent, temporary, disable / enable, and more
- `Redirect` helper that responds with HTTP 303 for Post/Redirect/Get

### Turbo Frames

- `<turbo-frame>` element helper with `src`, `loading`, `target`, `recurse`,
  `autoscroll`, and morph refresh attribute helpers
- Request inspection: `IsFrameRequest`, `FrameID`

### Turbo Streams

- All seven stream actions: `append`, `prepend`, `replace`, `update`,
  `remove`, `before`, `after`, plus `refresh`
- Extra stream attribute helpers: `targets`, `request-id`, `method="morph"`
- Response helpers: `IsStreamRequest`, `RequestID`, `StreamHeader`

### Turbo Streams over SSE (broadcasting)

- `StreamBroker` interface with an in-process default and a drop-in Redis
  PUB/SUB backend (`WithRedisStreamBroker`) for horizontal scaling
- `StreamSSEHandler` — ready-to-mount `http.Handler` with configurable
  heartbeat interval
- `StreamSourceSSE` — signed `<turbo-stream-source>` tag helper
- HMAC-SHA256 signed, TTL-scoped subscription tokens with same-origin check
- `Broadcast` — render multiple `StreamContent` values into a single message

### Templating

- Every helper works with both `html/template` (via `TemplateFuncMap`) and
  [a-h/templ](https://github.com/a-h/templ) (via component / spread-attribute
  syntax)
- `view` package for `html/template`: layout + partials (`_prefix`) + pages,
  lazy `Render(ctx, w)` handles

## Quick look

`html/template`:

```gotemplate
<head>
    {{ turboScriptImport }}
    {{ turboMetaViewTransition }}
</head>
<body>
    {{ turboFrame "posts" }}
        <p>content</p>
    {{ turboFrameEnd }}
    {{ turboStreamSourceSSE "messages" }}
</body>
```

a-h/templ:

```templ
<head>
    @turbo.ScriptImport()
    @turbo.MetaViewTransition()
</head>
<body>
    @turbo.Frame("posts") {
        <p>content</p>
    }
    @turbo.StreamSourceSSE("messages")
</body>
```

Wiring an SSE broadcaster:

```go
sb := turbo.NewStreamBroker() // or turbo.NewStreamBroker(turbo.WithRedisStreamBroker(rdb))
mux.Handle(turbo.StreamsSSEPath, turbo.StreamSSEHandler(sb))

// later, from any handler / worker:
_ = turbo.Broadcast(ctx, sb, "messages", streamContent)
```

> **Production note:** the SSE endpoint verifies subscription tokens with an
> HMAC key read from `HOTWIRE_TURBO_STREAM_SECRET`. If unset, a random key is
> generated per process — fine for a single-process dev setup, but every
> replica in a horizontally scaled deployment must export the same value.

## Documentation

Full API documentation lives on [pkg.go.dev](https://pkg.go.dev/github.com/kakkky/hotwire-go).
Each helper's godoc includes the exact HTML it produces, both `html/template`
and templ call examples, and links back to the Turbo Handbook / Reference.

A dedicated documentation site with end-to-end examples is planned.

## Compatibility

`turbo.ScriptImport()` pins the Turbo runtime it loads from the CDN. The
pinned version tracks what hotwire-go's helpers have been validated against
and is bumped alongside hotwire-go releases.

| hotwire-go | Turbo runtime |
| ---------- | ------------- |
| current    | 8.0.23        |

Applications that ship Turbo through their own bundler or import map do not
need `ScriptImport()` and are free to load any version compatible with the
helpers they use.

## Roadmap

- **Stimulus helpers** — tentative support for the Stimulus side of Hotwire.
- **WebSocket transport for Turbo Streams** — alongside the current SSE
  handler.
- **Redis broker** — expanded / production-hardened Redis-backed broker.

## License

[MIT](./LICENSE) © 2026 kakkky
