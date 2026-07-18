package view

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPage_Render(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		ctx      context.Context
		page     string
		data     map[string]any
		wantBody string
	}{
		{
			name: "home page rendered through the layout",
			dir:  "testdata/valid/ok",
			ctx:  context.Background(),
			page: "home",
			data: map[string]any{"Name": "Bob"},
			wantBody: strings.Join([]string{
				`<html>`,
				`  <head>`,
				`    <title>Home</title>`,
				`  </head>`,
				`  <body>`,
				`    <div class="brand">HI</div>`,
				`    SHARED`,
				`    <main>Hello Bob</main>`,
				`  </body>`,
				`</html>`,
			}, "\n") + "\n",
		},
		{
			name: "nested page rendered through the layout",
			dir:  "testdata/valid/ok",
			ctx:  context.Background(),
			page: "sub/page",
			data: map[string]any{"Name": "Test"},
			wantBody: strings.Join([]string{
				`<html>`,
				`  <head>`,
				`    <title>Default</title>`,
				`  </head>`,
				`  <body>`,
				`    <div class="brand">HI</div>`,
				`    SHARED`,
				`    <main>Sub page LOCAL</main>`,
				`  </body>`,
				`</html>`,
			}, "\n") + "\n",
		},
		{
			name:     ".Ctx is exposed to the template",
			dir:      "testdata/valid/ctx",
			ctx:      context.Background(),
			page:     "home",
			data:     nil,
			wantBody: "has-ctx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(testdataFS, tt.dir, WithFuncs(testFuncs))
			require.NoError(t, err)
			p := r.Page(tt.page, tt.data)
			var buf bytes.Buffer
			err = p.Render(tt.ctx, &buf)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, buf.String())
			assert.NotContains(t, tt.data, "Ctx", "Render must not mutate caller's data map")
		})
	}
}

func TestPage_Render_Error(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name       string
		page       string
		wantErrMsg string
	}{
		{
			name:       "unknown page",
			page:       "does_not_exist",
			wantErrMsg: `page "does_not_exist" not found`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := r.Page(tt.page, nil)
			var buf bytes.Buffer
			err := p.Render(context.Background(), &buf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
			assert.Empty(t, buf.String())
		})
	}
}
