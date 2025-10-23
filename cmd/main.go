package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/shortcuts/codes/internal/markdown"
	"github.com/shortcuts/codes/internal/template"
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
		path:     "movies",
		filename: "movies.md",
	},
	{
		path:     "links",
		filename: "links.md",
	},
}

//go:embed views/*.md views/*.html scripts/*.js css/*.css img/*.png img/*.jpg img/favicon.ico img/site.webmanifest
var views embed.FS

type server struct {
	isLocal      bool
	cacheClient  *ttlcache.Cache[string, []byte]
	httpClient   *http.Client
	lastModified string
	logger       *zap.Logger
	parser       markdown.MarkdownParser
	router       *echo.Echo
}

func (s *server) close() {
	if err := s.router.Close(); err != nil {
		s.logger.Error("error while closing router", zap.Error(err))
	}
}

func newServer() *server {
	cacheClient := ttlcache.New(ttlcache.WithTTL[string, []byte](24*time.Hour), ttlcache.WithDisableTouchOnHit[string, []byte](), ttlcache.WithCapacity[string, []byte](1024*1024*10))

	server := &server{
		isLocal:      os.Getenv("DEV") != "",
		lastModified: time.Now().Format(time.RFC1123),
		httpClient:   &http.Client{Timeout: 1 * time.Second},
		cacheClient:  cacheClient,
	}

	server.router = echo.New()

	server.router.Use(
		middleware.Recover(),
		middleware.RemoveTrailingSlash(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(20),
		)),
	)

	var err error

	if server.isLocal {
		server.logger, err = zap.NewDevelopment()
	} else {
		server.logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(fmt.Sprintf("unable to create logger: %s", err))
	}

	server.router.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status == http.StatusNotFound || v.Status == http.StatusTooManyRequests {
				return nil
			}

			if strings.Contains(v.URI, "assets") {
				return nil
			}

			server.logger.Info("request", zap.String("URI", v.URI), zap.Int("status", v.Status))

			return nil
		},
	}))

	server.router.StaticFS("/assets", views)

	server.router.Renderer = template.NewTemplate(&views)

	server.parser = markdown.NewParser(views)

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
			s.logger.Fatal(fmt.Sprintf("unable to register route %s: %s", route.path, err.Error()))
		}
	}

	baseRoute := ""

	if s.isLocal {
		baseRoute = "localhost"
	}

	err := s.router.Start(baseRoute + ":1313")
	if err != nil {
		s.logger.Fatal(err.Error())
	}
}
