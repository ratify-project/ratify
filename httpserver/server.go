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
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"

	"github.com/deislabs/ratify/config"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	ServerRootURL = "/ratify/gatekeeper/v1"
	certName      = "tls.crt"
	keyName       = "tls.key"
)

type (
	Server struct {
		Address       string
		Router        *mux.Router
		GetExecutor   config.GetExecutor
		Context       context.Context
		CertDirectory string
		CaCertFile    string
	}
)

func NewServer(context context.Context, address string, getExecutor config.GetExecutor, certDir string, caCertFile string) (*Server, error) {
	if address == "" {
		return nil, ServerAddrNotFoundError{}
	}

	server := &Server{
		Address:       address,
		GetExecutor:   getExecutor,
		Router:        mux.NewRouter(),
		Context:       context,
		CertDirectory: certDir,
		CaCertFile:    caCertFile,
	}
	server.registerHandlers()

	return server, nil
}

func (server *Server) Run() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server.Address)
	if err != nil {
		return err
	}

	lsnr, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	svr := &http.Server{
		Addr:    server.Address,
		Handler: server.Router,
	}

	if server.CertDirectory != "" {
		certFile := filepath.Join(server.CertDirectory, certName)
		keyFile := filepath.Join(server.CertDirectory, keyName)

		logrus.Info(fmt.Sprintf("%s: [%s:%s] [%s:%s]", "starting server using TLS", "certFile", certFile, "keyFile", keyFile))

		if server.CaCertFile != "" {
			caCert, err := ioutil.ReadFile(server.CaCertFile)
			if err != nil {
				panic(err)
			}

			clientCAs := x509.NewCertPool()
			clientCAs.AppendCertsFromPEM(caCert)

			config := &tls.Config{
				MinVersion: tls.VersionTLS13,
				ClientCAs:  clientCAs,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
			svr.TLSConfig = config
			logrus.Info(fmt.Sprintf("%s: [%s:%s] ", "loaded client CA certificate for mTLS", "CaFIle", server.CaCertFile))
		}
		return svr.ServeTLS(lsnr, certFile, keyFile)
	} else {
		logrus.Info("starting server without TLS")
		return svr.Serve(lsnr)
	}
}

func (server *Server) register(method, path string, handler ContextHandler) {
	server.Router.Methods(method).Path(path).Handler(&contextHandler{
		context: server.Context,
		handler: handler,
	})
}

func (server *Server) registerHandlers() {
	server.register("POST", ServerRootURL+"/verify", processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout()))
}

type ServerAddrNotFoundError struct{}

func (err ServerAddrNotFoundError) Error() string {
	return "The http server address configuration is not set. Skipping server creation"
}
