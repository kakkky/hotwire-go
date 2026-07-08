package turbo

import (
	"html/template"
	"net/http"
)

// Redirect writes an HTTP 303 See Other response with the given URL in the
// Location header. Use this after a successful form submission when the
// client is Turbo Drive or Turbo Frames.
//
// Turbo Drive intercepts form submissions via fetch and expects a 3xx
// redirect for success. A 2xx HTML response is treated as a validation
// error re-render (Drive stays on the same URL and swaps the form area).
// Returning 303 explicitly instructs the client to follow the Location with
// a GET request, which is the correct behavior for the Post/Redirect/Get
// pattern and works for both Turbo and non-Turbo clients.
//
// See the Turbo Handbook for the underlying contract:
// https://turbo.hotwired.dev/handbook/drive#redirecting-after-a-form-submission
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// TemplateFuncMap returns an html/template.FuncMap containing every Turbo
// template helper provided by this package. It must be registered before
// parsing any file that references the helpers, because html/template
// resolves function names at parse time. When using the sibling view
// package, pass the result to view.WithFuncs so it is attached to the
// layout and every page before parsing:
//
//	r, err := view.New(fsys, "views", view.WithFuncs(turbo.TemplateFuncMap()))
//
// When driving html/template directly, call Funcs on the base template
// before ParseFiles / ParseFS:
//
//	tmpl := template.New("").Funcs(turbo.TemplateFuncMap()).ParseFiles(...)
//
// Helpers are registered under names prefixed with turbo (for example
// turboScriptImport). See each helper's godoc for its purpose, arguments,
// and template call form.
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"turboScriptImport":              func() template.HTML { return ScriptImport().HTMLTag() },
		"turboAttrConfirm":               func(s string) template.HTMLAttr { return AttrConfirm(s).HTMLAttr() },
		"turboAttrSubmitsWith":           func(s string) template.HTMLAttr { return AttrSubmitsWith(s).HTMLAttr() },
		"turboAttrActionAdvance":         func() template.HTMLAttr { return AttrActionAdvance().HTMLAttr() },
		"turboAttrActionReplace":         func() template.HTMLAttr { return AttrActionReplace().HTMLAttr() },
		"turboAttrMethodDelete":          func() template.HTMLAttr { return AttrMethodDelete().HTMLAttr() },
		"turboAttrMethodPatch":           func() template.HTMLAttr { return AttrMethodPatch().HTMLAttr() },
		"turboAttrMethodPut":             func() template.HTMLAttr { return AttrMethodPut().HTMLAttr() },
		"turboMetaVisitControlReload":    func() template.HTML { return MetaVisitControlReload().HTMLTag() },
		"turboMetaCacheControlNoPreview": func() template.HTML { return MetaCacheControlNoPreview().HTMLTag() },
		"turboMetaCacheControlNoCache":   func() template.HTML { return MetaCacheControlNoCache().HTMLTag() },
		"turboMetaViewTransition":        func() template.HTML { return MetaViewTransition().HTMLTag() },
		"turboMetaRefreshMethodMorph":    func() template.HTML { return MetaRefreshMethodMorph().HTMLTag() },
		"turboMetaRefreshScrollPreserve": func() template.HTML { return MetaRefreshScrollPreserve().HTMLTag() },
		"turboMetaDisablePrefetch":       func() template.HTML { return MetaDisablePrefetch().HTMLTag() },
		"turboMetaRoot":                  func(s string) template.HTML { return MetaRoot(s).HTMLTag() },
		"turboAttrTrackReload":           func() template.HTMLAttr { return AttrTrackReload().HTMLAttr() },
		"turboAttrTrackDynamic":          func() template.HTMLAttr { return AttrTrackDynamic().HTMLAttr() },
		"turboAttrDisableTurbo":          func() template.HTMLAttr { return AttrDisableTurbo().HTMLAttr() },
		"turboAttrEnableTurbo":           func() template.HTMLAttr { return AttrEnableTurbo().HTMLAttr() },
		"turboAttrPreload":               func() template.HTMLAttr { return AttrPreload().HTMLAttr() },
		"turboAttrDisablePrefetch":       func() template.HTMLAttr { return AttrDisablePrefetch().HTMLAttr() },
		"turboAttrPermanent":             func() template.HTMLAttr { return AttrPermanent().HTMLAttr() },
		"turboAttrTemporary":             func() template.HTMLAttr { return AttrTemporary().HTMLAttr() },
		"turboAttrDisableEval":           func() template.HTMLAttr { return AttrDisableEval().HTMLAttr() },
	}
}

// turboVersion pins the Turbo library version served by ScriptImport.
// Update in step with hotwire-go releases when validated against a new Turbo release.
const turboVersion = "8.0.23"

const scriptImportHTML = `<script type="module" src="https://cdn.jsdelivr.net/npm/@hotwired/turbo@` + turboVersion + `/dist/turbo.es2017-esm.js"></script>`

// ScriptImport renders the <script> tag that loads the Turbo runtime from a
// CDN. It must be placed inside <head>; without it, no Turbo behavior
// (Drive, Frames, Streams) is active on the page.
//
// The URL points to the ESM build served by jsDelivr, pinned to turboVersion.
// Applications that manage assets via a bundler or import map do not need
// this helper and can load Turbo through their own asset pipeline.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboScriptImport }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the component syntax:
//
//	@turbo.ScriptImport()
//
// Turbo installation reference:
// https://turbo.hotwired.dev/handbook/installing
func ScriptImport() Tag {
	return scriptImportHTML
}

