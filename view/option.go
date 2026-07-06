package view

import "html/template"

type Config func(c *config)

type config struct {
	funcs         template.FuncMap
	layoutPath    string // default "layout"
	extensionName string // default ".gotmpl"
}

func WithFuncs(f template.FuncMap) Config {
	return func(c *config) { c.funcs = f }
}

func WithLayout(name string) Config {
	return func(c *config) { c.layoutPath = name }
}

func WithExtension(ext string) Config {
	return func(c *config) { c.extensionName = ext }
}
