package turbo

import (
	"net/http"
	"strings"
)

// IsStreamRequest reports whether the client accepts a Turbo Streams
// response — that is, whether "text/vnd.turbo-stream.html" appears in
// the request's Accept header.
//
// Turbo appends the Turbo Streams MIME type to Accept on every form
// submission it intercepts (via FormSubmission#prepareRequest). Use
// this to branch a handler between a full HTML response for direct
// navigation and a <turbo-stream> response for Turbo-driven
// submissions: when true, set Content-Type to
// "text/vnd.turbo-stream.html" on the response and write one or more
// StreamX elements; when false, render the surrounding layout as usual.
//
// Turbo source — StreamMessage.contentType:
// https://github.com/hotwired/turbo/blob/v8.0.23/src/core/streams/stream_message.js#L4
//
// Turbo source — FormSubmission (attaching the Accept type):
// https://github.com/hotwired/turbo/blob/v8.0.23/src/core/drive/form_submission.js#L118
func IsStreamRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/vnd.turbo-stream.html")
}

// RequestID returns the value of the X-Turbo-Request-Id request header —
// a UUID Turbo generates for every fetch it issues and remembers on the
// client for a short window (up to 20 recent ids). When the request did
// not originate from a Turbo fetch, it returns "".
//
// Echo this value back on a broadcast <turbo-stream action="refresh">
// via AttrRequestID so the originating client, which still remembers its
// own id in recentRequests, skips its own refresh event and only the
// other clients apply it. Without this echo, the submitter's page would
// refresh twice: once from the direct response and once from the
// broadcast.
//
// Turbo source — X-Turbo-Request-Id / recentRequests:
// https://github.com/hotwired/turbo/blob/v8.0.23/src/http/fetch.js#L10
func RequestID(r *http.Request) string {
	return r.Header.Get("X-Turbo-Request-Id")
}

// StreamHeader sets Content-Type on the response to
// "text/vnd.turbo-stream.html; charset=utf-8" so Turbo processes the body
// as one or more <turbo-stream> actions rather than rendering it as a
// full HTML document.
//
// Turbo dispatches fetch responses by inspecting Content-Type: when the
// value contains "text/vnd.turbo-stream.html", the body is parsed as
// Turbo Streams and each <turbo-stream> element is applied to the DOM.
// Any other Content-Type is treated as a regular HTML response and
// rendered into the initiating frame or as a full page visit. Omitting
// this header silently drops the stream — the request succeeds but no
// DOM mutation is applied.
//
// Call this before WriteHeader and any body write on a handler that
// responds with StreamX elements. Pair with IsStreamRequest to gate the
// stream branch:
//
//	if turbo.IsStreamRequest(r) {
//	    turbo.StreamHeader(w)
//	    w.WriteHeader(http.StatusOK)
//	    // write one or more <turbo-stream> elements
//	    return
//	}
//
// Turbo source — StreamMessage.contentType:
// https://github.com/hotwired/turbo/blob/v8.0.23/src/core/streams/stream_message.js#L4
func StreamHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
}

