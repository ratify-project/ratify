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
