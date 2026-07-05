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

// MetaCacheControlNoPreview renders <meta name="turbo-cache-control" content="no-preview">.
//
// The page is still cached and served instantly during restoration visits
// (browser back/forward), but Turbo Drive will not display the cached copy
// as a preview during application visits (regular link clicks). Use this on
// pages where showing a brief stale snapshot would be misleading — for
// example, screens whose state depends on authentication or real-time data.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaCacheControlNoPreview }}
//
// Turbo Handbook — Opting Out of Caching:
// https://turbo.hotwired.dev/handbook/building#opting-out-of-caching
func MetaCacheControlNoPreview() template.HTML {
	return `<meta name="turbo-cache-control" content="no-preview">`
}

// MetaCacheControlNoCache renders <meta name="turbo-cache-control" content="no-cache">.
//
// The page is never cached by Turbo Drive; every visit — including
// restoration visits from the browser back/forward buttons — issues a fresh
// network request. Use this for pages that must never be re-displayed from
// cache (for example, screens with strict privacy or freshness requirements).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaCacheControlNoCache }}
//
// Turbo Handbook — Opting Out of Caching:
// https://turbo.hotwired.dev/handbook/building#opting-out-of-caching
func MetaCacheControlNoCache() template.HTML {
	return `<meta name="turbo-cache-control" content="no-cache">`
}
