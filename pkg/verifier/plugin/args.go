package plugin

import (
	"fmt"
	"os"

	pluginCommon "github.com/deislabs/ratify/pkg/common/plugin"
)

type VerifierPluginArgs struct {
	Command          string
	Version          string
	SubjectReference string
}

var _ pluginCommon.PluginArgs = &VerifierPluginArgs{}

func (args *VerifierPluginArgs) AsEnviron() []string {
	env := os.Environ()

	env = append(env,
		"RATIFY_VERIFIER_COMMAND="+args.Command,
		"RATIFY_VERIFIER_SUBJECT="+args.SubjectReference,
		fmt.Sprintf("%s=%s", VersionEnvKey, args.Version),
	)
	return pluginCommon.MergeDuplicateEnviron(env)
}
