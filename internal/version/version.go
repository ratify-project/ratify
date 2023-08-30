package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	// Version is the current version of Ratify.
	Version = ""

	// UserAgent is the user agent for http request header.
	UserAgent = ""

	//GitTag
	GitTag = ""

	//GitCommitHash is the full commit hash.
	GitCommitHash = ""

	//GitTreeState shows if the tree is unmodified or modified
	GitTreeState = ""
)

func init() {
	UserAgent = generateUserAgent()
}

func generateUserAgent() string {
	vcsrevision := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, v := range info.Settings {
			switch v.Key {
			case "vcs.revision":
				vcsrevision = v.Value
			}
		}
	}
	if Version == "" {
		Version = vcsrevision
	}
	return fmt.Sprintf("%s+%s (%s/%s)", "ratify", Version, runtime.GOOS, runtime.GOARCH)
}
