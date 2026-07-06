package view

import "html/template"

// Config customizes how New discovers and parses view templates. Values
// of this type are produced by the With* helpers in this package and
// applied left to right; later options override earlier ones.
type Config func(c *config)

type config struct {
	funcs         template.FuncMap
	layoutPath    string // default "layout"
	extensionName string // default ".gotmpl"
}

// WithFuncs registers a template.FuncMap that is attached to the layout
// and every page before parsing. This must be used (not a later Funcs
// call on the returned templates) whenever the templates reference the
// helpers, because html/template resolves function names at parse time.
func WithFuncs(f template.FuncMap) Config {
	return func(c *config) { c.funcs = f }
}

// WithLayout overrides the layout file's path, expressed relative to
// the view directory passed to New and without the file extension.
// The default is "layout", which resolves to "<dir>/layout<ext>".
// Nested layouts such as "layouts/application" are supported.
func WithLayout(name string) Config {
	return func(c *config) { c.layoutPath = name }
}

// WithExtension overrides the file extension (including the leading
// dot) that New uses to detect template files. The default is
// ".gotmpl".
func WithExtension(ext string) Config {
	return func(c *config) { c.extensionName = ext }
}
