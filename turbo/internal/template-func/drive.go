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

// MetaRefreshMethodMorph renders <meta name="turbo-refresh-method" content="morph">.
//
// When the same URL is revisited (typically after a form submission that
// redirects back to the current page), Turbo Drive applies a DOM morph
// instead of replacing <body>. Focus, scroll, and in-place DOM state are
// preserved wherever the old and new pages match. The default without this
// meta is "replace" (full body swap).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaRefreshMethodMorph }}
//
// Turbo Handbook — Smooth page refreshes with morphing:
// https://turbo.hotwired.dev/handbook/page_refreshes
func MetaRefreshMethodMorph() template.HTML {
	return `<meta name="turbo-refresh-method" content="morph">`
}

// MetaRefreshScrollPreserve renders <meta name="turbo-refresh-scroll" content="preserve">.
//
// On a same-URL refresh, Turbo Drive keeps the current scroll position
// instead of resetting to the top. Naturally pairs with
// MetaRefreshMethodMorph so a morphed refresh does not visibly jump. The
// default without this meta is "reset" (scroll to top).
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaRefreshScrollPreserve }}
//
// Turbo Handbook — Smooth page refreshes with morphing:
// https://turbo.hotwired.dev/handbook/page_refreshes
func MetaRefreshScrollPreserve() template.HTML {
	return `<meta name="turbo-refresh-scroll" content="preserve">`
}

// MetaDisablePrefetch renders <meta name="turbo-prefetch" content="false">.
//
// It disables Turbo Drive's default behavior of prefetching link targets
// when the user hovers over a link. Use this on pages where prefetch would
// waste bandwidth (many links, mobile-heavy audiences) or trigger unwanted
// side effects on the server (analytics counted per request, non-idempotent
// GET endpoints, etc.). Individual links can also be excluded with
// AttrDisablePrefetch.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboMetaDisablePrefetch }}
//
// Turbo Handbook — Prefetching Links on Hover:
// https://turbo.hotwired.dev/handbook/drive#prefetching-links-on-hover
func MetaDisablePrefetch() template.HTML {
	return `<meta name="turbo-prefetch" content="false">`
}
