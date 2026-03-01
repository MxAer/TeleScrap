package templates

import (
    "html/template"
    "io"
    "io/fs"
    "path/filepath"
)

var Templates *template.Template

func LoadTemplates(dir string) error {
    var files []string
    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() && filepath.Ext(path) == ".html" {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        return err
    }

    Templates, err = template.ParseFiles(files...)
    return err
}

func Render(w io.Writer, name string, data interface{}) error {
    return Templates.ExecuteTemplate(w, name, data)
}