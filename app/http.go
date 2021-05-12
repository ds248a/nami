package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ds248a/nami/config"
	lg "github.com/ds248a/nami/log"
	"github.com/gin-gonic/gin"
)

// --------------------------------
//    Gin Server
// --------------------------------

// HTTP Server
func NewServer(r *gin.Engine, cfg *config.Config) {
	srv = &http.Server{
		Addr:           cfg.ServerAdr,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	callOnExit(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		dst := make(chan bool)
		go func() {
			if err := srv.Shutdown(ctx); err != nil {
				lg.LogErr(err)
			}
			dst <- true
			Debug("-- HTTP Close")
		}()

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-dst:
				break loop
			}
		}
	})
}
