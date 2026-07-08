package attrs

import (
	"html/template"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func TestAttrs_Items(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		want  []templ.KeyValue[string, any]
	}{
		{
			name:  "string value",
			attrs: Attrs{{Key: "data-turbo-method", Value: "delete"}},
			want:  []templ.KeyValue[string, any]{{Key: "data-turbo-method", Value: "delete"}},
		},
		{
			name:  "bool value",
			attrs: Attrs{{Key: "data-turbo-preload", Value: true}},
			want:  []templ.KeyValue[string, any]{{Key: "data-turbo-preload", Value: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.attrs.Items())
		})
	}
}

func TestAttrs_HTMLAttr(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		want  template.HTMLAttr
	}{
		{
			name:  "string value",
			attrs: Attrs{{Key: "data-turbo-method", Value: "delete"}},
			want:  ` data-turbo-method="delete"`,
		},
		{
			name:  "bool true renders bare name",
			attrs: Attrs{{Key: "data-turbo-preload", Value: true}},
			want:  ` data-turbo-preload`,
		},
		{
			name:  "bool false omits the attribute",
			attrs: Attrs{{Key: "data-turbo-preload", Value: false}},
			want:  "",
		},
		{
			name:  "escape special characters in value",
			attrs: Attrs{{Key: "data-turbo-confirm", Value: `<"&>`}},
			want:  ` data-turbo-confirm="&lt;&#34;&amp;&gt;"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.attrs.HTMLAttr())
		})
	}
}
