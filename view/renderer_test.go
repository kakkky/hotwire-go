package view

import (
	"bytes"
	"embed"
	"html/template"
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
		{
			name: "files with non-matching extension are skipped",
			dir:  "testdata/valid/mixed_extensions",
			wantPages: map[string]string{
				"home": "<body>mixed</body>\n",
			},
		},
		{
			name: "WithFuncs with no funcs is a no-op",
			dir:  "testdata/valid/mixed_extensions",
			cfgs: []Config{WithFuncs()},
			wantPages: map[string]string{
				"home": "<body>mixed</body>\n",
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
		{
			name:       "walk error when the directory does not exist",
			dir:        "testdata/does_not_exist",
			wantErrMsg: "walk",
		},
		{
			name:       "layout with malformed template body fails to parse",
			dir:        "testdata/invalid/bad_layout",
			wantErrMsg: "parse layout/partials",
		},
		{
			name:       "page with malformed template body fails to parse",
			dir:        "testdata/invalid/bad_page",
			wantErrMsg: "parse page",
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

