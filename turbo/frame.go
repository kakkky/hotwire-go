package turbo

import "net/http"

// IsFrameRequest reports whether the request originated from a
// <turbo-frame> navigation — that is, whether the Turbo-Frame request
// header is present.
//
// Turbo attaches the Turbo-Frame header to every fetch it issues on behalf
// of a <turbo-frame> (both src navigations and form submissions inside the
// frame). Use this to branch a handler between a full-page response and a
// frame-only response: when true, render just the frame's contents; when
// false, render the surrounding layout as usual.
//
// Turbo Handbook — Decompose with Turbo Frames:
// https://turbo.hotwired.dev/handbook/frames
func IsFrameRequest(r *http.Request) bool {
	return r.Header.Get("Turbo-Frame") != ""
}

// FrameID returns the value of the Turbo-Frame request header — the id of
// the <turbo-frame> element that initiated the request. When the request
// did not originate from a frame navigation, it returns "".
//
// Turbo matches a response back to the initiating frame by looking for a
// <turbo-frame> in the response whose id equals this value; a mismatch
// causes Turbo to log a "missing frame" error and abandon the swap.
// Embedding the returned id in the response's <turbo-frame id="..."> lets
// the same partial serve multiple frames without hard-coding the id.
//
// Turbo Handbook — Decompose with Turbo Frames:
// https://turbo.hotwired.dev/handbook/frames
func FrameID(r *http.Request) string {
	return r.Header.Get("Turbo-Frame")
}

// TurboFrame builds a <turbo-frame id="..."> element carrying the given
// id and any additional Frame-specific attributes (for example the
// results of AttrSrc or AttrLoadingLazy).
//
// The returned Elm renders the full <turbo-frame ...>{children}</turbo-frame>
// element when used as a templ.Component; for the html/template funcmap
// path, the "turboFrame" entry emits only the opening tag, paired with
// "turboFrameEnd" for the closing tag so template markup can be written
// in between.
//
// The id is required by the Turbo Frames contract: Turbo uses it to
// match responses back to the frame that initiated the navigation. It is
// HTML-escaped before insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the component-call syntax:
//
//	@turbo.TurboFrame("...") {
//	    ...
//	}
//
// Turbo Handbook — Decompose with Turbo Frames:
// https://turbo.hotwired.dev/handbook/frames
func TurboFrame(id string, extra ...Attrs) Elm {
	attrs := Attrs{{Key: "id", Value: id}}
	for _, extraAttrs := range extra {
		attrs = append(attrs, extraAttrs...)
	}
	return Elm{
		Tag:   Tag("turbo-frame"),
		Attrs: attrs,
	}
}

// AttrSrc renders src="{url}" on a <turbo-frame>.
//
// The frame fetches the given URL and swaps the matching <turbo-frame>
// from the response into itself. Combined with AttrLoadingLazy the
// fetch is deferred until the frame enters the viewport; without a
// loading attribute the fetch happens as soon as the containing page
// loads (loading="eager" is the default).
//
// The URL is HTML-escaped before insertion; pass a plain URL string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrSrc "...") }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrSrc("...")) {
//	    ...
//	}
//
// Turbo Handbook — Eager-Loading Frames:
// https://turbo.hotwired.dev/handbook/frames#eager-loading-frames
func AttrSrc(url string) Attrs {
	return Attrs{{Key: "src", Value: url}}
}

// AttrLoadingLazy renders loading="lazy" on a <turbo-frame>.
//
// Turbo defers loading the frame's src until the frame becomes visible
// in the viewport. Use it for below-the-fold content that is expensive
// to render on the server or that most visitors will not scroll to. The
// default without this attribute is loading="eager", which fetches
// immediately after the containing page loads.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrSrc "...") (turboAttrLoadingLazy) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrSrc("..."), turbo.AttrLoadingLazy()) {
//	    ...
//	}
//
// Turbo Handbook — Lazy-Loading Frames:
// https://turbo.hotwired.dev/handbook/frames#lazy-loading-frames
func AttrLoadingLazy() Attrs {
	return Attrs{{Key: "loading", Value: "lazy"}}
}