// StreamAppend builds a <turbo-stream action="append" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="append" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamAppend" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The append action inserts the template content as the last children
// of the target; existing children are left in place. Elements carrying an
// id duplicated in the target are re-inserted rather than duplicated (Turbo
// dedupes by id on append). The target is HTML-escaped before insertion;
// pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamAppend "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamAppend("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.append:
// https://turbo.hotwired.dev/reference/streams#append
func StreamAppend(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "append"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamPrepend builds a <turbo-stream action="prepend" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="prepend" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamPrepend" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The prepend action inserts the template content as the first
// children of the target; existing children are pushed down. Elements
// carrying an id duplicated in the target are re-inserted rather than
// duplicated. The target is HTML-escaped before insertion; pass a plain
// id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamPrepend "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamPrepend("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.prepend:
// https://turbo.hotwired.dev/reference/streams#prepend
func StreamPrepend(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "prepend"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamReplace builds a <turbo-stream action="replace" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="replace" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamReplace" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The replace action removes the target element and all its
// descendants and swaps in the template content in its place. Use
// StreamUpdate instead when only the inner content should change and the
// target element itself should remain. The target is HTML-escaped before
// insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamReplace "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamReplace("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.replace:
// https://turbo.hotwired.dev/reference/streams#replace
func StreamReplace(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "replace"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamUpdate builds a <turbo-stream action="update" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="update" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamUpdate" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The update action replaces only the target's children with the
// template content; the target element itself is preserved (its
// attributes, id, and any listeners bound to it stay intact). Use
// StreamReplace when the target element itself should be exchanged. The
// target is HTML-escaped before insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamUpdate "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamUpdate("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.update:
// https://turbo.hotwired.dev/reference/streams#update
func StreamUpdate(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "update"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamRemove builds a <turbo-stream action="remove" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="remove" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamRemove" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The remove action deletes the target element from the DOM; any
// template content is ignored by Turbo (an empty <template> is still
// emitted for structural consistency with the other actions). The target
// is HTML-escaped before insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamRemove "..." }}
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamRemove("...")
//
// Turbo Reference — StreamActions.remove:
// https://turbo.hotwired.dev/reference/streams#remove
func StreamRemove(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "remove"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamBefore builds a <turbo-stream action="before" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="before" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamBefore" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The before action inserts the template content as the target's
// immediately previous sibling; the target itself is left untouched. Use
// StreamPrepend when the content should become a child of the target
// instead of a sibling. The target is HTML-escaped before insertion; pass
// a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamBefore "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamBefore("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.before:
// https://turbo.hotwired.dev/reference/streams#before
func StreamBefore(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "before"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamAfter builds a <turbo-stream action="after" target="..."> element
// carrying the given target and any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="after" target="..."><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamAfter" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// The target is required by the Turbo Streams contract: Turbo dispatches
// on action and locates the DOM node to mutate by target (a single element
// id). The after action inserts the template content as the target's
// immediately next sibling; the target itself is left untouched. Use
// StreamAppend when the content should become a child of the target
// instead of a sibling. The target is HTML-escaped before insertion; pass
// a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamAfter "..." }}
//	    ...
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamAfter("...") {
//	    ...
//	}
//
// Turbo Reference — StreamActions.after:
// https://turbo.hotwired.dev/reference/streams#after
func StreamAfter(target string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "action", Value: "after"}}
	if target != "" {
		attrs = append(attrs, Attrs{{Key: "target", Value: target}}...)
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// StreamRefresh builds a <turbo-stream action="refresh"> element carrying
// any additional Stream-specific attributes.
//
// The returned Elm renders the full
// <turbo-stream action="refresh"><template>{children}</template></turbo-stream>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboStreamRefresh" entry emits only the opening
// <turbo-stream ...><template> pair, paired with "turboStreamEnd" for
// </template></turbo-stream> so template markup can be written in between.
//
// Unlike the other stream actions, refresh does not take a target: Turbo
// re-fetches the current page and morphs the new document into the
// existing one. Any template content is ignored by Turbo (an empty
// <template> is still emitted for structural consistency with the other
// actions). Pair with a request-id via extra Attrs to deduplicate refresh
// events broadcast to multiple clients — the client that initiated the
// change receives its own event and ignores it.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboStreamRefresh }}
//	{{ turboStreamEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.StreamRefresh()
//
// Turbo Reference — StreamActions.refresh:
// https://turbo.hotwired.dev/reference/streams#refresh
func StreamRefresh(extra ...Attrs) Elm {
	attrs := Attrs{
		{Key: "action", Value: "refresh"},
	}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:      Tag("turbo-stream"),
		InnerTag: Tag("template"),
		Attrs:    attrs,
	}
}

// AttrTargets renders targets="{selector}" on a <turbo-stream>.
//
// Turbo applies the stream action to every element matched by the given
// CSS selector, rather than to a single element identified by id. Use it
// with StreamAppend, StreamPrepend, StreamReplace, StreamUpdate,
// StreamRemove, StreamBefore, or StreamAfter when the same mutation
// should fan out to multiple targets in one stream element (for example,
// removing every ".notification" or updating every ".unread" badge).
// The targets attribute is mutually exclusive with target: pass an empty
// string as the StreamX target argument and add AttrTargets via extra
// Attrs so only targets is emitted.
//
// The selector is HTML-escaped before insertion; pass a plain CSS
// selector string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<turbo-stream action="..." {{ turboAttrTargets "..." }}>...</turbo-stream>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via a StreamX helper's extra Attrs:
//
//	@turbo.StreamRemove("", turbo.AttrTargets("...")) {
//	    ...
//	}
//
// Turbo Reference — targets:
// https://turbo.hotwired.dev/reference/streams#targeting-multiple-elements
func AttrTargets(selector string) Attrs {
	return Attrs{{Key: "targets", Value: selector}}
}

// AttrRequestID renders request-id="{id}" on a <turbo-stream>.
//
// Turbo uses the request-id to deduplicate refresh events when the same
// page-wide refresh is broadcast to multiple clients over Turbo Streams:
// the client that initiated the underlying change receives its own
// refresh event and skips it, while other clients apply it normally. Pair
// with StreamRefresh; other actions ignore this attribute.
//
// The id is HTML-escaped before insertion; pass a plain id string
// (typically the same id set on the request that produced the change,
// for example via the X-Turbo-Request-Id header).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<turbo-stream action="refresh" {{ turboAttrRequestID "..." }}></turbo-stream>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via a StreamRefresh's extra Attrs:
//
//	@turbo.StreamRefresh(turbo.AttrRequestID("..."))
//
// Turbo Reference — StreamActions.refresh:
// https://turbo.hotwired.dev/reference/streams#refresh
func AttrRequestID(id string) Attrs {
	return Attrs{{Key: "request-id", Value: id}}
}

// AttrMethodMorph renders method="morph" on a <turbo-stream>.
//
// Turbo switches the stream action from full-node replacement to a
// morphdom-based diff swap: attributes are patched in place, children
// are reconciled by id where possible, and event listeners, form state,
// and focus bound to unchanged nodes are preserved. Pair with
// StreamReplace, StreamUpdate, or StreamRefresh; other actions ignore
// this attribute. Note: this attribute is scoped to <turbo-stream> and
// is distinct from AttrRefreshMorph, which renders refresh="morph" on
// <turbo-frame> for frame-level page refresh behavior.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<turbo-stream action="replace" target="..." {{ turboAttrMethodMorph }}>...</turbo-stream>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via a StreamX helper's extra Attrs:
//
//	@turbo.StreamReplace("...", turbo.AttrMethodMorph()) {
//	    ...
//	}
//
// Turbo Reference — Morphing Turbo Stream actions:
// https://turbo.hotwired.dev/reference/streams#morphing
func AttrMethodMorph() Attrs {
	return Attrs{{Key: "method", Value: "morph"}}
}
