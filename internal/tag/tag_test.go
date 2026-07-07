package tag

import (
	"bytes"
	"context"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTag_Render(t *testing.T) {
	tests := []struct {
		name string
		tag  Tag
		want string
	}{
		{
			name: "meta tag",
			tag:  `<meta name="turbo-cache-control" content="no-preview">`,
			want: `<meta name="turbo-cache-control" content="no-preview">`,
		},
		{
			name: "empty tag",
			tag:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.tag.Render(context.Background(), &buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestTag_HTMLTag(t *testing.T) {
	tests := []struct {
		name string
		tag  Tag
		want template.HTML
	}{
		{
			name: "meta tag",
			tag:  `<meta name="turbo-cache-control" content="no-preview">`,
			want: template.HTML(`<meta name="turbo-cache-control" content="no-preview">`),
		},
		{
			name: "empty tag",
			tag:  "",
			want: template.HTML(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tag.HTMLTag())
		})
	}
}
