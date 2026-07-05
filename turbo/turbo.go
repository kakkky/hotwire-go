package turbo

import (
	"html/template"
	"net/http"

	templatefunc "github.com/kakkky/hotwire-go/turbo/internal/template-func"
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
// template helper provided by this package. Register it with the template
// before parsing any file that references the helpers, since html/template
// resolves function names at parse time:
//
//	tmpl := template.New("").Funcs(turbo.TemplateFuncMap()).ParseFiles(...)
//
// Helpers are registered under names prefixed with turbo (for example
// turboScriptImport). See each helper's godoc for its purpose, arguments,
// and template call form.
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"turboScriptImport":              templatefunc.ScriptImport,
		"turboAttrConfirm":               templatefunc.AttrConfirm,
		"turboAttrSubmitsWith":           templatefunc.AttrSubmitsWith,
		"turboAttrActionAdvance":         templatefunc.AttrActionAdvance,
		"turboAttrActionReplace":         templatefunc.AttrActionReplace,
		"turboAttrMethodDelete":          templatefunc.AttrMethodDelete,
		"turboAttrMethodPatch":           templatefunc.AttrMethodPatch,
		"turboAttrMethodPut":             templatefunc.AttrMethodPut,
		"turboMetaVisitControlReload":    templatefunc.MetaVisitControlReload,
		"turboMetaCacheControlNoPreview": templatefunc.MetaCacheControlNoPreview,
		"turboMetaCacheControlNoCache":   templatefunc.MetaCacheControlNoCache,
		"turboMetaViewTransition":        templatefunc.MetaViewTransition,
		"turboMetaRefreshMethodMorph":    templatefunc.MetaRefreshMethodMorph,
		"turboMetaRefreshScrollPreserve": templatefunc.MetaRefreshScrollPreserve,
		"turboMetaDisablePrefetch":       templatefunc.MetaDisablePrefetch,
		"turboMetaRoot":                  templatefunc.MetaRoot,
		"turboAttrTrackReload":           templatefunc.AttrTrackReload,
		"turboAttrTrackDynamic":          templatefunc.AttrTrackDynamic,
		"turboAttrDisableTurbo":          templatefunc.AttrDisableTurbo,
		"turboAttrEnableTurbo":           templatefunc.AttrEnableTurbo,
		"turboAttrPreload":               templatefunc.AttrPreload,
		"turboAttrDisablePrefetch":       templatefunc.AttrDisablePrefetch,
		"turboAttrPermanent":             templatefunc.AttrPermanent,
	}
}
