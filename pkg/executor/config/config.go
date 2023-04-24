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

const PassthroughExecutionMode = "passthrough"

// ExecutorConfig represents the configuration for the executor
type ExecutorConfig struct {
	// ExecutionMode can be "passthrough" or empty.
	ExecutionMode string `json:"executionMode,omitempty"`
	// UseRegoPolicy is a flag to enable/disable rego policy.
	UseRegoPolicy bool `json:"useRegoPolicy,omitempty"`
	// Gatekeeper default verification webhook timeout is 3 seconds. 100ms network buffer added
	VerificationRequestTimeout *int `json:"verificationRequestTimeout"`
	// Gatekeeper default mutation webhook timeout is 1 seconds. 50ms network buffer added
	MutationRequestTimeout *int `json:"mutationRequestTimeout"`
	// TODO Add cache config
}
