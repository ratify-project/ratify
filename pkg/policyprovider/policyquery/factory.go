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

package policyquery

import "fmt"

var policyQueryFactories = make(map[string]Factory)

// Config is a configuration for a policy query.
type Config struct {
	Name   string
	Policy string
}

// Factory is an interface for creating policy queries.
type Factory interface {
	Create(policy string) (PolicyQuery, error)
}

// Register adds the factory to the built-in policyQueryies map.
func Register(name string, factory Factory) {
	if factory == nil {
		panic("policy query factory cannot be nil")
	}
	_, registered := policyQueryFactories[name]
	if registered {
		panic(fmt.Sprintf("policy query factory named %s already registered", name))
	}

	policyQueryFactories[name] = factory
}

// CreateQueryFromConfig creates a policy query from the provided configuration.
func CreateQueryFromConfig(queryConfig Config) (PolicyQuery, error) {
	policyQueryName := queryConfig.Name
	if policyQueryName == "" {
		return nil, fmt.Errorf("policy query name must be specified")
	}

	factory, ok := policyQueryFactories[policyQueryName]
	if !ok {
		return nil, fmt.Errorf("policy query factory named %s not registered", policyQueryName)
	}

	policyQuery, err := factory.Create(queryConfig.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy query, err: %+w", err)
	}

	return policyQuery, nil
}
