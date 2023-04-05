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

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/httpserver"
	"github.com/deislabs/ratify/pkg/manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	serveUse = "serve"
)

type serveCmdOptions struct {
	configFilePath    string
	httpServerAddress string
	certDirectory     string
	caCertFile        string
	enableCrdManager  bool
	cacheSize         int
	cacheTTL          time.Duration
	metricsEnabled    bool
	metricsType       string
	metricsPort       int
}

func NewCmdServe(argv ...string) *cobra.Command {
	var opts serveCmdOptions

	cmd := &cobra.Command{
		Use:     serveUse,
		Short:   "Run ratify as a server",
		Example: "ratify server",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&opts.httpServerAddress, "http", "", "HTTP Address")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringVar(&opts.certDirectory, "cert-dir", "", "Path to ratify certs")
	flags.StringVar(&opts.caCertFile, "ca-cert-file", "", "Path to CA cert file")
	flags.BoolVar(&opts.enableCrdManager, "enable-crd-manager", false, "Start crd manager if enabled (default: false)")
	flags.IntVar(&opts.cacheSize, "cache-size", httpserver.DefaultCacheMaxSize, fmt.Sprintf("Cache size for the verifier http server (default: %d)", httpserver.DefaultCacheMaxSize))
	flags.DurationVar(&opts.cacheTTL, "cache-ttl", httpserver.DefaultCacheTTL, fmt.Sprintf("Cache TTL for the verifier http server (default: %fs)", httpserver.DefaultCacheTTL.Seconds()))
	flags.BoolVar(&opts.metricsEnabled, "metrics-enabled", false, "Enable metrics exporter if enabled (default: false)")
	flags.StringVar(&opts.metricsType, "metrics-type", httpserver.DefaultMetricsType, fmt.Sprintf("Metrics exporter type to use (default: %s)", httpserver.DefaultMetricsType))
	flags.IntVar(&opts.metricsPort, "metrics-port", httpserver.DefaultMetricsPort, fmt.Sprintf("Metrics exporter port to use (default: %d)", httpserver.DefaultMetricsPort))
	return cmd
}

func serve(opts serveCmdOptions) error {
	// in crd mode, the manager gets latest store/verifier from crd and pass on to the http server
	if opts.enableCrdManager {
		logrus.Infof("starting crd manager")
		go manager.StartManager()
		manager.StartServer(opts.httpServerAddress, opts.configFilePath, opts.certDirectory, opts.caCertFile, opts.cacheSize, opts.cacheTTL, opts.metricsEnabled, opts.metricsType, opts.metricsPort)

		return nil
	}

	getExecutor, err := config.GetExecutorAndWatchForUpdate(opts.configFilePath)
	if err != nil {
		return err
	}

	if opts.httpServerAddress != "" {
		server, err := httpserver.NewServer(context.Background(), opts.httpServerAddress, getExecutor, opts.certDirectory, opts.caCertFile, opts.cacheSize, opts.cacheTTL, opts.metricsEnabled, opts.metricsType, opts.metricsPort)
		if err != nil {
			return err
		}
		logrus.Infof("starting server at" + opts.httpServerAddress)
		if err := server.Run(); err != nil {
			return err
		}
	}

	return nil
}
