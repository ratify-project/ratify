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

package store

import (
	"fmt"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify/v2/internal/store/factory"
	_ "github.com/notaryproject/ratify/v2/internal/store/factory/filesystemocistore" // Register the filesystem store factory
	_ "github.com/notaryproject/ratify/v2/internal/store/factory/registrystore"      // Register the registry store factory
)

// NewStore creates a new StoreMux instance.
func NewStore(opts []*factory.NewStoreOptions, globalScopes []string) (ratify.Store, error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("no store options provided")
	}
	storeMux := ratify.NewStoreMux()
	for _, storeOptions := range opts {
		if len(storeOptions.Scopes) == 0 {
			// if no scopes are provided, use the global scopes of the executor.
			storeOptions.Scopes = globalScopes
		}
		store, err := factory.NewStore(storeOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create store for type %q: %w", storeOptions.Type, err)
		}
		for _, scope := range storeOptions.Scopes {
			if err = storeMux.Register(scope, store); err != nil {
				return nil, fmt.Errorf("failed to register store for scope %q: %w", scope, err)
			}
		}
	}

	return storeMux, nil
}
