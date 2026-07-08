// Package elm defines Elm, an intermediate representation of an HTML
// element (a tag name plus its attributes) that adapts to both
// html/template and a-h/templ (https://github.com/a-h/templ) without
// duplicating helper implementations for each engine.
//
// Unlike tag.Tag — which carries a fully rendered fragment — Elm keeps
// the tag name and attributes structured so it can render either the
// opening side (with children when used as a templ component) or the
// closing side of the same element. This lets a single helper back both
// the templ path (self-contained "<tag>...</tag>" render via the
// component-call syntax) and the html/template path, where the opening
// and closing tags are registered as separate funcmap entries and
// template markup is written between them.
//
// Helpers in the caller package return Elm and let the user pick the
// rendering path that matches their template engine:
//
//   - For html/template, HTMLTag returns template.HTML — the type
//     html/template treats as a trusted HTML fragment — so the helper is
//     spliced directly into the document.
//   - For a-h/templ, Elm implements templ.Component via Render, so the
//     helper can be used with the component-call syntax and receive
//     children through the spread syntax.
//
// This package is internal: it is an implementation detail of the
// engine-adapter layer and is not meant to be imported by application code.
package elm
