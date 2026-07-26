// Package stimulus provides server-side helpers for building web
// applications that use Hotwire's Stimulus runtime with Go and
// html/template.
//
// The package is a thin server-side layer: it does not ship any of
// the Stimulus JavaScript itself, and does not attempt to replace an
// HTTP mux or web framework. ScriptLoad emits the importmap and boot
// script that pull @hotwired/stimulus from jsDelivr and register the
// caller-supplied controller modules; the Attr* helpers render the
// data-* attributes that wire elements to controllers (controllers,
// actions, action params, targets, values, outlets, and CSS classes).
//
// Helpers are exposed two ways: register TemplateFuncMap on an
// html/template.FuncMap to call them as {{ stimulus* ... }} funcs in
// templates, or call them directly from a-h/templ using the
// spread-attributes syntax. Each exported helper's godoc shows both
// call forms.
//
// # Stimulus.debug toggle
//
// The package-level Debug variable controls whether ScriptLoad's
// module script sets Stimulus.debug = true. Assign it once at process
// start before any template that calls ScriptLoad renders; the flag
// is read at render time, so later mutations affect subsequent
// renders but not pages already sent to the client.
package stimulus
