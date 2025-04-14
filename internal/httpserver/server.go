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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/executor"
	"github.com/ratify-project/ratify/v2/internal/httpserver/config"
	"github.com/sirupsen/logrus"
)

const (
	serverRootURL        = "/ratify/gatekeeper/v2"
	verifyPath           = "verify"
	defaultVerifyTimeout = 5 * time.Second
	readTimeout          = 5 * time.Second
	writeTimeout         = 5 * time.Second
	idleTimeout          = 60 * time.Second
)

var tlsCert atomic.Value

type server struct {
	address              string
	certFile             string
	keyFile              string
	router               *mux.Router
	executor             *ratify.Executor
	verifyRequestTimeout time.Duration
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

	// VerifyTimeout is the duration to wait for a verification request to
	// complete before timing out. Default is 5 seconds if not specified.
	// Optional.
	VerifyTimeout time.Duration
}

// StartServer initializes and starts the Ratify server with provided options
// and configuration file path.
func StartServer(opts *ServerOptions, configPath string) {
	executorOpts, err := config.Load(configPath)
	if err != nil {
		logrus.Errorf("Failed to load executor options: %v", err)
		os.Exit(1)
	}

	server, err := newServer(opts, executorOpts)
	if err != nil {
		logrus.Errorf("Failed to create server: %v", err)
		os.Exit(1)
	}

	logrus.Infof("Starting server at port: %s", opts.HTTPServerAddress)
	if err := server.Run(); err != nil {
		logrus.Errorf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

func newServer(serverOpts *ServerOptions, executorOpts *executor.Options) (*server, error) {
	e, err := executor.NewExecutor(executorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}
	server := &server{
		router:               mux.NewRouter(),
		address:              serverOpts.HTTPServerAddress,
		certFile:             serverOpts.CertFile,
		keyFile:              serverOpts.KeyFile,
		verifyRequestTimeout: serverOpts.VerifyTimeout,
		executor:             e,
	}
	if server.verifyRequestTimeout == 0 {
		server.verifyRequestTimeout = defaultVerifyTimeout
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
	if err := s.registerMutateHandler(); err != nil {
		return err
	}
	return nil
}

// TODO: implement mutate handler.
func (s *server) registerMutateHandler() error {
	return nil
}

func (s *server) registerVerifyHandler() error {
	verifyURL, err := url.JoinPath(serverRootURL, verifyPath)
	if err != nil {
		return err
	}
	s.router.Methods(http.MethodPost).Path(verifyURL).Handler(middlewareWithTimeout(s.verifyHandler(), s.verifyRequestTimeout))
	return nil
}

func (s *server) verifyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = s.verify(r.Context(), w, r)
	}
}

// verify handles the verification request from Gatekeeper.
func (s *server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return fmt.Errorf("failed to unmarshal request body to provider request: %w", err)
	}

	results := make([]externaldata.Item, len(providerRequest.Request.Keys))
	for idx, key := range providerRequest.Request.Keys {
		item := externaldata.Item{
			Key: key,
		}
		opts := ratify.ValidateArtifactOptions{
			Subject: key,
		}
		result, err := s.executor.ValidateArtifact(ctx, opts)
		if err != nil {
			item.Error = err.Error()
		}
		item.Value = convertResult(result)
		results[idx] = item
	}

	response := externaldata.ProviderResponse{
		APIVersion: "externaldata.gatekeeper.sh/v1beta1",
		Kind:       "ProviderResponse",
		Response: externaldata.Response{
			Items: results,
		},
	}
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
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
func (s *server) Run() error {
	srv := &http.Server{
		Addr:         s.address,
		Handler:      s.router,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}
	go func() {
		if s.certFile != "" && s.keyFile != "" {
			logrus.Infof("starting server with TLS at %s", s.address)
			if err := loadCertificate(s.certFile, s.keyFile); err != nil {
				logrus.Errorf("failed to load certificate: %v", err)
				return
			}
			// Use GetCertificate to dynamically load certificates.
			srv.TLSConfig = &tls.Config{
				MinVersion:     tls.VersionTLS13,
				GetCertificate: getCertificate,
			}
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logrus.Errorf("failed to start server: %v", err)
			}
		} else {
			logrus.Infof("starting server without TLS at %s", s.address)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Errorf("failed to start server: %v", err)
			}
		}
	}()

	// Handle graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), s.verifyRequestTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown server: %v", err)
		return err
	}
	return nil
}

func loadCertificate(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}
	tlsCert.Store(&cert)
	return nil
}

func getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert := tlsCert.Load(); cert != nil {
		return cert.(*tls.Certificate), nil
	}
	return nil, nil
}
