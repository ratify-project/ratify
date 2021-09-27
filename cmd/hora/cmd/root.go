package cmd

import "github.com/spf13/cobra"

const (
	use       = "hora"
	shortDesc = "Hora is a reference artifact tool for managing and verifying reference artifacts"
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

	// TODO debug logging
	return root
}
