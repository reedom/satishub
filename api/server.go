package api

import (
	"context"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/reedom/satishub/pkg/satis"
)

// Server manages the web servers for staishub services.
type Server struct {
	service satis.Service
	log     *log.Logger
	debug   bool
}

// NewServer creates Server.
func NewServer(service satis.Service, logger *log.Logger, debug bool) Server {
	return Server{
		service: service,
		log:     logger,
		debug:   debug,
	}
}

// Serve starts serving new HTTP web server.
func (s Server) Serve(ctx context.Context, addr string) error {
	srv := http.Server{
		Addr:    addr,
		Handler: s.setupHandler(),
	}

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-ch:
		return err
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctxShutdown)
}

// ServeTLS starts serving new HTTPS web server.
func (s Server) ServeTLS(ctx context.Context, addr, certFile, keyFile string) error {
	srv := http.Server{
		Addr:    addr,
		Handler: s.setupHandler(),
	}

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- srv.ListenAndServeTLS(certFile, keyFile)
	}()

	select {
	case <-ctx.Done():
	case err := <-ch:
		return err
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctxShutdown)
}

func (s Server) setupHandler() *gin.Engine {
	if !s.debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.POST("/webhook/gitlab", s.handleGitlab)
	r.GET("/config", s.readConfig)
	r.StaticFile("/", path.Join(s.service.RepoPath(), "index.html"))
	r.Use(static.Serve("/", static.LocalFile(s.service.RepoPath(), false)))
	return r
}
