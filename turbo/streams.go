package turbo

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
	attrs := Attrs{
		{Key: "action", Value: "append"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "prepend"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "replace"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "update"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "remove"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "before"},
		{Key: "target", Value: target},
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
	attrs := Attrs{
		{Key: "action", Value: "after"},
		{Key: "target", Value: target},
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
