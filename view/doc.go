// Package view provides a thin html/template based renderer that
// composes a single layout, shared partials, and per-page templates
// into HTTP responses.
//
// A view directory is expected to follow three simple conventions:
//   - exactly one layout file (default name "layout", default extension
//     ".gotmpl") which acts as the entry point of every response;
//   - files whose base name starts with "_" are treated as partials and
//     are parsed into every page's template set;
//   - every other file with the configured extension is a page, and is
//     addressable by its path relative to the view directory with the
//     extension stripped (for example "sub/page.gotmpl" -> "sub/page").
//
// Both the layout path and the file extension can be customized via the
// Config options returned by WithLayout and WithExtension. Template
// helper functions can be registered with WithFuncs; they must be
// provided before parsing since html/template resolves function names
// at parse time.
package view
