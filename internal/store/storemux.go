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

	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/store/factory"
	_ "github.com/ratify-project/ratify/v2/internal/store/factory/filesystemocistore" // Register the filesystem store factory
	_ "github.com/ratify-project/ratify/v2/internal/store/factory/registrystore"      // Register the registry store factory
)

// PatternOptions defines a map of string keys to [factory.NewStoreOptions]
// values. Keys are patterns used to match artifact references.
type PatternOptions map[string]factory.NewStoreOptions

// NewStore creates a new StoreMux instance.
func NewStore(opts PatternOptions) (ratify.Store, error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("no store options provided")
	}
	storeMux := ratify.NewStoreMux()
	for pattern, storeOptions := range opts {
		store, err := factory.NewStore(storeOptions)
		if err != nil {
			return nil, err
		}
		if err = storeMux.Register(pattern, store); err != nil {
			return nil, err
		}
	}

	return storeMux, nil
}
