package cmd

import (
	"context"
	"errors"

	"github.com/deislabs/hora/config"
	e "github.com/deislabs/hora/pkg/executor"
	ef "github.com/deislabs/hora/pkg/executor/core"
	pf "github.com/deislabs/hora/pkg/policyprovider/configpolicy"
	sf "github.com/deislabs/hora/pkg/referrerstore/factory"
	vf "github.com/deislabs/hora/pkg/verifier/factory"
	"github.com/spf13/cobra"
)

const (
	verifyUse = "verify"
)

type verifyCmdOptions struct {
	configFilePath string
	subject        string
	artifactTypes  []string
	silentMode     bool
}

func NewCmdVerify(argv ...string) *cobra.Command {

	var opts verifyCmdOptions

	cmd := &cobra.Command{
		Use:     verifyUse,
		Short:   "Verify a subject",
		Example: "sample example",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return verify(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.subject, "subject", "s", "", "Subject Reference")
	flags.StringVarP(&opts.configFilePath, "config", "c", "", "Config File Path")
	flags.StringArrayVarP(&opts.artifactTypes, "artifactType", "t", nil, "artifact type to filter")
	flags.BoolVar(&opts.silentMode, "silent", false, "Silent output")
	return cmd
}

func TestVerify(subject string) {
	verify((verifyCmdOptions{
		subject: subject,
	}))
}

func verify(opts verifyCmdOptions) error {
	if opts.subject == "" {
		return errors.New("subject parameter is required")
	}

	cf, err := config.Load(opts.configFilePath)
	if err != nil {
		return err
	}

	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, config.GetDefaultPluginPath())

	if err != nil {
		return err
	}

	verifiers, err := vf.CreateVerifiersFromConfig(cf.VerifiersConfig, config.GetDefaultPluginPath())

	if err != nil {
		return err
	}

	executor := ef.Executor{
		Verifiers:      verifiers,
		ReferrerStores: stores,
		PolicyEnforcer: pf.PolicyEnforcer{},
	}

	verifyParameters := e.VerifyParameters{
		Subject:        opts.subject,
		ReferenceTypes: opts.artifactTypes,
	}

	result, err := executor.VerifySubject(context.Background(), verifyParameters)

	if err != nil {
		return err
	}

	if !opts.silentMode {
		return PrintJSON(result)
	}

	return nil
}
