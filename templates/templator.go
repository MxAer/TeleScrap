package templates

import (
    "html/template"
    "io"
    "io/fs"
    "path/filepath"
    "time"
)

type ChatItem struct {
    ID          string
    Name        string
    Avatar      string
    Subscribers string
    Description string
    IsGroup     bool
}

var funcMap = template.FuncMap{
    "formatDate": func(t time.Time) string {
        if t.IsZero() {
            return ""
        }
        return t.Format("02.01.2006 15:04")
    },
    "eq": func(a, b interface{}) bool {
        return a == b
    },
    "mul": func(a interface{}, b interface{}) float64 {
        var v1, v2 float64
        switch aa := a.(type) {
        case int:
            v1 = float64(aa)
        case int64:
            v1 = float64(aa)
        case float64:
            v1 = aa
        }
        switch bb := b.(type) {
        case int:
            v2 = float64(bb)
        case int64:
            v2 = float64(bb)
        case float64:
            v2 = bb
        }
        return v1 * v2
    },
    "slice": func(s string, start, end int) string {
        if s == "" {
            return ""
        }
        if start < 0 {
            start = 0
        }
        if end > len(s) {
            end = len(s)
        }
        if start > end {
            return ""
        }
        return s[start:end]
    },
}

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

    Templates, err = template.New("").Funcs(funcMap).ParseFiles(files...)
    return err
}

func Render(w io.Writer, name string, data interface{}) error {
    return Templates.ExecuteTemplate(w, "layout.html", data)
}