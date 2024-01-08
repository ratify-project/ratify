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

package referrerstores

import (
	"github.com/deislabs/ratify/pkg/referrerstore"
)

type ReferrrerStores interface {
	// Stores returns the list of referrer stores for the given scope.
	// For K8s users, scope can be either the namespace or empty string("") for cluster-wide scope.
	// For CLI users, scope is always empty string("") as there is no namespace specified.
	Stores(scope string) []referrerstore.ReferrerStore

	AddStores(scope string, stores []referrerstore.ReferrerStore)
}
