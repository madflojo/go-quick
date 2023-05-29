/*
Package app is the primary runtime service.
*/
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/redis"
	"github.com/madflojo/tasks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Common errors returned by this app.
var (
	ErrShutdown = fmt.Errorf("application shutdown gracefully")
)

// Server represents the main server structure.
type Server struct {
	// cfg is used across the app package to contain configuration.
	cfg *viper.Viper

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router

	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// kv is the global reference for the K/V Store.
	kv hord.Database

	// log is used across the app package for logging.
	log *logrus.Logger

	// runCancel is a global context cancelFunc used to trigger the shutdown of applications.
	runCancel context.CancelFunc

	// runCtx is a global context used to control shutdown of the application.
	runCtx context.Context

	// scheduler is an internal task scheduler for recurring tasks.
	scheduler *tasks.Scheduler
}

// New creates a new instance of the Server struct.
// It takes a `cfg` parameter of type `*viper.Viper` for configuration.
// It returns a pointer to the created Server instance.
func New(cfg *viper.Viper) *Server {
	srv := &Server{cfg: cfg}

	// Create App Context
	srv.runCtx, srv.runCancel = context.WithCancel(context.Background())

	// Initiate a new logger
	srv.log = logrus.New()
	if srv.cfg.GetBool("debug") {
		srv.log.Level = logrus.DebugLevel
		srv.log.Debug("Enabling Debug Logging")
	}
	if srv.cfg.GetBool("trace") {
		srv.log.Level = logrus.TraceLevel
		srv.log.Debug("Enabling Trace Logging")
	}
	if srv.cfg.GetBool("disable_logging") {
		srv.log.Level = logrus.FatalLevel
	}

	return srv

}

// Run starts the primary application. It handles starting background services,
// populating package globals & structures, and clean up tasks.
func (srv *Server) Run() error {
	var err error

	// Initiate a new logger
	srv.log = logrus.New()
	if srv.cfg.GetBool("debug") {
		srv.log.Level = logrus.DebugLevel
		srv.log.Debug("Enabling Debug Logging")
	}
	if srv.cfg.GetBool("trace") {
		srv.log.Level = logrus.TraceLevel
		srv.log.Debug("Enabling Trace Logging")
	}
	if srv.cfg.GetBool("disable_logging") {
		srv.log.Level = logrus.FatalLevel
	}

	// Setup Scheduler
	srv.scheduler = tasks.New()
	defer srv.scheduler.Stop()

	// Config Reload
	if srv.cfg.GetInt("config_watch_interval") > 0 {
		_, err := srv.scheduler.Add(&tasks.Task{
			Interval: time.Duration(srv.cfg.GetInt("config_watch_interval")) * time.Second,
			TaskFunc: func() error {
				// Reload config using Viper's Watch capabilities
				err := srv.cfg.WatchRemoteConfig()
				if err != nil {
					return err
				}

				// Support hot enable/disable of debug logging
				if srv.cfg.GetBool("debug") {
					srv.log.Level = logrus.DebugLevel
				}

				// Support hot enable/disable of trace logging
				if srv.cfg.GetBool("trace") {
					srv.log.Level = logrus.TraceLevel
				}

				// Support hot enable/disable of all logging
				if srv.cfg.GetBool("disable_logging") {
					srv.log.Level = logrus.FatalLevel
				}

				srv.log.Tracef("Config reloaded from Consul")
				return nil
			},
		})
		if err != nil {
			srv.log.Errorf("Error scheduling Config watcher - %s", err)
		}
	}

	// Setup the DB Connection
	srv.kv, err = redis.Dial(redis.Config{
		Server:   srv.cfg.GetString("kv_server"),
		Password: srv.cfg.GetString("kv_password"),
	})
	if err != nil {
		return fmt.Errorf("could not establish database connection - %s", err)
	}
	defer srv.kv.Close()

	// Initialize the DB
	err = srv.kv.Setup()
	if err != nil {
		return fmt.Errorf("could not setup database - %s", err)
	}

	// Setup the HTTP Server
	srv.httpRouter = httprouter.New()
	srv.httpServer = &http.Server{
		Addr:    srv.cfg.GetString("listen_addr"),
		Handler: srv.httpRouter,
	}

	// Setup TLS Configuration
	if srv.cfg.GetBool("enable_tls") {
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
		srv.log.Infof("Received shutdown signal %s", s)

		// Shutdown the HTTP Server
		err := srv.httpServer.Shutdown(context.Background())
		if err != nil {
			srv.log.Errorf("Received errors when shutting down HTTP sessions %s", err)
		}

		// Close DB Sessions
		srv.kv.Close()

		// Shutdown the app via runCtx
		srv.runCancel()
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
	srv.log.Infof("Starting Listener on %s", srv.cfg.GetString("listen_addr"))
	if srv.cfg.GetBool("enable_tls") {
		err := srv.httpServer.ListenAndServeTLS(srv.cfg.GetString("cert_file"), srv.cfg.GetString("key_file"))
		if err != nil {
			if err == http.ErrServerClosed {
				// Wait until all outstanding requests are done
				<-srv.runCtx.Done()
				return ErrShutdown
			}
			return err
		}
	}
	err = srv.httpServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			// Wait until all outstanding requests are done
			<-srv.runCtx.Done()
			return ErrShutdown
		}
		return err
	}

	return nil
}

// Stop is used to gracefully shutdown the server.
func (srv *Server) Stop() {
	err := srv.httpServer.Shutdown(context.Background())
	if err != nil {
		srv.log.Errorf("Unexpected error while shutting down HTTP server - %s", err)
	}
	defer srv.runCancel()
}
