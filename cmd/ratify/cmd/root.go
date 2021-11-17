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

import "github.com/spf13/cobra"

const (
	use       = "ratify"
	shortDesc = "Ratify is a reference artifact tool for managing and verifying reference artifacts"
)

var Root = New(use, shortDesc)

func New(use, short string) *cobra.Command {
	root := &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
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

	// TODO debug logging
	return root
}
