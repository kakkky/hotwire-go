package elm

import (
	"bytes"
	"context"
	"html/template"
	"io"

	"github.com/a-h/templ"
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/tag"
)

// Elm is an HTML element carrying a tag name and attributes. It renders
// either the element's opening form (with a-h/templ children when used as
// a Component) or a standalone closing tag, so it can back both the
// templ path (self-contained `<tag>...</tag>` render) and the
// html/template path (open/close pair driven by separate funcmap entries).
type Elm struct {
	// Tag is the element's tag name (e.g. "turbo-frame"). The tag.Tag
	// type is reused for the typed-string ergonomics; the value carried
	// here is a bare tag name rather than a pre-rendered fragment.
	Tag tag.Tag

	// InnerTag optionally names a wrapper element that sits between the
	// outer Tag and its children (for example "template" inside
	// <turbo-stream>, which requires <template>...</template> around
	// the action's content). When non-empty, opening mode emits
	// "<Tag attrs><InnerTag>" — no attrs on the wrapper — and closing
	// mode emits "</InnerTag></Tag>". Leave empty for elements that
	// need no wrapper (the default), such as <turbo-frame>.
	InnerTag tag.Tag

	// IsClosingTag switches HTMLTag into closing-tag mode. When true,
	// HTMLTag emits "</InnerTag></Tag>" (or just "</Tag>" when InnerTag
	// is empty) and Attrs is ignored. Render ignores this field: the
	// templ path always renders the element atomically, so splitting into
	// open/close sides is a concern of the html/template funcmap layer
	// only.
	IsClosingTag bool

	// Attrs are the attributes rendered on the opening tag. Ignored when
	// IsClosingTag is true.
	Attrs attrs.Attrs

	LazyAttrs func() attrs.Attrs
}

// Close returns a copy of e configured to render as its closing tag.
// It lets a single element definition (tag name + attrs) act as both
// the opening and the closing side without duplicating the tag name,
// so callers can write helpers like `Frame(id, attrs...).Close()` for
// the "end" funcmap entry. Attrs are preserved on the returned value
// but ignored by both Render and HTMLTag while IsClosingTag is true.
func (e Elm) Close() Elm {
	e.IsClosingTag = true
	return e
}

// Render satisfies templ.Component. Opening mode emits the full element
// including any children carried on ctx, so `@Elm(...) { children }`
// works with the a-h/templ spread syntax. Closing mode emits "</tag>".
func (e Elm) Render(ctx context.Context, w io.Writer) error {
	a := e.Attrs
	if e.LazyAttrs != nil {
		lazy := e.LazyAttrs()
		a = append(a, lazy...)
	}
	if err := writeOpenTag(ctx, w, e.Tag, a, nil); err != nil {
		return err
	}
	if e.InnerTag != "" {
		if err := writeOpenTag(ctx, w, e.InnerTag, nil, nil); err != nil {
			return err
		}
	}
	if children := templ.GetChildren(ctx); children != nil {
		if err := children.Render(templ.ClearChildren(ctx), w); err != nil {
			return err
		}
	}
	if e.InnerTag != "" {
		if err := writeCloseTag(w, e.InnerTag); err != nil {
			return err
		}
	}
	return writeCloseTag(w, e.Tag)
}

// HTMLTag renders one side of the element for html/template. Opening mode
// returns only "<tag attrs...>" — without children and without the close
// tag — so it can be paired with a separately registered closing-tag
// helper (for example a "turboFrameEnd" funcmap entry), with template
// markup written in between. Closing mode returns "</tag>".
//
// Additional attributes may be passed as variadic template.HTMLAttr
// arguments and are appended after the structured Attrs. This is the
// entry point used by html/template funcmap wrappers (which receive
// pre-rendered template.HTMLAttr fragments from sibling attr helpers)
// so the concatenation lives in Elm rather than in the funcmap.
//
// When LazyAttrs is set, HTMLTag drives it with context.Background —
// use HTMLTagCtx from a funcmap wrapper that closes over a
// request-scoped ctx (see view.Renderer's per-request funcmap
// mechanism) when a ctx-dependent attribute is required.
func (e Elm) HTMLTag(extra ...template.HTMLAttr) template.HTML {
	var buf bytes.Buffer
	switch e.IsClosingTag {
	case true:
		if e.InnerTag != "" {
			_ = writeCloseTag(&buf, e.InnerTag)
		}
		_ = writeCloseTag(&buf, e.Tag)
	case false:
		ctx := context.Background()

		a := e.Attrs
		if e.LazyAttrs != nil {
			lazy := e.LazyAttrs()
			a = append(a, lazy...)
		}
		_ = writeOpenTag(ctx, &buf, e.Tag, a, extra)
		if e.InnerTag != "" {
			_ = writeOpenTag(ctx, &buf, e.InnerTag, nil, nil)
		}
	}
	return template.HTML(buf.String())
}

func writeOpenTag(ctx context.Context, w io.Writer, name tag.Tag, a attrs.Attrs, extra []template.HTMLAttr) error {
	if _, err := io.WriteString(w, "<"+string(name)); err != nil {
		return err
	}
	if err := templ.RenderAttributes(ctx, w, a); err != nil {
		return err
	}
	for _, x := range extra {
		if _, err := io.WriteString(w, string(x)); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, ">")
	return err
}

func writeCloseTag(w io.Writer, name tag.Tag) error {
	_, err := io.WriteString(w, "</"+string(name)+">")
	return err
}
