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

import "github.com/deislabs/ratify/pkg/referrerstore"

type stores struct {
	GlobalStores     []referrerstore.ReferrerStore
	NamespacedStores []referrerstore.ReferrerStore
}

func (s *stores) Stores(scope string) []referrerstore.ReferrerStore {
	if scope == "" {
		return s.GlobalStores
	}
	return s.NamespacedStores
}

func (s *stores) AddStores(scope string, stores []referrerstore.ReferrerStore) {
	if scope == "" {
		s.GlobalStores = append(s.GlobalStores, stores...)
	} else {
		s.NamespacedStores = append(s.NamespacedStores, stores...)
	}
}
