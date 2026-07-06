package view

import (
	"html/template"
	"maps"
)

// Config customizes how New discovers and parses view templates. Values
// of this type are produced by the With* helpers in this package and
// applied left to right; later Configs override earlier ones.
type Config func(c *config)

type config struct {
	funcs         template.FuncMap
	layoutPath    string // default "layout"
	extensionName string // default ".gotmpl"
}

// WithFuncs registers one or more template.FuncMaps that are attached
// to the layout and every page before parsing. Later maps override
// earlier ones on key collisions, so callers can pass helper packages
// in ascending priority (for example
// view.WithFuncs(turbo.TemplateFuncMap(), stimulus.TemplateFuncMap())
// makes stimulus helpers win on any shared name).
//
// This must be used (not a later Funcs call on the returned templates)
// whenever the templates reference the helpers, because html/template
// resolves function names at parse time.
func WithFuncs(funcs ...template.FuncMap) Config {
	return func(c *config) {
		if len(funcs) == 0 {
			return
		}
		var size int
		for _, fm := range funcs {
			size += len(fm)
		}
		merged := make(template.FuncMap, size)
		for _, fm := range funcs {
			maps.Copy(merged, fm)
		}
		c.funcs = merged
	}
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
