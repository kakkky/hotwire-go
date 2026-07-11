// Package view provides a thin html/template based renderer that
// composes a single layout, shared partials, and per-page templates
// and hands out lazy Page and Partial values that render themselves to
// a caller-supplied io.Writer. Callers own where the bytes go and any
// HTTP concerns around them (status code, Content-Type, buffering).
//
// A view directory is expected to follow three simple conventions:
//   - exactly one layout file (default name "layout", default extension
//     ".gotmpl") which acts as the entry point of every page;
//   - files whose base name starts with "_" are treated as partials and
//     are parsed into every page's template set;
//   - every other file with the configured extension is a page, and is
//     addressable by its path relative to the view directory with the
//     extension stripped (for example "sub/page.gotmpl" -> "sub/page").
//
// Both the layout path and the file extension can be customized via the
// Configs returned by WithLayout and WithExtension. Template helper
// functions can be registered with WithFuncs; they must be provided
// before parsing since html/template resolves function names at parse
// time.
//
// Typical use is to build a Renderer once at start-up and then, per
// request, call Renderer.Page or Renderer.Partial and Render the
// returned value to the destination writer.
package view
