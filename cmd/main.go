package main

import (
	"embed"
	"fmt"
	gotemplate "html/template"
	"net/http"
	"time"

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

//go:embed views/*.md views/*.html css/*.css img/*.png img/favicon.ico img/site.webmanifest
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

func newServer() server {
	e := echo.New()

	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Skipper: func(c echo.Context) bool {
				return c.Response().Status == http.StatusNotFound || c.Response().Status == http.StatusTooManyRequests
			},
		}),
		middleware.Recover(),
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Skipper: middleware.DefaultSkipper,
			Timeout: 1 * time.Second,
		}),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(20),
		)),
	)

	e.StaticFS("assets", views)

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

	s.http.Logger.Fatal(s.http.Start(":1313"))
}
