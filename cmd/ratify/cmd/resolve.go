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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deislabs/ratify/cache"
	"github.com/deislabs/ratify/config"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	su "github.com/deislabs/ratify/pkg/referrerstore/utils"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	resolveUse = "resolve"
)

type resolveCmdOptions struct {
	configFilePath string
	subject        string
	cacheType      string
	cacheSize      int
	cacheKeyNumber int
}

func NewCmdResolve(argv ...string) *cobra.Command {
	if len(argv) == 0 {
		argv = []string{os.Args[0]}
	}

	eg := fmt.Sprintf(`  # Resolve digest of a subject that is referenced by a tag
  %s resolve -c ./config.yaml -s myregistry/myrepo:v1`, strings.Join(argv, " "))

	var opts resolveCmdOptions

	cmd := &cobra.Command{
		Use:     resolveUse,
		Short:   "Resolve digest of a subject that is referenced by a tag",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return resolve(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringVar(&opts.cacheType, "cache-type", cache.DefaultCacheType, fmt.Sprintf("Cache type to use (default: %s)", cache.DefaultCacheType))
	flags.IntVar(&opts.cacheSize, "cache-size", cache.DefaultCacheMaxSize, fmt.Sprintf("Cache size (default: %d)", cache.DefaultCacheMaxSize))
	flags.IntVar(&opts.cacheKeyNumber, "cache-key-number", cache.DefaultCacheKeyNumber, fmt.Sprintf("Cache Key Size (default: %d)", cache.DefaultCacheKeyNumber))
	return cmd
}

func resolve(opts resolveCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	// initialize global cache of specified type
	_, err := cache.NewCacheProvider(opts.cacheType, opts.cacheSize, opts.cacheKeyNumber)
	if err != nil {
		return fmt.Errorf("error initializing cache of type %s: %w", opts.cacheType, err)
	}
	logrus.Debugf("initialized cache of type %s", opts.cacheType)

	subRef, err := utils.ParseSubjectReference(opts.subject)
	if err != nil {
		return err
	}

	cf, err := config.Load(opts.configFilePath)
	if err != nil {
		return err
	}

	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, config.GetDefaultPluginPath())

	if err != nil {
		return err
	}

	result, err := su.ResolveSubjectDescriptor(context.Background(), &stores, subRef)

	if err != nil {
		return err
	}

	fmt.Println(result.Digest)
	return nil
}
