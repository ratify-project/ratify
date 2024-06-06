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
	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"
	"github.com/ratify-project/ratify/pkg/referrerstore"
	sf "github.com/ratify-project/ratify/pkg/referrerstore/factory"
	su "github.com/ratify-project/ratify/pkg/referrerstore/utils"
	"github.com/ratify-project/ratify/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

const (
	discoverUse = "discover"
)

type discoverCmdOptions struct {
	configFilePath string
	subject        string
	artifactTypes  []string
	flatOutput     bool
}

func NewCmdDiscover(argv ...string) *cobra.Command {
	if len(argv) == 0 {
		argv = []string{os.Args[0]}
	}

	eg := fmt.Sprintf(`  # List referrers for a subject
  %s discover -c ./config.yaml -s myregistry/myrepo@sha256:34343`, strings.Join(argv, " "))

	var opts discoverCmdOptions

	cmd := &cobra.Command{
		Use:     discoverUse,
		Short:   "Discover referrers for a subject",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return discover(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringArrayVarP(&opts.artifactTypes, "artifactType", "t", nil, "artifact type to filter")
	flags.BoolVar(&opts.flatOutput, "flat", false, "Output referrers in a flat list format (default is tree format)")
	return cmd
}

type listResult struct {
	Name       string                         `json:"storeName"`
	References []ocispecs.ReferenceDescriptor `json:"References,omitempty"`
}

func Test(subject string) {
	_ = discover((discoverCmdOptions{
		subject:       subject,
		artifactTypes: []string{""},
	}))
}

func discover(opts discoverCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	subRef, err := utils.ParseSubjectReference(opts.subject)
	if err != nil {
		return err
	}

	if subRef.Digest == "" {
		fmt.Println(taggedReferenceWarning)
	}

	cf, err := config.Load(opts.configFilePath)
	if err != nil {
		return err
	}

	if err := logger.InitLogConfig(cf.LoggerConfig); err != nil {
		return err
	}

	rootImage := treeprint.NewWithRoot(subRef.String())

	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, config.GetDefaultPluginPath())
	if err != nil {
		return err
	}

	if subRef.Digest == "" {
		desc, err := su.ResolveSubjectDescriptor(context.Background(), &stores, subRef)
		if err != nil {
			return err
		}
		subRef.Digest = desc.Digest
	}

	results := []listResult{}
	for _, referrerStore := range stores {
		storeNode := rootImage.AddBranch(referrerStore.Name())
		result, err := listReferrersForStore(subRef, opts.artifactTypes, referrerStore, storeNode)
		if err != nil {
			return err
		}
		results = append(results, *result)
	}

	if !opts.flatOutput {
		fmt.Println(rootImage.String())
		return nil
	}

	return PrintJSON(results)
}

func listReferrersForStore(subRef common.Reference, artifactTypes []string, store referrerstore.ReferrerStore, treeNode treeprint.Tree) (*listResult, error) {
	var continuationToken string
	result := listResult{
		Name: store.Name(),
	}

	for {
		// subject descriptor has not been resolved thus nil passed in to ListReferrers
		lr, err := store.ListReferrers(context.Background(), subRef, artifactTypes, continuationToken, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get referrers list from subject %s: %w", subRef.Original, err)
		}

		continuationToken = lr.NextToken
		for _, ref := range lr.Referrers {
			refNode := treeNode.AddBranch(fmt.Sprintf("[%s]%s", ref.ArtifactType, ref.Digest.String()))
			sr := common.Reference{
				Path:     subRef.Path,
				Digest:   ref.Digest,
				Original: fmt.Sprintf("%s@%s", subRef.Path, ref.Digest),
			}

			subResult, err := listReferrersForStore(sr, artifactTypes, store, refNode)
			if err != nil {
				return nil, err
			}
			result.References = append(result.References, subResult.References...)
		}
		result.References = append(result.References, lr.Referrers...)
		if continuationToken == "" {
			break
		}
	}

	return &result, nil
}
