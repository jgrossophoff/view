// Package view can be used to parse all templates inside a folder with
// a matching file extension.
//
// It allows safe reparsing of all files on each template execution to remove
// the need for server restarts during development.
package view

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Views struct {
	Reload bool

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
	if v.Reload {
		err := v.ParseTemplates()
		if err != nil {
			return err
		}
	}
	return v.tmpls.ExecuteTemplate(w, name, data)
}

// Execute exposes the same API as {html,text}/template.Template.
func (v *Views) Execute(w io.Writer, data interface{}) error {
	if v.Reload {
		err := v.ParseTemplates()
		if err != nil {
			return err
		}
	}
	return v.tmpls.Execute(w, data)
}

// ParseTemplates parses all template files with the matching extension inside
// the folder path.
// Save for concurrent use.
func (v *Views) ParseTemplates() error {
	var files []string

	err := filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Err while walking template dir: ", err)
		}
		if !info.IsDir() && filepath.Ext(path) == v.tmplExt {
			files = append(files, path)
		}

		return err
	})
	if err != nil {
		return err
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.tmpls = t
	return nil
}
