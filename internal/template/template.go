package template

import (
	"embed"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate(views *embed.FS) *Template {
	return &Template{templates: template.Must(template.ParseFS(views, "views/*.html"))}
}
