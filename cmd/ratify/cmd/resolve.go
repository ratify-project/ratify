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

	"github.com/ratify-project/ratify/config"
	"github.com/ratify-project/ratify/internal/logger"
	sf "github.com/ratify-project/ratify/pkg/referrerstore/factory"
	su "github.com/ratify-project/ratify/pkg/referrerstore/utils"
	"github.com/ratify-project/ratify/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	resolveUse = "resolve"
)

type resolveCmdOptions struct {
	configFilePath string
	subject        string
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
	return cmd
}

func resolve(opts resolveCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	subRef, err := utils.ParseSubjectReference(opts.subject)
	if err != nil {
		return err
	}

	cf, err := config.Load(opts.configFilePath)
	if err != nil {
		return err
	}

	if err := logger.InitLogConfig(cf.LoggerConfig); err != nil {
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
