package main

import (
	"embed"
	"fmt"
	gotemplate "html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shortcuts/codes/pkg/markdown"
	"github.com/shortcuts/codes/pkg/template"
	"golang.org/x/time/rate"
)

type route struct {
	path     string
	filename string
}

var routes = []route{
	{
		path:     "",
		filename: "home.md",
	},
	{
		path:     "resume",
		filename: "resume.md",
	},
	{
		path:     "links",
		filename: "links.md",
	},
}

//go:embed views/*.md views/*.html
var views embed.FS

type server struct {
	http   *echo.Echo
	parser *markdown.MarkdownParser
	navbar *gotemplate.HTML
}

func (s *server) close() error {
	if err := s.http.Close(); err != nil {
		return err
	}

	return nil
}

func errorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	c.Logger().Error(err)

	errorPage := fmt.Sprintf("%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
}

func newServer() server {
	e := echo.New()

	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(20),
		)),
	)

	e.Static("assets", "cmd/css")
	e.Static("views", "cmd/views")

	e.HTTPErrorHandler = errorHandler

	e.Renderer = template.NewTemplate(&views)

	parser := markdown.NewParser(&views)

	navbar, err := parser.ToHTML("views/navbar.md")
	if err != nil {
		panic(err)
	}

	return server{
		http:   e,
		parser: &parser,
		navbar: navbar,
	}
}

func (s *server) registerRoute(route route) error {
	content, err := s.parser.ToHTML(fmt.Sprintf("views/%s", route.filename))
	if err != nil {
		return err
	}

	s.http.GET(fmt.Sprintf("/%s", route.path), func(c echo.Context) error {
		return c.Render(http.StatusOK, "layout", map[string]any{
			"Filename": route.filename,
			"Navbar":   s.navbar,
			"Content":  content,
		})
	})

	return nil
}

func main() {
	s := newServer()
	defer s.close()

	for _, route := range routes {
		if err := s.registerRoute(route); err != nil {
			s.http.Logger.Fatalf("unable to register route %s: %w", err)
		}
	}

	s.http.Logger.Fatal(s.http.Start(":8080"))
}