// AttrConfirm renders data-turbo-confirm="{message}" on a link or form.
//
// Turbo shows a browser confirm dialog with the given message when the
// element is activated; the request is only issued if the user accepts.
// Applies to any element that initiates a Turbo request (forms and links
// with data-turbo-method), so it works across Drive, Frames, and Streams.
//
// The message is HTML-escaped before insertion; pass a plain string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<a href="..." {{ turboAttrMethodDelete }} {{ turboAttrConfirm "..." }}>...</a>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<a href="..." { turbo.AttrConfirm("...")... }>...</a>
//
// Turbo Reference — data-turbo-confirm:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrConfirm(message string) Attrs {
	return Attrs{Key: "data-turbo-confirm", Value: message}
}

// AttrSubmitsWith renders data-turbo-submits-with="{text}" on a form
// submitter (an input or button element).
//
// While the form is being submitted, Turbo replaces the submitter's label
// with the given text (for example, "Saving..."); the original label is
// restored after the request completes. Works with any Turbo-initiated form
// submission, so it applies across Drive, Frames, and Streams.
//
// The text is HTML-escaped before insertion; pass a plain string.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<button {{ turboAttrSubmitsWith "..." }}>...</button>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<button { turbo.AttrSubmitsWith("...")... }>...</button>
//
// Turbo Reference — data-turbo-submits-with:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrSubmitsWith(text string) Attrs {
	return Attrs{Key: "data-turbo-submits-with", Value: text}
}

// MetaVisitControlReload renders <meta name="turbo-visit-control" content="reload">.
//
// When Turbo navigates to this page — including navigations that originate
// from a <turbo-frame> — it performs a full browser reload instead of a
// Turbo visit. Use this on pages that are incompatible with the visit
// lifecycle (third-party embeds that assume a fresh document, pages that
// mutate global state during load, etc.).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaVisitControlReload }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the component syntax:
//
//	@turbo.MetaVisitControlReload()
//
// Turbo Reference — turbo-visit-control:
// https://turbo.hotwired.dev/reference/attributes#meta-tags
func MetaVisitControlReload() Tag {
	return `<meta name="turbo-visit-control" content="reload">`
}

// AttrActionAdvance renders data-turbo-action="advance" on a link or a
// <turbo-frame>.
//
// On a link, it forces Turbo Drive to record the visit as a new browser
// history entry (the default for regular navigation). On a <turbo-frame>,
// it promotes a frame navigation to a full page visit, so the URL bar and
// history stack update to reflect the destination of the frame.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<turbo-frame id="..." {{ turboAttrActionAdvance }}>...</turbo-frame>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<turbo-frame id="..." { turbo.AttrActionAdvance()... }>...</turbo-frame>
//
// Turbo Reference — data-turbo-action:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrActionAdvance() Attrs {
	return Attrs{Key: "data-turbo-action", Value: "advance"}
}

// AttrActionReplace renders data-turbo-action="replace" on a link or a
// <turbo-frame>.
//
// On a link, it tells Turbo Drive to replace the current history entry
// instead of pushing a new one, so the back button skips over the visit.
// On a <turbo-frame>, it also promotes a frame navigation to a page visit
// but replaces the current history entry rather than adding to it.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<a href="..." {{ turboAttrActionReplace }}>...</a>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<a href="..." { turbo.AttrActionReplace()... }>...</a>
//
// Turbo Reference — data-turbo-action:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrActionReplace() Attrs {
	return Attrs{Key: "data-turbo-action", Value: "replace"}
}

// AttrMethodDelete renders data-turbo-method="delete" on a link.
//
// When the link is activated, Turbo issues a DELETE request instead of the
// default GET. This lets you express destructive actions as plain links
// without wrapping them in a form. Pair with AttrConfirm to guard against
// accidental clicks. Works for links in Drive and inside Frames.
//
// Non-GET requests are ideally triggered by forms; use this only where a
// form would be awkward (compact icon lists, table row actions, etc.).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<a href="..." {{ turboAttrMethodDelete }} {{ turboAttrConfirm "..." }}>...</a>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<a href="..." { turbo.AttrMethodDelete()... } { turbo.AttrConfirm("...")... }>...</a>
//
// Turbo Reference — data-turbo-method:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrMethodDelete() Attrs {
	return Attrs{Key: "data-turbo-method", Value: "delete"}
}

// AttrMethodPatch renders data-turbo-method="patch" on a link.
//
// When the link is activated, Turbo issues a PATCH request instead of the
// default GET. Useful for "partial update" actions rendered as compact
// links (approve/reject buttons, toggles, etc.) where a full form would be
// overkill. Works for links in Drive and inside Frames.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<a href="..." {{ turboAttrMethodPatch }}>...</a>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<a href="..." { turbo.AttrMethodPatch()... }>...</a>
//
// Turbo Reference — data-turbo-method:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrMethodPatch() Attrs {
	return Attrs{Key: "data-turbo-method", Value: "patch"}
}

// AttrMethodPut renders data-turbo-method="put" on a link.
//
// When the link is activated, Turbo issues a PUT request instead of the
// default GET. Useful for "full replacement" actions expressed as links.
// Works for links in Drive and inside Frames.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	<a href="..." {{ turboAttrMethodPut }}>...</a>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<a href="..." { turbo.AttrMethodPut()... }>...</a>
//
// Turbo Reference — data-turbo-method:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrMethodPut() Attrs {
	return Attrs{Key: "data-turbo-method", Value: "put"}
}
