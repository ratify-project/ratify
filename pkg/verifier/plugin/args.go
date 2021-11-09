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
	PluginArgs       [][2]string
}

var _ pluginCommon.PluginArgs = &VerifierPluginArgs{}

func (args *VerifierPluginArgs) AsEnv() []string {
	env := os.Environ()
	pluginArgsStr := pluginCommon.Stringify(args.PluginArgs)

	// Duplicated values which come first will be overridden, so we must put the
	// custom values in the end to avoid being overridden by the process environments.
	// TODO replace the args
	env = append(env,
		"RATIFY_VERIFIER_COMMAND="+args.Command,
		"RATIFY_VERIFIER_SUBJECT="+args.SubjectReference,
		"RATIFY_VERIFIER_ARGS="+pluginArgsStr,
		fmt.Sprintf("%s=%s", VersionEnvKey, args.Version),
	)
	return pluginCommon.DedupEnv(env)
}