// AttrLoadingEager renders loading="eager" on a <turbo-frame>.
//
// Turbo fetches the frame's src as soon as the containing page loads.
// Because eager is already the default, this helper is only useful for
// explicitly overriding a dynamic or template-level value; most call
// sites can omit it.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrSrc "...") (turboAttrLoadingEager) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrSrc("..."), turbo.AttrLoadingEager()) {
//	    ...
//	}
//
// Turbo Handbook — Eager-Loading Frames:
// https://turbo.hotwired.dev/handbook/frames#eager-loading-frames
func AttrLoadingEager() Attrs {
	return Attrs{{Key: "loading", Value: "eager"}}
}

// AttrDisabled renders disabled on a <turbo-frame> (a boolean
// attribute with no value).
//
// The frame suspends all navigation: link clicks, form submissions, and
// src changes are ignored. Use it to temporarily "freeze" a frame while
// keeping its currently rendered content visible.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrDisabled) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrDisabled()) {
//	    ...
//	}
//
// Turbo Reference — disabled:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrDisabled() Attrs {
	return Attrs{{Key: "disabled", Value: true}}
}

// AttrTarget renders target="{id}" on a <turbo-frame>.
//
// Descendant links and forms whose activation would normally update this
// frame instead update the referenced element. Pass another
// <turbo-frame>'s id, or "_top" to escape frames entirely and navigate
// the whole window. Individual descendants can still override the target
// with a data-turbo-frame attribute.
//
// The value is HTML-escaped before insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrTarget "...") }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrTarget("...")) {
//	    ...
//	}
//
// Turbo Handbook — Targeting Navigation Into or Out of a Frame:
// https://turbo.hotwired.dev/handbook/frames#targeting-navigation-into-or-out-of-a-frame
func AttrTarget(id string) Attrs {
	return Attrs{{Key: "target", Value: id}}
}

// AttrRecurse renders recurse="{id}" on a <turbo-frame>.
//
// Place it on an intermediate <turbo-frame src="..."> in a response to
// tell Turbo where to keep looking for a target frame. When Turbo
// fetches a response for a frame and does not find a matching id at
// the top level, it looks for a frame carrying recurse~="{target id}"
// with a src, loads that intermediate frame first, and then searches
// the newly loaded content for the target frame — repeating as
// needed. Pass the id of the target frame that lives inside this
// intermediate frame's own response.
//
// The value is HTML-escaped before insertion; pass a plain id string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrRecurse "...") }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrRecurse("...")) {
//	    ...
//	}
//
// Turbo Reference — recurse:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrRecurse(id string) Attrs {
	return Attrs{{Key: "recurse", Value: id}}
}

// AttrAutoscroll renders autoscroll on a <turbo-frame> (a boolean
// attribute with no value).
//
// After the frame finishes loading, Turbo scrolls it into view.
// Fine-tune the scroll destination with the
// AttrAutoscrollBlockStart / Center / End / Nearest helpers, and
// the animation with AttrAutoscrollBehaviorAuto / Smooth. Without
// those extras the defaults are block="end" and behavior="auto".
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll()) {
//	    ...
//	}
//
// Turbo Reference — autoscroll:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscroll() Attrs {
	return Attrs{{Key: "autoscroll", Value: true}}
}

// AttrAutoscrollBlockStart renders data-autoscroll-block="start"
// on a <turbo-frame>.
//
// After the frame loads and scrolls into view, its top edge is aligned
// with the top of the viewport. Corresponds to
// Element.scrollIntoView({ block: "start" }). Takes effect only when
// AttrAutoscroll is also present on the same frame.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBlockStart) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBlockStart()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-block:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBlockStart() Attrs {
	return Attrs{{Key: "data-autoscroll-block", Value: "start"}}
}

