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

package policyengine

import "fmt"

var engineFactories = make(map[string]EngineFactory)

// Config represents the configuration for an OPA policy engine.
type Config struct {
	// Name is the name of the policy engine.
	Name string
	// QueryLanguage is the language of the policy query.
	QueryLanguage string
	// Query is the policy used for query.
	Policy string
}

// EngineFactory is an interface for creating OPA policy engines.
type EngineFactory interface {
	// Create creates a new engine.
	Create(policy string, queryLanguage string) (PolicyEngine, error)
}

// Register adds the factory to the built-in opaEngines map.
func Register(name string, factory EngineFactory) {
	if factory == nil {
		panic("opa engine factory cannot be nil")
	}
	_, registered := engineFactories[name]
	if registered {
		panic(fmt.Sprintf("opa engine factory named %s already registered", name))
	}

	engineFactories[name] = factory
}

// CreateEngineFromConfig creates an OPA policy engine from the provided configuration.
func CreateEngineFromConfig(engineConfig Config) (PolicyEngine, error) {
	engineName := engineConfig.Name
	if engineName == "" {
		return nil, fmt.Errorf("policy engine name must be specified")
	}

	factory, ok := engineFactories[engineName]
	if !ok {
		return nil, fmt.Errorf("policy engine factory named %s not registered", engineName)
	}

	engine, err := factory.Create(engineConfig.Policy, engineConfig.QueryLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy engine: %w", err)
	}

	return engine, nil
}
