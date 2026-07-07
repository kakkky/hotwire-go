// Package tag defines Tag, an intermediate representation of a
// pre-rendered HTML fragment (a full element such as <script>, <meta>, or
// <link>) that adapts to both html/template and a-h/templ
// (https://github.com/a-h/templ) without duplicating helper implementations
// for each engine.
//
// Helpers in the caller package return Tag and let the user pick the
// rendering path that matches their template engine:
//
//   - For html/template, HTMLTag returns template.HTML — the type
//     html/template treats as a trusted HTML fragment — so the helper is
//     spliced directly into the document.
//   - For a-h/templ, Tag implements templ.Component via Render, so the
//     helper can be used with the component-call syntax.
//
// Because Tag values are produced only by the caller package from fixed
// strings (or from inputs escaped by the caller), the fragments carried
// by Tag are already safe to emit unescaped through both template.HTML
// and templ.Component.
//
// This package is internal: it is an implementation detail of the
// engine-adapter layer and is not meant to be imported by application code.
package tag