// AttrAutoscrollBlockCenter renders data-autoscroll-block="center"
// on a <turbo-frame>.
//
// After the frame loads and scrolls into view, it is centered
// vertically in the viewport. Corresponds to
// Element.scrollIntoView({ block: "center" }). Takes effect only when
// AttrAutoscroll is also present on the same frame.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBlockCenter) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBlockCenter()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-block:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBlockCenter() Attrs {
	return Attrs{{Key: "data-autoscroll-block", Value: "center"}}
}

// AttrAutoscrollBlockEnd renders data-autoscroll-block="end" on a
// <turbo-frame>.
//
// After the frame loads and scrolls into view, its bottom edge is
// aligned with the bottom of the viewport. Corresponds to
// Element.scrollIntoView({ block: "end" }). This is the default when
// AttrAutoscroll is present without a data-autoscroll-block
// attribute; the helper is useful for making the alignment explicit.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBlockEnd) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBlockEnd()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-block:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBlockEnd() Attrs {
	return Attrs{{Key: "data-autoscroll-block", Value: "end"}}
}

// AttrAutoscrollBlockNearest renders data-autoscroll-block="nearest"
// on a <turbo-frame>.
//
// After the frame loads and scrolls into view, only the minimum scroll
// required to make it visible is applied. Corresponds to
// Element.scrollIntoView({ block: "nearest" }). Takes effect only when
// AttrAutoscroll is also present on the same frame.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBlockNearest) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBlockNearest()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-block:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBlockNearest() Attrs {
	return Attrs{{Key: "data-autoscroll-block", Value: "nearest"}}
}

// AttrAutoscrollBehaviorAuto renders
// data-autoscroll-behavior="auto" on a <turbo-frame>.
//
// The scroll-into-view step jumps instantly using the browser's default
// behavior. Corresponds to Element.scrollIntoView({ behavior: "auto" }).
// This is the default when AttrAutoscroll is present without a
// data-autoscroll-behavior attribute; the helper is useful for making
// the behavior explicit.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBehaviorAuto) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBehaviorAuto()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-behavior:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBehaviorAuto() Attrs {
	return Attrs{{Key: "data-autoscroll-behavior", Value: "auto"}}
}

// AttrAutoscrollBehaviorSmooth renders
// data-autoscroll-behavior="smooth" on a <turbo-frame>.
//
// The scroll-into-view step animates smoothly instead of jumping.
// Corresponds to Element.scrollIntoView({ behavior: "smooth" }). Takes
// effect only when AttrAutoscroll is also present on the same
// frame.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrAutoscroll) (turboAttrAutoscrollBehaviorSmooth) }}
//	    ...
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrAutoscroll(), turbo.AttrAutoscrollBehaviorSmooth()) {
//	    ...
//	}
//
// Turbo Reference — data-autoscroll-behavior:
// https://turbo.hotwired.dev/reference/frames#html-attributes
func AttrAutoscrollBehaviorSmooth() Attrs {
	return Attrs{{Key: "data-autoscroll-behavior", Value: "smooth"}}
}

// AttrRefreshMorph renders refresh="morph" on a <turbo-frame>.
//
// When the frame is reloaded — via a page refresh whose response
// contains a matching frame, or by calling reload() on the FrameElement
// — Turbo morphs the new content into the existing DOM instead of
// replacing it. This preserves in-place state such as focus, form
// values, and media playback wherever the old and new DOM match. Pair
// it with AttrSrc so the frame has a source to reload from.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboFrame "..." (turboAttrSrc "...") (turboAttrRefreshMorph) }}
//	{{ turboFrameEnd }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) via the TurboFrame helper:
//
//	@turbo.TurboFrame("...", turbo.AttrSrc("..."), turbo.AttrRefreshMorph())
//
// Turbo Handbook — Turbo frames (Page refreshes):
// https://turbo.hotwired.dev/handbook/page_refreshes#turbo-frames
func AttrRefreshMorph() Attrs {
	return Attrs{{Key: "refresh", Value: "morph"}}
}
