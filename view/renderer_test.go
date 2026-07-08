package view

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed all:testdata
var testdataFS embed.FS

var testFuncs = template.FuncMap{"upper": strings.ToUpper}

func TestNew(t *testing.T) {
	data := map[string]string{"Name": "Test"}

	tests := []struct {
		name      string
		dir       string
		cfgs      []Config
		wantPages map[string]string // page name -> full expected rendered output
	}{
		{
			name: "default settings",
			dir:  "testdata/valid/ok",
			cfgs: []Config{WithFuncs(testFuncs)},
			wantPages: map[string]string{
				"home": strings.Join([]string{
					`<html>`,
					`  <head>`,
					`    <title>Home</title>`,
					`  </head>`,
					`  <body>`,
					`    <div class="brand">HI</div>`,
					`    SHARED`,
					`    <main>Hello Test</main>`,
					`  </body>`,
					`</html>`,
				}, "\n") + "\n",
				"sub/page": strings.Join([]string{
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
		},
		{
			name: "trailing slash normalized by path.Clean",
			dir:  "testdata/valid/ok/",
			cfgs: []Config{WithFuncs(testFuncs)},
			wantPages: map[string]string{
				"home": strings.Join([]string{
					`<html>`,
					`  <head>`,
					`    <title>Home</title>`,
					`  </head>`,
					`  <body>`,
					`    <div class="brand">HI</div>`,
					`    SHARED`,
					`    <main>Hello Test</main>`,
					`  </body>`,
					`</html>`,
				}, "\n") + "\n",
				"sub/page": strings.Join([]string{
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
		},
		{
			name: "layout at nested path",
			dir:  "testdata/valid/nested_layout",
			cfgs: []Config{WithLayout("layouts/application")},
			wantPages: map[string]string{
				"home": "[APP]nested layout page[/APP]\n",
			},
		},
		{
			name: "custom extension",
			dir:  "testdata/valid/custom_ext",
			cfgs: []Config{WithExtension(".tpl")},
			wantPages: map[string]string{
				"home": "<body>custom ext page</body>\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(testdataFS, tt.dir, tt.cfgs...)
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Len(t, r.pages, len(tt.wantPages))

			for page, want := range tt.wantPages {
				tmpl, ok := r.pages[page]
				require.Truef(t, ok, "page %q not in map", page)

				var buf bytes.Buffer
				err := tmpl.ExecuteTemplate(&buf, r.layoutExecName, data)
				require.NoErrorf(t, err, "execute page %q", page)

				assert.Equalf(t, want, buf.String(), "page %q body mismatch", page)
			}
		})
	}
}

func TestNew_Error(t *testing.T) {
	tests := []struct {
		name       string
		dir        string
		cfgs       []Config
		wantErrMsg string
	}{
		{
			name:       "layout file missing",
			dir:        "testdata/invalid/no_layout",
			wantErrMsg: "layout",
		},
		{
			name:       "no pages",
			dir:        "testdata/invalid/no_pages",
			wantErrMsg: "no page templates",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(testdataFS, tt.dir, tt.cfgs...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestRender(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name       string
		status     int
		page       string
		data       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "custom status is respected",
			status:     http.StatusUnprocessableEntity,
			page:       "home",
			data:       map[string]string{"Name": "Bob"},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   "Hello Bob",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := r.Render(w, tt.status, tt.page, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

func TestRender_Error(t *testing.T) {
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
			w := httptest.NewRecorder()
			err := r.Render(w, http.StatusOK, tt.page, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestRenderPartial(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name       string
		status     int
		partial    string
		data       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "shared partial rendered without layout",
			status:     http.StatusOK,
			partial:    "shared",
			data:       nil,
			wantStatus: http.StatusOK,
			wantBody:   "SHARED",
		},
		{
			name:       "partial receives data",
			status:     http.StatusOK,
			partial:    "greet",
			data:       map[string]string{"Name": "Bob"},
			wantStatus: http.StatusOK,
			wantBody:   "Hello Bob",
		},
		{
			name:       "custom status is respected",
			status:     http.StatusUnprocessableEntity,
			partial:    "shared",
			data:       nil,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   "SHARED",
		},
		{
			name:       "partial in a nested directory is reachable",
			status:     http.StatusOK,
			partial:    "local",
			data:       nil,
			wantStatus: http.StatusOK,
			wantBody:   "LOCAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := r.RenderPartial(w, tt.status, tt.partial, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
			assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		})
	}
}

func TestRenderPartial_Error(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name       string
		partial    string
		wantErrMsg string
	}{
		{
			name:       "unknown partial",
			partial:    "does_not_exist",
			wantErrMsg: `partial "does_not_exist" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := r.RenderPartial(w, http.StatusOK, tt.partial, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
			assert.Equal(t, http.StatusOK, w.Code) // untouched (recorder default)
			assert.Empty(t, w.Body.String())
		})
	}
}
