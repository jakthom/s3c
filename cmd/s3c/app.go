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
	fileorigin "github.com/jakthom/s3c/pkg/origin/file"
	s3notimplemented "github.com/jakthom/s3c/pkg/s3/notimplemented"
	s3object "github.com/jakthom/s3c/pkg/s3/object"
	s3service "github.com/jakthom/s3c/pkg/s3/service"
	"github.com/jakthom/s3c/pkg/util"
	"github.com/rs/zerolog/log"
)

var VERSION string

type S3c struct {
	config         *config.Config
	serviceHandler *s3service.ServiceHandler
<<<<<<< Updated upstream
=======
	bucketHandler  *s3bucket.BucketHandler
	objectHandler  *s3object.ObjectHandler
>>>>>>> Stashed changes
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
	s.serviceHandler = &s3service.ServiceHandler{
		Controller: &fileorigin.FileOriginServiceController{},
	}
<<<<<<< Updated upstream

=======
	s.bucketHandler = &s3bucket.BucketHandler{
		Controller: s.origin.BucketController,
	}
	s.objectHandler = &s3object.ObjectHandler{
		Controller: s.origin.ObjectController,
	}
	s.initializeServer()
>>>>>>> Stashed changes
}

func (s *S3c) Run() {
	log.Info().Msg("Running s3c")
	// Set up http server and register s2 handlers

	router := mux.NewRouter()
	s3notimplemented.AddNotImplementedRoutes(router)
	router.Use(middleware.RequestIdMiddleware)
	// router.Use(middleware.RequestLoggerMiddleware)
	router.Use(middleware.DebugMiddleware)
	router.Handle("/s3c/health", http.HandlerFunc(handler.HealthcheckHandler))
	router.Handle("/", http.HandlerFunc(s.serviceHandler.Get)) // Service
<<<<<<< Updated upstream
	router.Handle("/{path}", http.HandlerFunc(handler.HealthcheckHandler))

=======
	// S3 Object
	s3object.AddSubrouter(router, s.objectHandler)
	// S3 Bucket
	s3bucket.AddSubrouter(router, s.bucketHandler)
>>>>>>> Stashed changes
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
