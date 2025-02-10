package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jakthom/s3c/pkg/config"
	"github.com/jakthom/s3c/pkg/handler"
	"github.com/jakthom/s3c/pkg/middleware"
	"github.com/jakthom/s3c/pkg/util"
	"github.com/rs/zerolog/log"
)

var VERSION string

type S3c struct {
	config *config.Config
}

func (s *S3c) configure() {
	log.Debug().Msg("Configuring s3c")
	// Read configuration file
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read configuration file")
	}
	s.config = &conf
	util.Pprint(conf) // TODO -> fixme
}

func (s *S3c) Initialize() {
	log.Info().Msg("Initializing s3c")
	s.configure()
}

func (s *S3c) Run() {
	log.Info().Msg("Running s3c")
	// Set up http server and register s2 handlers

	router := mux.NewRouter()
	router.Handle("/s3c/health", middleware.RequestIdMiddleware(middleware.RequestLoggerMiddleware(http.HandlerFunc(handler.HealthcheckHandler))))
	router.Handle("/{path}", middleware.RequestIdMiddleware(middleware.RequestLoggerMiddleware(http.HandlerFunc(handler.HealthcheckHandler))))

	srv := &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: router,
	}
	go func() {
		log.Info().Msg("s3c is running with version: " + VERSION)
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Info().Msgf("s3c server shut down")
		}
	}()
	// Safe shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down s3c server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Stack().Err(err).Msg("server forced to shutdown")
	}
}
