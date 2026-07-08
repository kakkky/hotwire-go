package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const (
	partialPrefix        = "_"
	defaultLayoutPath    = "layout"
	defaultExtensionName = ".gotmpl"
)

// Renderer holds a set of page templates that have been pre-parsed
// together with a common layout and shared partials. It is safe for
// concurrent use once constructed by New, and is intended to be built
// once at start-up and reused for the lifetime of the process.
type Renderer struct {
	base           *template.Template
	pages          map[string]*template.Template
	layoutExecName string
}

// New walks dir on fsys and constructs a Renderer from the templates it
// finds. Files are classified by name:
//
//   - the layout file (default "layout.gotmpl", relative to dir) is
//     used as the entry point for every page;
//   - files whose base name starts with "_" are parsed as partials
//     and made available to every page;
//   - every other file with the configured extension becomes a page,
//     keyed by its path relative to dir with the extension trimmed.
//
// The layout path and file extension can be overridden with WithLayout
// and WithExtension. Template helper functions must be registered with
// WithFuncs, because html/template resolves function names at parse
// time.
//
// An error is returned if the walk fails, if the layout file is
// missing, or if no page templates are found under dir.
func New(fsys fs.FS, dir string, cfgs ...Config) (*Renderer, error) {
	c := &config{
		layoutPath:    defaultLayoutPath,
		extensionName: defaultExtensionName,
	}
	for _, cfg := range cfgs {
		cfg(c)
	}

	dir = path.Clean(dir)
	wantLayoutPath := path.Join(dir, c.layoutPath+c.extensionName)

	var layoutFile string
	var partialFiles, pageFiles []string

	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path.Ext(p) != c.extensionName {
			return nil
		}
		switch {
		case p == wantLayoutPath:
			layoutFile = p
		case strings.HasPrefix(path.Base(p), partialPrefix):
			partialFiles = append(partialFiles, p)
		default:
			pageFiles = append(pageFiles, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: walk %q: %w", dir, err)
	}
	if layoutFile == "" {
		return nil, fmt.Errorf("view: layout %q not found", wantLayoutPath)
	}
	if len(pageFiles) == 0 {
		return nil, fmt.Errorf("view: no page templates found in %s", dir)
	}

	baseTemplate := template.New("")
	if c.funcs != nil {
		baseTemplate = baseTemplate.Funcs(c.funcs)
	}
	baseFiles := append([]string{layoutFile}, partialFiles...)
	if _, err := baseTemplate.ParseFS(fsys, baseFiles...); err != nil {
		return nil, fmt.Errorf("view: parse layout/partials: %w", err)
	}

	pages := make(map[string]*template.Template, len(pageFiles))
	for _, pageFile := range pageFiles {
		t, err := baseTemplate.Clone()
		if err != nil {
			return nil, fmt.Errorf("view: clone for %s: %w", pageFile, err)
		}
		if _, err := t.ParseFS(fsys, pageFile); err != nil {
			return nil, fmt.Errorf("view: parse page %s: %w", pageFile, err)
		}
		rel := strings.TrimPrefix(pageFile, dir+"/")
		page := strings.TrimSuffix(rel, c.extensionName)
		pages[page] = t
	}

	return &Renderer{
		base:           baseTemplate,
		pages:          pages,
		layoutExecName: path.Base(c.layoutPath) + c.extensionName,
	}, nil
}

// Render executes the template registered under page against data and
// writes the result to w. The layout is used as the entry point of the
// execution, so any {{define}} blocks provided by the page override the
// layout's placeholders in the usual html/template way.
//
// Before writing the body, Render sets Content-Type to
// "text/html; charset=utf-8" and writes the given HTTP status code.
// The rendered output is buffered in memory first, so template
// execution errors are returned without ever writing a partial
// response. An unknown page returns an error without touching w.
func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data any) error {
	t, ok := r.pages[page]
	if !ok {
		return fmt.Errorf("view: page %q not found", page)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, r.layoutExecName, data); err != nil {
		return fmt.Errorf("view: execute %q: %w", page, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}

// RenderPartial executes a single shared partial against data and
// writes the result to w without going through the layout. name is any
// {{define "name"}} declared by a shared partial file (a file whose
// base name starts with "_") — the same identifier Render resolves
// when the layout references it via {{template "name" .}} or
// {{block "name" .}}. {{define}} blocks declared inside a page file
// are not addressable through this method; put anything that needs to
// be rendered on its own into a shared partial file.
//
// It is intended for Turbo-Frame responses (see turbo.IsFrameRequest):
// the same {{define}} that supplies a portion of the full page can be
// returned on its own so Turbo swaps just the matching <turbo-frame>,
// without paying for the surrounding layout on the wire.
//
// Response-writing semantics match Render: Content-Type is set to
// "text/html; charset=utf-8", the body is buffered before writing so
// execution errors do not produce a partial response, and an unknown
// name returns an error without touching w.
func (r *Renderer) RenderPartial(w http.ResponseWriter, status int, name string, data any) error {
	if r.base.Lookup(name) == nil {
		return fmt.Errorf("view: partial %q not found", name)
	}
	var buf bytes.Buffer
	if err := r.base.ExecuteTemplate(&buf, name, data); err != nil {
		return fmt.Errorf("view: execute %q: %w", name, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}
