package tag

import (
	"context"
	"html/template"
	"io"
)

// Tag is a pre-rendered HTML fragment in an engine-neutral form. The
// underlying string is emitted verbatim by both rendering paths; see the
// package doc for how the two paths are wired up and why unescaped
// emission is safe.
type Tag string

// Render satisfies templ.Component by writing the fragment to w as-is.
// It is what allows Tag to be used with the a-h/templ component-call
// syntax `@tagHelper()`.
func (t Tag) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(t))
	return err
}

// HTMLTag renders the fragment for html/template as template.HTML, so it
// is spliced into a template without further escaping.
func (t Tag) HTMLTag() template.HTML {
	return template.HTML(t)
}
