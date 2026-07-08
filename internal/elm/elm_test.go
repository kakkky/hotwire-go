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
		name string
		elm  Elm
		ctx  context.Context
		want string
	}{
		{
			name: "opening with no attrs and no children",
			elm:  Elm{Tag: tag.Tag("turbo-frame")},
			ctx:  context.Background(),
			want: `<turbo-frame></turbo-frame>`,
		},
		{
			name: "opening with attrs and no children",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			ctx:  context.Background(),
			want: `<turbo-frame id="msg"></turbo-frame>`,
		},
		{
			name: "opening with children from ctx",
			elm: Elm{
				Tag:   "turbo-frame",
				Attrs: attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			ctx:  templ.WithChildren(context.Background(), children),
			want: `<turbo-frame id="msg"><p>hi</p></turbo-frame>`,
		},
		{
			name: "closing tag emits only end",
			elm:  Elm{Tag: "turbo-frame", IsClosingTag: true},
			ctx:  context.Background(),
			want: `</turbo-frame>`,
		},
		{
			name: "closing tag ignores attrs",
			elm: Elm{
				Tag:          "turbo-frame",
				IsClosingTag: true,
				Attrs:        attrs.Attrs{{Key: "id", Value: "msg"}},
			},
			ctx:  context.Background(),
			want: `</turbo-frame>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.elm.Render(tt.ctx, &buf)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.elm.HTMLTag(tt.extra...))
		})
	}
}
