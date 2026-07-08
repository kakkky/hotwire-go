// Package attrs defines Attrs, an intermediate representation of a
// collection of HTML attributes (key/value pairs) that adapts to both
// html/template and a-h/templ (https://github.com/a-h/templ) without
// duplicating helper implementations for each engine.
//
// Helpers in the caller package return Attrs and let the user pick the
// rendering path that matches their template engine:
//
//   - For html/template, HTMLAttr renders the attributes as
//     template.HTMLAttr — the type html/template treats as a trusted
//     attribute fragment — so the helper is spliced directly into the tag.
//   - For a-h/templ, Attrs is a slice of templ.KeyValue[string, any] and
//     Items satisfies templ.Attributes, so the helper can be used with
//     the spread-attributes syntax.
//
// This package is internal: it is an implementation detail of the
// engine-adapter layer and is not meant to be imported by application code.
package attrs
