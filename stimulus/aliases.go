package stimulus

import (
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/tag"
)

type Tag = tag.Tag

// Attrs is the return type of every attribute helper in this package
// (for example AttrControllers). It carries a collection of HTML
// attributes in an engine-neutral form so the same helper works with
// both html/template — via the HTMLAttr method registered through
// TemplateFuncMap — and a-h/templ (https://github.com/a-h/templ) via
// the spread-attributes syntax `{ stimulus.AttrX(...)... }`.
type Attrs = attrs.Attrs
