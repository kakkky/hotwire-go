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

type Renderer struct {
	pages          map[string]*template.Template
	layoutExecName string
}

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
		pages:          pages,
		layoutExecName: path.Base(c.layoutPath) + c.extensionName,
	}, nil
}

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
