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
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/featureflag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	use       = "ratify"
	shortDesc = "Ratify is a reference artifact tool for managing and verifying reference artifacts"
)

var Root = New(use, shortDesc)

func New(use, short string) *cobra.Command {
	featureflag.InitFeatureFlagsFromEnv()
	var enableDebug bool
	root := &cobra.Command{
		Use:   use,
		Short: short,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if enableDebug {
				common.SetLoggingLevel("debug", logrus.StandardLogger())
			} else {
				common.SetLoggingLevelFromEnv(logrus.StandardLogger())
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	root.AddCommand(NewCmdReferrer(use, referrerUse))
	root.AddCommand(NewCmdVerify(use, verifyUse))
	root.AddCommand(NewCmdServe(use, serveUse))
	root.AddCommand(NewCmdDiscover(use, discoverUse))
	root.AddCommand(NewCmdVersion(use, versionUse))
	root.AddCommand(NewCmdResolve(use, resolveUse))

	root.PersistentFlags().BoolVarP(&enableDebug, "debug", "d", false, "Enable debug mode. If enabled, set logger level to debug")
	return root
}
