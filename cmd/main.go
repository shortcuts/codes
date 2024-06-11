package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shortcuts/codes/pkg/markdown"
	"github.com/shortcuts/codes/pkg/template"
	"golang.org/x/time/rate"
)

var routes = []string{"", "resume", "links"}

type server struct {
	http   *echo.Echo
	parser markdown.MarkdownParser
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

func (s *server) registerRoute(route string) error {
	title := route

	if route == "" {
		title = "home"
	}

	navbar, err := s.parser.ToHTML("views/navbar.md")
	if err != nil {
		return err
	}

	content, err := s.parser.ToHTML(fmt.Sprintf("views/%s.md", title))
	if err != nil {
		return err
	}

	s.http.GET(fmt.Sprintf("/%s", route), func(c echo.Context) error {
		return c.Render(http.StatusOK, "layout", map[string]any{
			"Filename": title,
			"Navbar":   navbar,
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

	s.http.Logger.Fatal(s.http.Start("localhost:42069"))
}
