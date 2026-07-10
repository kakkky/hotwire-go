package elm

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"testing"

	"github.com/a-h/templ"
	"github.com/kakkky/hotwire-go/internal/attrs"
	"github.com/kakkky/hotwire-go/internal/tag"
	"github.com/stretchr/testify/assert"
)

func TestElm_Render(t *testing.T) {
	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<p>hi</p>")
		return err
	})

	tests := []struct {
		name  string
		elm   Elm
		child templ.Component
		want  string
	}{
		{
			name: "opening with no attrs and no children",
			elm:  Elm{Tag: tag.Tag("turbo-frame")},
			want: `<turbo-frame></turbo-frame>`,
		},
		{
			name: "opening with attrs and no children",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: `<turbo-frame id="msg"></turbo-frame>`,
		},
		{
			name: "opening with children from ctx",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			child: children,
			want:  `<turbo-frame id="msg"><p>hi</p></turbo-frame>`,
		},
		{
			name: "InnerTag wraps empty content when no children",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
			},
			want: `<turbo-stream><template></template></turbo-stream>`,
		},
		{
			name: "InnerTag wraps children between outer and close",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
				Attrs: attrs.Attrs{
					{Key: "action", Value: "append"},
					{Key: "target", Value: "msg"},
				},
			},
			child: children,
			want:  `<turbo-stream action="append" target="msg"><template><p>hi</p></template></turbo-stream>`,
		},
		{
			name: "InnerTag carries no attrs of its own",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
				Attrs:    attrs.Attrs{{Key: "action", Value: "remove"}},
			},
			want: `<turbo-stream action="remove"><template></template></turbo-stream>`,
		},
		{
			name: "Render ignores IsClosingTag and renders atomically",
			elm: Elm{
				Tag:          "turbo-frame",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: `<turbo-frame id="msg"></turbo-frame>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.child != nil {
				ctx = templ.WithChildren(ctx, tt.child)
			}
			var buf bytes.Buffer
			err := tt.elm.Render(ctx, &buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestElm_Close(t *testing.T) {
	tests := []struct {
		name string
		elm  Elm
		want Elm
	}{
		{
			name: "flips IsClosingTag on and preserves other fields",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: Elm{
				Tag:          "turbo-frame",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "id", Value: "msg"}},
			},
		},
		{
			name: "already-closing stays closing",
			elm:  Elm{Tag: "turbo-frame", IsClosingTag: true},
			want: Elm{Tag: "turbo-frame", IsClosingTag: true},
		},
		{
			name: "does not mutate the receiver",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: Elm{
				Tag:          "turbo-frame",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "id", Value: "msg"}},
			},
		},
		{
			name: "preserves InnerTag",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
				Attrs:    attrs.Attrs{{Key: "action", Value: "append"}},
			},
			want: Elm{
				Tag:          "turbo-stream",
				InnerTag:     "template",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "action", Value: "append"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.elm
			got := tt.elm.Close()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, original, tt.elm, "receiver should not be mutated")
		})
	}
}

func TestElm_HTMLTag(t *testing.T) {
	tests := []struct {
		name  string
		elm   Elm
		extra []template.HTMLAttr
		want  template.HTML
	}{
		{
			name: "opening with no attrs",
			elm:  Elm{Tag: "turbo-frame"},
			want: template.HTML(`<turbo-frame>`),
		},
		{
			name: "opening with attrs",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: template.HTML(`<turbo-frame id="msg">`),
		},
		{
			name: "opening appends extra HTMLAttrs after structured attrs",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			extra: []template.HTMLAttr{` src="/messages"`, ` loading="lazy"`},
			want:  template.HTML(`<turbo-frame id="msg" src="/messages" loading="lazy">`),
		},
		{
			name:  "opening with only extra HTMLAttrs and no structured attrs",
			elm:   Elm{Tag: "turbo-frame"},
			extra: []template.HTMLAttr{` src="/messages"`},
			want:  template.HTML(`<turbo-frame src="/messages">`),
		},
		{
			name: "closing tag",
			elm:  Elm{Tag: "turbo-frame", IsClosingTag: true},
			want: template.HTML(`</turbo-frame>`),
		},
		{
			name: "closing tag ignores attrs",
			elm: Elm{
				Tag:          "turbo-frame",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			want: template.HTML(`</turbo-frame>`),
		},
		{
			name:  "closing tag ignores extra HTMLAttrs",
			elm:   Elm{Tag: "turbo-frame", IsClosingTag: true},
			extra: []template.HTMLAttr{` src="/messages"`},
			want:  template.HTML(`</turbo-frame>`),
		},
		{
			name: "opening emits InnerTag after outer open",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
				Attrs: attrs.Attrs{
					{Key: "action", Value: "append"},
					{Key: "target", Value: "msg"},
				},
			},
			want: template.HTML(`<turbo-stream action="append" target="msg"><template>`),
		},
		{
			name: "opening puts extras on outer tag only, not InnerTag",
			elm: Elm{
				Tag:      "turbo-stream",
				InnerTag: "template",
				Attrs:    attrs.Attrs{{Key: "action", Value: "append"}},
			},
			extra: []template.HTMLAttr{` target="msg"`},
			want:  template.HTML(`<turbo-stream action="append" target="msg"><template>`),
		},
		{
			name: "closing emits InnerTag before outer close",
			elm: Elm{
				Tag:          "turbo-stream",
				InnerTag:     "template",
				IsClosingTag: true,
			},
			want: template.HTML(`</template></turbo-stream>`),
		},
		{
			name: "closing with InnerTag ignores attrs and extras",
			elm: Elm{
				Tag:          "turbo-stream",
				InnerTag:     "template",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "action", Value: "append"}},
			},
			extra: []template.HTMLAttr{` target="msg"`},
			want:  template.HTML(`</template></turbo-stream>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.elm.HTMLTag(tt.extra...))
		})
	}
}
