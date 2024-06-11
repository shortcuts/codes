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

func (s *server) registerBlock(title string) {
	s.http.GET(fmt.Sprintf("/views/%s", title), func(c echo.Context) error {
		content, err := s.parser.ToHTML(fmt.Sprintf("views/%s.md", title))
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, string(content))
	})
}

func (s *server) registerRoute(title string) {
	route := title

	if route == "home" {
		route = ""
	}

	s.registerBlock(title)

	s.http.GET(fmt.Sprintf("/%s", route), func(c echo.Context) error {
		return c.Render(http.StatusOK, "layout", map[string]any{
			"Filename": title,
		})
	})
}

func main() {
	s := newServer()
	defer s.close()

	s.registerRoute("home")
	s.registerRoute("resume")
	s.registerRoute("links")

	s.registerBlock("navbar")

	s.http.Logger.Fatal(s.http.Start("localhost:42069"))
}
