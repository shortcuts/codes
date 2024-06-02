package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/shortcuts/codes/cmd/pkg"
)

type router struct {
	*gin.Engine
}

func (r *router) add(route, path string) {
	r.GET(route, func(ctx *gin.Context) {
		content, err := pkg.ReadMarkdownFile(path)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)

			return
		}

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", pkg.MarkdownToHTML(content))
	})
}

func newRouter() router {
	r := router{gin.Default()}

	r.Use(
		cors.New(cors.Config{
			AllowMethods:     []string{http.MethodGet},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
			AllowAllOrigins:  true,
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		}),
	)

	r.add("/", "cmd/home.md")
	r.add("/resume", "cmd/resume.md")

	return r
}

func main() {
	router := newRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Println("receive interrupt signal")
		if err := server.Close(); err != nil {
			log.Fatal("Server Close:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Println("Server closed under request")

			return
		}

		log.Fatal("Server closed unexpect")
	}

	log.Println("Server exiting")
}
