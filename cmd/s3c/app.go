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
	s3auth "github.com/jakthom/s3c/pkg/s3/auth"
	s3bucket "github.com/jakthom/s3c/pkg/s3/bucket"
	s3handler "github.com/jakthom/s3c/pkg/s3/handler"
	s3middleware "github.com/jakthom/s3c/pkg/s3/middleware"
	s3object "github.com/jakthom/s3c/pkg/s3/object"
	s3service "github.com/jakthom/s3c/pkg/s3/service"
	"github.com/jakthom/s3c/pkg/util"
	"github.com/rs/zerolog/log"
)

var VERSION string

type S3c struct {
	config         *config.Config
	server         *http.Server
	origin         *fileorigin.FileOrigin
	authController *s3auth.BasicAuthController // TODO -> Add more customizable authorization
	serviceHandler *s3service.ServiceHandler
	bucketHandler  *s3bucket.BucketHandler
	objectHandler  *s3object.ObjectHandler
}

func (s *S3c) configure() {
	log.Debug().Msg("Configuring s3c")
	// Read configuration file
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read configuration file")
	}
	s.config = &conf
	util.Pprint(s.config)
}

func (s *S3c) initializeServer() {
	router := mux.NewRouter()
	// Generic Middleware
	router.Use(middleware.RequestIdMiddleware)
	// router.Use(middleware.DebugMiddleware)
	// S3 Middleware
	router.Use(s3middleware.EtagMiddleware)
	router.Use(s3middleware.AuthenticationMiddleware(s.authController))
	// s3c metadata routes
	router.Handle("/s3c/health", http.HandlerFunc(handler.HealthcheckHandler))
	// S3 Service
	router.Handle("/", http.HandlerFunc(s.serviceHandler.Get)) // Service
	// S3 Object
	s3object.AddSubrouter(router, s.objectHandler)
	// S3 Bucket
	s3bucket.AddSubrouter(router, s.bucketHandler)
	// Not Implemented routes
	s3handler.AddNotImplementedRoutes(router)
	// Method Not Allowed
	router.MethodNotAllowedHandler = s3handler.MethodNotAllowedHandler()
	// Not Found Handler
	router.NotFoundHandler = s3handler.NotFoundHandler()
	s.server = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: router,
	}
}

func (s *S3c) Initialize() {
	log.Info().Msg("Initializing s3c")
	s.configure()
	s.origin = fileorigin.NewOrigin("data")
	s.authController = &s3auth.BasicAuthController{
		Region:          "us-east-1", // TODO -> Fixme
		AccessKeyId:     s.config.Auth.KeyID,
		SecretAccessKey: s.config.Auth.Secret,
	}
	s.serviceHandler = &s3service.ServiceHandler{
		Controller: s.origin.ServiceController,
	}
	s.bucketHandler = &s3bucket.BucketHandler{
		Controller: s.origin.BucketController,
	}
	s.objectHandler = &s3object.ObjectHandler{
		Controller: s.origin.ObjectController,
	}
	s.initializeServer()
}

func (s *S3c) Run() {
	log.Info().Msg("Running s3c")
	go func() {
		log.Info().Msg("s3c is running with version: " + VERSION)
		if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
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
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatal().Stack().Err(err).Msg("server forced to shutdown")
	}
}
