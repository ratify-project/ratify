package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/ocispecs"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/utils"
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
	digest         string
	storeName      string
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

func discover(opts discoverCmdOptions) error {

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
