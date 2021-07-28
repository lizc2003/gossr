package util

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lizc2003/gossr/common/tlog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewGinEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())

	_ = e.SetTrustedProxies([]string{"127.0.0.1/8", "192.168.0.1/16", "172.16.0.1/12", "10.0.0.1/8"})
	return e
}

func GraceHttpServe(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var serveErr error
	serveEnd := make(chan bool, 1)

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			tlog.Error("server start error:", err)
			serveErr = err
			serveEnd <- true
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-serveEnd:
		return serveErr
	case <-quit:
		tlog.Info("Shutting down server...")
		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			tlog.Error("Server shutdown error:", err)
		}
		return nil
	}
}
