package templatefunc

import (
	"fmt"
	"html/template"
)

// turboVersion pins the Turbo library version served by ScriptImport.
// Update in step with hotwire-go releases when validated against a new Turbo release.
const turboVersion = "8.0.23"

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
// Turbo installation reference:
// https://turbo.hotwired.dev/handbook/installing
func ScriptImport() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<script type="module" src="https://cdn.jsdelivr.net/npm/@hotwired/turbo@%s/dist/turbo.es2017-esm.js"></script>`,
		turboVersion,
	))
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
//	<a href="/posts/1" {{ turboAttrMethodDelete }} {{ turboAttrConfirm "Are you sure?" }}>Delete</a>
//
// Turbo Reference — data-turbo-confirm:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrConfirm(message string) template.HTMLAttr {
	return template.HTMLAttr(fmt.Sprintf(
		`data-turbo-confirm="%s"`,
		template.HTMLEscapeString(message),
	))
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
//	<button {{ turboAttrSubmitsWith "Saving..." }}>Save</button>
//
// Turbo Reference — data-turbo-submits-with:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrSubmitsWith(text string) template.HTMLAttr {
	return template.HTMLAttr(fmt.Sprintf(
		`data-turbo-submits-with="%s"`,
		template.HTMLEscapeString(text),
	))
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
// Turbo Reference — turbo-visit-control:
// https://turbo.hotwired.dev/reference/attributes#meta-tags
func MetaVisitControlReload() template.HTML {
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
//	<turbo-frame id="msg" {{ turboAttrActionAdvance }}>...</turbo-frame>
//
// Turbo Reference — data-turbo-action:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrActionAdvance() template.HTMLAttr {
	return `data-turbo-action="advance"`
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
//	<a href="/edit" {{ turboAttrActionReplace }}>Edit</a>
//
// Turbo Reference — data-turbo-action:
// https://turbo.hotwired.dev/reference/attributes#data-attributes
func AttrActionReplace() template.HTMLAttr {
	return `data-turbo-action="replace"`
}
