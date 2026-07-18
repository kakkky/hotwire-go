package view

import (
	"context"
	"fmt"
	"io"
)

// Partial is a lazy handle to a shared partial together with the data
// that should be executed against it. It captures a template name and
// its data at construction time; the actual html/template execution is
// deferred until Render is called. The same Partial value can therefore
// be handed off to any sink that accepts a Render(ctx, w) writer,
// without paying for the render until the sink actually needs bytes.
//
// Construct one with Renderer.Partial; Partial has no zero value worth
// using directly.
type Partial struct {
	renderer *Renderer
	name     string
	data     map[string]any
}

// Partial returns a Partial bound to a shared partial named name and
// executed against data. name is any {{define "name"}} declared by a
// shared partial file (a file whose base name starts with "_") — the
// same identifier the layout resolves via {{template "name" .}} or
// {{block "name" .}}. {{define}} blocks declared inside a page file
// are not addressable through a Partial; put anything that needs to be
// rendered on its own into a shared partial file.
//
// Templates address entries via the usual dot notation
// ({{ .Title }}). A nil map is accepted and renders as an empty
// context.
//
// Partial is inert: an unknown name is only reported as an error when
// Render is actually called, not at construction time.
func (r *Renderer) Partial(name string, data map[string]any) *Partial {
	return &Partial{renderer: r, name: name, data: data}
}

// Render executes the underlying partial against its data and writes
// the result to w. The context parameter is accepted so Partial fits
// generic Render(ctx, w) writer interfaces; the current implementation
// does not consult it. An unknown partial name returns an error
// without writing to w; a template execution error may leave a partial
// write on w — a caller that needs an all-or-nothing response should
// render into an intermediate buffer first.
func (p *Partial) Render(_ context.Context, w io.Writer) error {
	if p.renderer.base.Lookup(p.name) == nil {
		return fmt.Errorf("view: partial %q not found", p.name)
	}
	if err := p.renderer.base.ExecuteTemplate(w, p.name, p.data); err != nil {
		return fmt.Errorf("view: execute partial %q: %w", p.name, err)
	}
	return nil
}
