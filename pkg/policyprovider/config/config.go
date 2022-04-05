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

package config

// PolicyPluginConfig represents the configuration of a policy plugin
type PolicyPluginConfig map[string]interface{}

// PoliciesConfig describes policies that are defined in the configuration
type PoliciesConfig struct {
	Version      string             `json:"version,omitempty"`
	PolicyPlugin PolicyPluginConfig `json:"policyPlugin"`
}
