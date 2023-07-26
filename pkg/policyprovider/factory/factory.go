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

package factory

import (
	"fmt"
	"strings"

	ratifyerrors "github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	"github.com/deislabs/ratify/pkg/verifier/types"
	"github.com/sirupsen/logrus"
)

var builtInPolicyProviders = make(map[string]PolicyFactory)

// PolicyFactory is an interface that defines methods to create a policy provider
type PolicyFactory interface {
	Create(policyConfig config.PolicyPluginConfig) (policyprovider.PolicyProvider, error)
}

// Register adds the factory to the built in providers map
func Register(name string, factory PolicyFactory) {
	if factory == nil {
		panic("policy factory cannot be nil")
	}
	_, registered := builtInPolicyProviders[name]
	if registered {
		panic(fmt.Sprintf("policy factory named %s already registered", name))
	}

	builtInPolicyProviders[name] = factory
}

// CreatePolicyProvidersFromConfig creates a policy provider from the provided configuration
func CreatePolicyProviderFromConfig(policyConfig config.PoliciesConfig) (policyprovider.PolicyProvider, error) {
	if policyConfig.PolicyPlugin == nil {
		return nil, ratifyerrors.ErrorCodeConfigInvalid.WithComponentType(ratifyerrors.PolicyProvider).WithDetail("policy provider config must be specified")
	}

	policyProviderName, ok := policyConfig.PolicyPlugin[types.Name]
	if !ok {
		return nil, ratifyerrors.ErrorCodeConfigInvalid.WithComponentType(ratifyerrors.PolicyProvider).WithDetail(fmt.Sprintf("failed to find policy provider name in the policy config with key: %s", types.Name))
	}

	providerNameStr := strings.ToLower(fmt.Sprintf("%s", policyProviderName))

	policyFactory, ok := builtInPolicyProviders[providerNameStr]
	if !ok {
		return nil, ratifyerrors.ErrorCodeProviderNotFound.WithComponentType(ratifyerrors.PolicyProvider).WithPluginName(providerNameStr).WithDetail("failed to find registered policy provider")
	}

	policyProvider, err := policyFactory.Create(policyConfig.PolicyPlugin)
	if err != nil {
		return nil, ratifyerrors.ErrorCodePluginInitFailure.WithComponentType(ratifyerrors.PolicyProvider).WithPluginName(providerNameStr).WithDetail("failed to create policy provider").WithError(err)
	}

	logrus.Infof("selected policy provider: %s", providerNameStr)
	return policyProvider, nil
}
