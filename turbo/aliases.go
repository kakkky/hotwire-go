package turbo

import (
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/tag"
)

// Attrs is the return type of every attribute helper in this package (for
// example AttrConfirm, AttrMethodDelete). It carries a single HTML
// attribute in an engine-neutral form so the same helper works with both
// html/template — via the HTMLAttr method registered through
// TemplateFuncMap — and a-h/templ (https://github.com/a-h/templ) via the
// spread-attributes syntax `{ turbo.AttrX(...)... }`.
type Attrs = attrs.Attrs

// Tag is the return type of every element helper in this package (for
// example ScriptImport, MetaVisitControlReload). It carries a
// pre-rendered HTML fragment in an engine-neutral form so the same helper
// works with both html/template — via the HTMLTag method registered
// through TemplateFuncMap — and a-h/templ
// (https://github.com/a-h/templ) via the component-call syntax
// `@turbo.TagX()`.
type Tag = tag.Tag
