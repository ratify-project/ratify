package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/deislabs/ratify/internal/version"
	"github.com/spf13/cobra"
)

const (
	versionUse = "version"
)

func NewCmdVersion(argv ...string) *cobra.Command {
	eg := `  Example - print version:
ratify version`

	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   "Show the ratify version information",
		Example: eg,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion()
		},
	}

	return cmd
}

func runVersion() error {
	items := [][]string{
		{"Go version", runtime.Version()},
	}

	// Tag is inserted as the version
	if len(version.GitTag) > 0 {
		items = append([][]string{{"Version", version.GitTag}}, items...)
	}

	if len(version.GitCommitHash) > 0 {
		items = append(items, []string{"Git commit", version.GitCommitHash})
	}

	if len(version.GitTreeState) > 0 &&
		version.GitTreeState != "unmodified" {
		items = append(items, []string{"Git tree", version.GitTreeState})
	}

	// Get max string lenght of first column
	var size = 0
	for _, item := range items {
		if size < len(item[0]) {
			size = len(item[0])
		}
	}

	for _, item := range items {
		fmt.Println(item[0] + ": " + strings.Repeat(" ", size-len(item[0])) + item[1])
	}

	return nil
}
