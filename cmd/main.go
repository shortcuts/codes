package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/shortcuts/codes/pkg/markdown"
	"github.com/shortcuts/codes/pkg/template"
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

//go:embed views/*.md views/*.html scripts/*.js css/*.css img/*.png img/*.jpg img/favicon.ico img/site.webmanifest
var views embed.FS

type server struct {
	router       *echo.Echo
	parser       markdown.MarkdownParser
	lastModified string `envconfig:"LAST_MODIFIED" required:"false"`
}

func (s *server) close() error {
	if err := s.router.Close(); err != nil {
		return err
	}

	return nil
}

func newServer() *server {
	err := godotenv.Load("cmd/.env")
	if err != nil {
		panic(err)
	}

	server := &server{}

	err = envconfig.Process("", server)
	if err != nil {
		panic(err)
	}

	server.router = echo.New()

	server.router.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.RemoveTrailingSlash(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(20),
		)),
	)

	server.router.StaticFS("/assets", views)

	server.router.Renderer = template.NewTemplate(&views)

	server.parser = markdown.NewParser(&views)

	content, err := server.parser.ToHTML("views/404.md")
	if err != nil {
		panic(err)
	}

	server.router.RouteNotFound("*", func(c echo.Context) error {
		c.Response().Header().Add("Last-Modified", server.lastModified)
		return c.Render(http.StatusOK, "layout", map[string]any{"Content": content})
	})

	return server
}

func (s *server) registerRoute(route route) error {
	content, err := s.parser.ToHTML(fmt.Sprintf("views/%s", route.filename))
	if err != nil {
		return err
	}

	s.router.GET(fmt.Sprintf("/%s", route.path), func(c echo.Context) error {
		c.Response().Header().Add("Last-Modified", s.lastModified)
		return c.Render(http.StatusOK, "layout", map[string]any{"Content": content})
	})

	return nil
}

func main() {
	s := newServer()
	defer s.close()

	for _, route := range routes {
		if err := s.registerRoute(route); err != nil {
			s.router.Logger.Fatalf("unable to register route %s: %w", err)
		}
	}

	s.router.Logger.Fatal(s.router.Start("localhost:1313"))
}
