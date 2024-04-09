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

package main

import (
	"os"

	"github.com/deislabs/ratify/cmd/ratify/cmd"
	_ "github.com/deislabs/ratify/pkg/cache/dapr"                  // register dapr cache
	_ "github.com/deislabs/ratify/pkg/cache/ristretto"             // register ristretto cache
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy" // register configpolicy policy provider
	_ "github.com/deislabs/ratify/pkg/policyprovider/regopolicy"   // register regopolicy policy provider
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"          // register oras referrer store
	_ "github.com/deislabs/ratify/pkg/verifier/cosign"             // register cosign verifier
	_ "github.com/deislabs/ratify/pkg/verifier/notation"           // register notation verifier
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
