package plugin

import (
	"fmt"
	"os"

	pluginCommon "github.com/deislabs/ratify/pkg/common/plugin"
)

type ReferrerStorePluginArgs struct {
	Command          string
	Version          string
	SubjectReference string
	PluginArgs       [][2]string
}

var _ pluginCommon.PluginArgs = &ReferrerStorePluginArgs{}

func (args *ReferrerStorePluginArgs) AsEnviron() []string {
	env := os.Environ()
	pluginArgsStr := pluginCommon.Concat(args.PluginArgs)

	// Duplicated values which come first will be overridden, so we must put the
	// custom values in the end to avoid being overridden by the process environments.
	// TODO replace the args
	env = append(env,
		"RATIFY_STORE_COMMAND="+args.Command,
		"RATIFY_STORE_SUBJECT="+args.SubjectReference,
		"RATIFY_STORE_ARGS="+pluginArgsStr,
		fmt.Sprintf("%s=%s", VersionEnvKey, args.Version),
	)
	return pluginCommon.MergeDuplicateEnviron(env)
}
