package attrs

import (
	"bytes"
	"context"
	"html/template"

	"github.com/a-h/templ"
)

// Attrs is a single HTML attribute (key/value pair) in an engine-neutral
// form. It is defined as templ.KeyValue[string, any] so a-h/templ can use
// it directly, and it also exposes an html/template rendering via
// HTMLAttr. See the package doc for how the two paths are wired up.
type Attrs templ.KeyValue[string, any]

// Items satisfies the templ.Attributes interface by returning the receiver
// as a single-element slice. It is what allows Attrs to be used with the
// a-h/templ spread-attributes syntax `<... { attrHelper()... }>`.
func (a Attrs) Items() []templ.KeyValue[string, any] {
	return []templ.KeyValue[string, any]{templ.KeyValue[string, any](a)}
}

// HTMLAttr renders the attribute for html/template. It delegates to
// templ.RenderAttributes so the escaping and formatting rules match the
// a-h/templ path (boolean-valued attributes render as a bare name, string
// values are attribute-quoted and escaped) and returns the result as
// template.HTMLAttr so it is spliced into a tag without further escaping.
func (a Attrs) HTMLAttr() template.HTMLAttr {
	var buf bytes.Buffer
	_ = templ.RenderAttributes(context.Background(), &buf, a)
	return template.HTMLAttr(buf.String())
}
