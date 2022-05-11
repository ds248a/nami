package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ds248a/nami/config"
	"github.com/ds248a/nami/log"
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
		if err := srv.ListenAndServeTLS("./sec/server.crt", "./sec/server.key"); err != nil && err != http.ErrServerClosed {
			fmt.Printf("err:%v \n", err)
			log.Fatal(err)
		}
	}()

	callOnExit(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		dst := make(chan bool)
		go func() {
			if err := srv.Shutdown(ctx); err != nil {
				log.Err(err)
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
