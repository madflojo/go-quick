/*
Package app is the primary runtime service.
*/
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go-quick/config"
	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/redis"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Common errors returned by this app.
var (
	ErrShutdown = fmt.Errorf("application shutdown gracefully")
)

// srv is the global reference for the HTTP Server.
var srv *server

// db is the global reference for the DB Server.
var db hord.Database

// runCtx is a global context used to control shutdown of the application.
var runCtx context.Context

// runCancel is a global context cancelFunc used to trigger the shutdown of applications.
var runCancel context.CancelFunc

// cfg is used across the app package to contain configuration.
var cfg config.Config

// log is used across the app package for logging.
var log *logrus.Logger

// Run starts the primary application. It handles starting background services,
// populating package globals & structures, and clean up tasks.
func Run(c config.Config) error {
	var err error

	// Create App Context
	runCtx, runCancel = context.WithCancel(context.Background())

	// Apply config provided by main
	cfg = c

	// Initiate the logger
	log = logrus.New()
	if cfg.Debug {
		log.Level = logrus.DebugLevel
		log.Debug("Enabling Debug Logging")
	}
	if cfg.DisableLogging {
		log.Level = logrus.FatalLevel
	}

	// Setup the DB Connection
	db, err = redis.Dial(redis.Config{
		Server:   cfg.DBServer,
		Password: cfg.DBPassword,
	})
	if err != nil {
		return fmt.Errorf("could not establish database connection - %s", err)
	}
	defer db.Close()

	// Initialize the DB
	err = db.Setup()
	if err != nil {
		return fmt.Errorf("could not setup database - %s", err)
	}

	// Setup the HTTP Server
	srv = &server{
		httpRouter: httprouter.New(),
	}
	srv.httpServer = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: srv.httpRouter,
	}

	// Setup TLS Configuration
	if cfg.EnableTLS {
		srv.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	// Kick off Graceful Shutdown Go Routine
	go func() {
		// Make the Trap
		trap := make(chan os.Signal, 1)
		signal.Notify(trap, syscall.SIGTERM)

		// Wait for a signal then action
		s := <-trap
		log.Infof("Received shutdown signal %s", s)

		// Shutdown the HTTP Server
		err := srv.httpServer.Shutdown(context.Background())
		if err != nil {
			log.Errorf("Received errors when shutting down HTTP sessions %s", err)
		}

		// Close DB Sessions
		db.Close()

		// Shutdown the app via runCtx
		runCancel()
	}()

	// Register Health Check Handler used for Liveness checks
	srv.httpRouter.GET("/health", srv.middleware(srv.Health))

	// Register Health Check Handler used for Readiness checks
	srv.httpRouter.GET("/ready", srv.middleware(srv.Ready))

	// Register Hello World Handler
	srv.httpRouter.GET("/hello", srv.middleware(srv.Hello))
	srv.httpRouter.POST("/hello", srv.middleware(srv.SetHello))
	srv.httpRouter.PUT("/hello", srv.middleware(srv.SetHello))

	// Start HTTP Listener
	log.Infof("Starting Listener on %s", cfg.ListenAddr)
	if cfg.EnableTLS {
		err := srv.httpServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			if err == http.ErrServerClosed {
				// Wait until all outstanding requests are done
				<-runCtx.Done()
				return ErrShutdown
			}
			return err
		}
	}
	err = srv.httpServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			// Wait until all outstanding requests are done
			<-runCtx.Done()
			return ErrShutdown
		}
		return err
	}

	return nil
}

// Stop is used to gracefully shutdown the server.
func Stop() {
	defer srv.httpServer.Shutdown(context.Background())
	defer runCancel()
}
