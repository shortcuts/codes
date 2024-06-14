package main

import (
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

type server struct {
	http   *echo.Echo
	parser markdown.MarkdownParser
	navbar *gotemplate.HTML
}

func (s *server) close() error {
	if err := s.http.Close(); err != nil {
		return err
	}

	return nil
}

func newServer() server {
	s := server{
		http:   echo.New(),
		parser: markdown.NewParser(),
	}

	navbar, err := s.parser.ToHTML("views/navbar.md")
	if err != nil {
		panic(err)
	}

	s.navbar = navbar

	s.http.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(20),
		)),
	)

	s.http.Static("assets", "css")
	s.http.Static("views", "views")

	s.http.Renderer = template.NewTemplate()

	return s
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

	s.http.Logger.Fatal(s.http.Start("localhost:8080"))
}
