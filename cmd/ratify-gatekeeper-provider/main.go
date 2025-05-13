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

package main

import (
	"errors"
	"flag"
	"time"

	"github.com/ratify-project/ratify/v2/internal/httpserver"
	"github.com/sirupsen/logrus"
)

// main is the entry point for the Ratify server.
func main() {
	if err := startRatify(parse()); err != nil {
		logrus.Errorf("Failed to start Ratify: %v", err)
		panic(err)
	}
}

type options struct {
	configFilePath    string
	httpServerAddress string
	certFile          string
	keyFile           string
	verifyTimeout     time.Duration
}

func parse() *options {
	opts := &options{}
	flag.StringVar(&opts.configFilePath, "config", "", "Path to the Ratify configuration file")
	flag.StringVar(&opts.httpServerAddress, "address", "", "HTTP server address")
	flag.StringVar(&opts.certFile, "cert-file", "", "Path to the TLS certificate file")
	flag.StringVar(&opts.keyFile, "key-file", "", "Path to the TLS key file")
	flag.DurationVar(&opts.verifyTimeout, "verify-timeout", 5*time.Second, "Verification timeout duration (e.g. 5s, 1m), default is 5 seconds")

	flag.Parse()
	logrus.Infof("Starting Ratify with options: %+v", opts)
	return opts
}

func startRatify(opts *options) error {
	if len(opts.httpServerAddress) == 0 {
		return errors.New("HTTP server address is required")
	}
	serverOpts := &httpserver.ServerOptions{
		HTTPServerAddress: opts.httpServerAddress,
		CertFile:          opts.certFile,
		KeyFile:           opts.keyFile,
		VerifyTimeout:     opts.verifyTimeout,
	}
	return httpserver.StartServer(serverOpts, opts.configFilePath)
}
