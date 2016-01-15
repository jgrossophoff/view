// Package view can be used to parse all templates inside a folder with
// a matching file extension.
//
// Templates are named by their path relative to the template folder without the
// file extension. A template file inside templates/foo/bar/index.tmpl for example can be
// executed by the name foo/bar/index.
//
// That way, blocks introduced in Go 1.6 can be used to achieve template extension with ease.
//
// It allows safe reparsing of all files on each template execution to remove
// the need for server restarts during development.
package view

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Views struct {
	reload bool

	path    string
	tmplExt string
	tmpls   *template.Template
	mu      *sync.Mutex
}

// NewViews will parse all files initially. Returns parse errors.
func NewViews(path, tmplExt string, reload bool) (*Views, error) {
	v := &Views{
		reload,
		path,
		tmplExt,
		nil,
		new(sync.Mutex),
	}

	err := v.ParseTemplates()
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ExecuteTemplate exposes the same API as {html,text}/template.Template.
func (v *Views) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	if v.reload {
		err := v.ParseTemplates()
		if err != nil {
			return err
		}
	}
	return v.tmpls.ExecuteTemplate(w, name, data)
}

// Execute exposes the same API as {html,text}/template.Template.
func (v *Views) Execute(w io.Writer, data interface{}) error {
	if v.reload {
		err := v.ParseTemplates()
		if err != nil {
			return err
		}
	}
	return v.tmpls.Execute(w, data)
}

// Reload defines if all templates will be reparsed on execution.
// Safe for concurrent access.
func (v *Views) Reload(f bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.reload = f
}

// ParseTemplates parses all template files with the matching extension inside
// the folder path.
// Save for concurrent use.
func (v *Views) ParseTemplates() error {
	t := template.New("all")

	err := filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Err while walking template dir: ", err)
		}
		if !info.IsDir() && filepath.Ext(path) == v.tmplExt {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			t, err = t.New(v.TemplateName(path)).Parse(string(b))
			if err != nil {
				return err
			}
		}

		return err
	})
	if err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.tmpls = t
	return nil
}

// TemplateName returns the file name relative to the template folder
// without the template file extension or leading slashes.
func (v *Views) TemplateName(s string) string {
	// Folder name without dots or (back-)slashes
	cf := strings.Trim(v.path, "./\\")
	// Remove parent folder and leading (back-)slashes
	s = strings.TrimLeft(strings.Replace(s, cf, "", 1), "/\\")

	// Return without file ext
	return s[:len(s)-len(v.tmplExt)]
}
