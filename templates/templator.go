package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
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
	"fileURL": func(path string) string {
		return strings.TrimPrefix(path, "files/")
	},
	"isImage": func(path string) bool {
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif"
	},
	"isVideo": func(path string) bool {
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".mp4" || ext == ".webm" || ext == ".mov"
	},
	"attr": func(key, value string) template.HTMLAttr {
		return template.HTMLAttr(fmt.Sprintf(`%s="%s"`, key, value))
	},
	"json": func(v interface{}) template.JS {
		if v == nil {
			return template.JS("[]")
		}
		b, err := json.Marshal(v)
		if err != nil {
			return template.JS("[]")
		}
		return template.JS(b)
	},
	// Новые функции для работы с путями
	"base": func(path string) string {
		return filepath.Base(path)
	},
	"ext": func(path string) string {
		return filepath.Ext(path)
	},
	"lower": func(s string) string {
		return strings.ToLower(s)
	},
	"upper": func(s string) string {
		return strings.ToUpper(s)
	},
	"uuid": func() string {
		return uuid.New().String()
	},
	"replace": strings.ReplaceAll,
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
