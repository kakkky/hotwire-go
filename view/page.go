package view

import (
	"context"
	"fmt"
	"io"
	"maps"
)

// Page is a lazy handle to a page template together with the data that
// should be executed against it. It captures the page name and its data
// at construction time; the actual html/template execution — including
// composing the page through the layout — is deferred until Render is
// called. The same Page value can therefore be handed off to any sink
// that accepts a Render(ctx, w) writer.
//
// Construct one with Renderer.Page; Page has no zero value worth using
// directly.
type Page struct {
	renderer *Renderer
	name     string
	data     map[string]any
}

// Page returns a Page bound to the page template named name, executed
// against data. name is the path of the page relative to the view
// directory with the file extension stripped (for example the file
// "sub/edit.gotmpl" is addressed as "sub/edit"), the same key Renderer
// registered when New parsed the directory.
//
// Templates address entries via the usual dot notation
// ({{ .Title }}). A nil map is accepted; the template then sees only
// .Ctx. The key "Ctx" is reserved for Render's ctx injection;
// caller-provided values under that key are overwritten.
//
// Page is inert: an unknown name is only reported as an error when
// Render is actually called, not at construction time.
func (r *Renderer) Page(name string, data map[string]any) *Page {
	return &Page{renderer: r, name: name, data: data}
}

// Render executes the page through the layout against its data and
// writes the result to w. Any {{define}} blocks provided by the page
// override the layout's placeholders in the usual html/template way.
//
// ctx is exposed to the template under the reserved key .Ctx.
// The caller's data map is not mutated: a fresh copy is built per call.
//
// An unknown page name returns an error without writing to w; a
// template execution error may leave a partial write on w — a caller
// that needs an all-or-nothing response should render into an
// intermediate buffer first.
func (p *Page) Render(ctx context.Context, w io.Writer) error {
	t, ok := p.renderer.pages[p.name]
	if !ok {
		return fmt.Errorf("view: page %q not found", p.name)
	}
	data := make(map[string]any, len(p.data)+1)
	maps.Copy(data, p.data)
	data["Ctx"] = ctx
	if err := t.ExecuteTemplate(w, p.renderer.layoutExecName, data); err != nil {
		return fmt.Errorf("view: execute page %q: %w", p.name, err)
	}
	return nil
}
