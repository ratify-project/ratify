/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/executor"
	"github.com/ratify-project/ratify/v2/internal/httpserver/config"
	"github.com/ratify-project/ratify/v2/internal/httpserver/tlssecret"
	"github.com/sirupsen/logrus"
)

const (
	serverRootURL        = "/ratify/gatekeeper/v2"
	verifyPath           = "verify"
	mutatePath           = "mutate"
	defaultVerifyTimeout = 5 * time.Second
	defaultMutateTimeout = 2 * time.Second
	readTimeout          = 5 * time.Second
	writeTimeout         = 5 * time.Second
	idleTimeout          = 60 * time.Second
)

type server struct {
	router   *mux.Router
	executor *ratify.Executor
	ServerOptions
}

// ServerOptions holds the configuration options for the Ratify server.
type ServerOptions struct {
	// HTTPServerAddress is the address where the server will listen for
	// incoming requests.
	// It should be in the format "host:port" (e.g., ":8080").
	// Required.
	HTTPServerAddress string

	// CertFile is the path to the TLS certificate file. If not provided, the
	// server will run without TLS.
	// Optional.
	CertFile string

	// KeyFile is the path to the TLS key file. If not provided, the server
	// will run without TLS.
	// Optional.
	KeyFile string

	// GatekeeperCACertFile is the path to the Gatekeeper CA certificate file.
	// Optional.
	GatekeeperCACertFile string

	// VerifyTimeout is the duration to wait for a verification request to
	// complete before timing out. Default is 5 seconds if not specified.
	// Optional.
	VerifyTimeout time.Duration

	// MutateTimeout is the duration to wait for a mutation request to
	// complete before timing out. Default is 2 seconds if not specified.
	// Optional.
	MutateTimeout time.Duration

	// DisableMutation indicates whether to disable the mutation handler.
	// If set to true, the mutation handler will not be registered.
	// Optional.
	DisableMutation bool

	// CertRotatorReady is a channel that signals when the certificate rotator
	// is ready. If not provided, the server will run without rotating the TLS
	// certificates.
	// Optional.
	CertRotatorReady chan struct{}
}

// StartServer initializes and starts the Ratify server with provided options
// and configuration file path.
func StartServer(opts *ServerOptions, configPath string) error {
	executorOpts, err := config.Load(configPath)
	if err != nil {
		logrus.Errorf("Failed to load executor options: %v", err)
		return err
	}

	server, err := newServer(opts, executorOpts)
	if err != nil {
		logrus.Errorf("Failed to create server: %v", err)
		return err
	}

	logrus.Infof("Starting server at port: %s", opts.HTTPServerAddress)
	return server.Run(opts.CertRotatorReady)
}

func newServer(serverOpts *ServerOptions, executorOpts *executor.Options) (*server, error) {
	e, err := executor.NewExecutor(executorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}
	server := &server{
		router:        mux.NewRouter(),
		executor:      e,
		ServerOptions: *serverOpts,
	}
	if server.VerifyTimeout == 0 {
		server.VerifyTimeout = defaultVerifyTimeout
	}
	if server.MutateTimeout == 0 {
		server.MutateTimeout = defaultMutateTimeout
	}
	if err := server.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register handlers: %w", err)
	}
	return server, nil
}

func (s *server) registerHandlers() error {
	if err := s.registerVerifyHandler(); err != nil {
		return err
	}

	if !s.DisableMutation {
		if err := s.registerMutateHandler(); err != nil {
			return err
		}
	}
	return nil
}

// TODO: implement mutate handler.
func (s *server) registerMutateHandler() error {
	mutateURL, err := url.JoinPath(serverRootURL, mutatePath)
	if err != nil {
		return err
	}
	s.router.Methods(http.MethodPost).PathPrefix(mutateURL).Handler(middlewareWithTimeout(s.mutateHandler(), s.MutateTimeout))
	return nil
}

func (s *server) registerVerifyHandler() error {
	verifyURL, err := url.JoinPath(serverRootURL, verifyPath)
	if err != nil {
		return err
	}
	s.router.Methods(http.MethodPost).Path(verifyURL).Handler(middlewareWithTimeout(s.verifyHandler(), s.VerifyTimeout))
	return nil
}

func (s *server) verifyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = s.verify(r.Context(), w, r)
	}
}

func (s *server) mutateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = s.mutate(r.Context(), w, r)
	}
}

func middlewareWithTimeout(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Run starts the HTTP server and listens for incoming requests.
// It also handles graceful shutdown on receiving an interrupt signal.
func (s *server) Run(certRotatorReady chan struct{}) error {
	srv := &http.Server{
		Addr:         s.HTTPServerAddress,
		Handler:      s.router,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}
	go func() {
		if s.CertFile != "" && s.KeyFile != "" {
			logrus.Infof("starting server with TLS at %s", s.HTTPServerAddress)
			if certRotatorReady != nil {
				<-certRotatorReady
				logrus.Infof("cert rotator is ready")
			}

			certWatcher, err := tlssecret.NewTLSSecretWatcher(s.GatekeeperCACertFile, s.CertFile, s.KeyFile)
			if err != nil {
				logrus.Errorf("failed to create TLS secret watcher: %v", err)
				return
			}
			// Use GetConfigForClient to dynamically load certificates.
			srv.TLSConfig = &tls.Config{
				MinVersion:         tls.VersionTLS13,
				GetConfigForClient: certWatcher.GetConfigForClient,
			}
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logrus.Errorf("failed to start server: %v", err)
			}
		} else {
			logrus.Infof("starting server without TLS at %s", s.HTTPServerAddress)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Errorf("failed to start server: %v", err)
			}
		}
	}()

	// Handle graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), s.VerifyTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown server: %v", err)
		return err
	}
	return nil
}
