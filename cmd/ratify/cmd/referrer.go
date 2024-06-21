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
	"github.com/ratify-project/ratify/pkg/ocispecs"
	sf "github.com/ratify-project/ratify/pkg/referrerstore/factory"
	"github.com/ratify-project/ratify/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	referrerUse = "referrer"
)

type referrerCmdOptions struct {
	configFilePath string
	subject        string
	digest         string
	storeName      string
}

func NewCmdReferrer(argv ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   referrerUse,
		Short: "Discover referrers for a subject",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(NewCmdShowBlob(argv...))
	cmd.AddCommand(NewCmdShowRefManifest(argv...))
	return cmd
}

func NewCmdShowBlob(argv ...string) *cobra.Command {
	var opts referrerCmdOptions

	if len(argv) == 0 {
		argv = []string{os.Args[0]}
	}

	eg := fmt.Sprintf(`  # Show blob contents in a store
  %s show-blob -c ./config.yaml -s myregistry/myrepo@sha256:34343 --store ociregistry -d sha256:3435`, strings.Join(argv, " "))

	cmd := &cobra.Command{
		Use:     "show-blob [OPTIONS]",
		Short:   "show blob at a digest",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showBlob(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringVarP(&opts.digest, "digest", "d", "", "blob digest")
	flags.StringVar(&opts.storeName, "store", "", "store name")
	return cmd
}

func NewCmdShowRefManifest(argv ...string) *cobra.Command {
	var opts referrerCmdOptions

	if len(argv) == 0 {
		argv = []string{os.Args[0]}
	}

	eg := fmt.Sprintf(`  # Show reference manifest for the subject with a given digest
  %s show-manifest -c ./config.yaml -s myregistry/myrepo@sha256:34343 --store ociregistry -d sha256:3456`, strings.Join(argv, " "))

	cmd := &cobra.Command{
		Use:     "show-manifest [OPTIONS]",
		Short:   "show rference manifest at a digest",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showRefManifest(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringVarP(&opts.digest, "digest", "d", "", "blob digest")
	flags.StringVar(&opts.storeName, "store", "", "store name")
	return cmd
}

func showBlob(opts referrerCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	if opts.digest == "" {
		return errors.New("digest parameter is required")
	}

	if opts.storeName == "" {
		return errors.New("store name parameter is required")
	}

	subRef, err := utils.ParseSubjectReference(opts.subject)
	if err != nil {
		return err
	}

	digest, err := utils.ParseDigest(opts.digest)
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

	for _, referrerStore := range stores {
		if referrerStore.Name() == opts.storeName {
			content, err := referrerStore.GetBlobContent(context.Background(), subRef, digest)
			if err != nil {
				return err
			}
			os.Stdout.Write(content)
			return nil
		}
	}

	return fmt.Errorf("cannot find store with name %s", opts.storeName)
}

func showRefManifest(opts referrerCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	if opts.storeName == "" {
		return errors.New("store name parameter is required")
	}

	if opts.digest == "" {
		return errors.New("digest parameter is required")
	}

	subRef, err := utils.ParseSubjectReference(opts.subject)
	if err != nil {
		return err
	}

	if subRef.Digest == "" {
		fmt.Println(taggedReferenceWarning)
	}

	digest, err := utils.ParseDigest(opts.digest)
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

	ref := fmt.Sprintf("%s@%s", subRef.Path, digest)

	manifestRef, err := utils.ParseSubjectReference(ref)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, referrerStore := range stores {
		if referrerStore.Name() == opts.storeName {
			manifestDesc, err := referrerStore.GetSubjectDescriptor(ctx, manifestRef)
			if err != nil {
				return fmt.Errorf("failed to resolve subject descriptor from store: %w", err)
			}

			manifestReferenceDesc := ocispecs.ReferenceDescriptor{
				Descriptor: manifestDesc.Descriptor,
			}

			manifest, err := referrerStore.GetReferenceManifest(ctx, subRef, manifestReferenceDesc)
			if err != nil {
				return fmt.Errorf("failed to fetch manifest for reference %s: %w", subRef.Original, err)
			}
			return PrintJSON(manifest)
		}
	}

	return fmt.Errorf("cannot find store with name %s", opts.storeName)
}
