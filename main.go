package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/pawancldcvr/audit-db/initializer"
	"github.com/rs/zerolog/log"
)

func main() {
	router, stopCh := initializer.Init()
	pprof.Register(router)

	//serverConf, _ := config.Server()
	var srv = http.Server{
		Addr:    "localhost:9080", //serverConf.Port,
		Handler: router,
	}

	go func() {
		log.Info().Msg("Starting server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msg(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Single go routine to shutdown connection
	stopCh <- true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msg("Server forced to shutdown: ")
	}

	log.Info().Msg("Server exiting")
}
