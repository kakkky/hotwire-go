package turbo

import (
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/elm"
	"github.com/kakkky/hotwire-go/internal/tag"
)

// Attrs is the return type of every attribute helper in this package (for
// example AttrConfirm, AttrMethodDelete). It carries a collection of HTML
// attributes in an engine-neutral form so the same helper works with both
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

// Elm is the return type of every element helper in this package that
// wraps children (for example TurboFrame). It carries a structured tag
// name and its attributes in an engine-neutral form so the same helper
// works with both html/template — where TemplateFuncMap registers
// separate opening and closing funcmap entries (for example
// turboFrame and turboFrameEnd) that template markup is written
// between — and a-h/templ (https://github.com/a-h/templ) via the
// component-call syntax `@turbo.TurboX(...) { ... }`.
type Elm = elm.Elm
