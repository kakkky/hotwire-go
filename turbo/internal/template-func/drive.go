package templatefunc

import "html/template"

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

// MetaViewTransition renders <meta name="view-transition" content="same-origin">.
//
// It opts the page into cross-document view transitions on browsers that
// support the View Transition API; when Turbo Drive navigates between two
// same-origin pages that both declare this meta tag, the browser animates
// the swap. On unsupported browsers the tag has no effect and Drive falls
// back to its regular render.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaViewTransition }}
//
// Turbo Handbook — View transitions:
// https://turbo.hotwired.dev/handbook/drive#view-transitions
func MetaViewTransition() template.HTML {
	return `<meta name="view-transition" content="same-origin">`
}
