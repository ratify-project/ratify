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

	"github.com/ratify-project/ratify/config"
	"github.com/ratify-project/ratify/internal/constants"
	"github.com/ratify-project/ratify/internal/logger"
	e "github.com/ratify-project/ratify/pkg/executor"
	ef "github.com/ratify-project/ratify/pkg/executor/core"
	pf "github.com/ratify-project/ratify/pkg/policyprovider/factory"
	sf "github.com/ratify-project/ratify/pkg/referrerstore/factory"
	"github.com/ratify-project/ratify/pkg/utils"
	vf "github.com/ratify-project/ratify/pkg/verifier/factory"
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

func NewCmdVerify(_ ...string) *cobra.Command {
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
	_ = verify((verifyCmdOptions{
		subject: subject,
	}))
}

func verify(opts verifyCmdOptions) error {
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

	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, config.GetDefaultPluginPath())

	if err != nil {
		return err
	}

	verifiers, err := vf.CreateVerifiersFromConfig(cf.VerifiersConfig, config.GetDefaultPluginPath(), constants.EmptyNamespace)

	if err != nil {
		return err
	}

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(cf.PoliciesConfig)

	if err != nil {
		return err
	}

	executor := ef.Executor{
		Verifiers:      verifiers,
		ReferrerStores: stores,
		PolicyEnforcer: policyEnforcer,
		Config:         &cf.ExecutorConfig,
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
