package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deislabs/hora/config"
	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	sf "github.com/deislabs/hora/pkg/referrerstore/factory"
	"github.com/deislabs/hora/pkg/utils"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

const (
	referrerUse = "referrer"
)

type referrerCmdOptions struct {
	configFilePath string
	subject        string
	artifactTypes  []string
	digest         string
	storeName      string
	flatOutput     bool
}

func NewCmdReferrer(argv ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   referrerUse,
		Short: "Discover referrers for a subject",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmd.AddCommand(NewCmdReferrerList(argv...))
	cmd.AddCommand(NewCmdShowBlob(argv...))
	cmd.AddCommand(NewCmdShowRefManifest(argv...))
	return cmd
}

func NewCmdReferrerList(argv ...string) *cobra.Command {
	var opts referrerCmdOptions

	if len(argv) == 0 {
		argv = []string{os.Args[0]}
	}

	eg := fmt.Sprintf(`  # List referrers for a subject
  %s list -c ./config.yaml -s myregistry/myrepo@sha256:34343`, strings.Join(argv, " "))

	cmd := &cobra.Command{
		Use:     "list [OPTIONS]",
		Short:   "List referrers to a subject",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listReferrers(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringArrayVarP(&opts.artifactTypes, "artifactType", "t", nil, "artifact type to filter")
	flags.BoolVar(&opts.flatOutput, "flat", false, "Output referrers in a flat list format (default is tree format)")
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

type listResult struct {
	Name       string                         `json:"storeName"`
	References []ocispecs.ReferenceDescriptor `json:"References,omitempty"`
}

func Test(subject string) {
	listReferrers((referrerCmdOptions{
		subject:       subject,
		artifactTypes: []string{"myartifact"},
	}))
}

func listReferrers(opts referrerCmdOptions) error {

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

	// TODO replace with code
	rootImage := treeprint.NewWithRoot(subRef.String())

	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, config.GetDefaultPluginPath())

	if err != nil {
		return err
	}

	type Result struct {
		Name       string
		References []ocispecs.ReferenceDescriptor
	}

	var results []listResult

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
		lr, err := store.ListReferrers(context.Background(), subRef, artifactTypes, continuationToken)
		if err != nil {
			return nil, err
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

	desc := ocispecs.ReferenceDescriptor{Descriptor: v1.Descriptor{Digest: digest}}
	for _, referrerStore := range stores {
		if referrerStore.Name() == opts.storeName {
			manifest, err := referrerStore.GetReferenceManifest(context.Background(), subRef, desc)
			if err != nil {
				return err
			}
			return PrintJSON(manifest)

		}
	}

	return fmt.Errorf("cannot find store with name %s", opts.storeName)
}
