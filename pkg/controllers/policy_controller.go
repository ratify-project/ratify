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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type policy struct {
	// The name of the policy.
	Name string
	// The policy enforcer making a decision.
	Enforcer policyprovider.PolicyProvider
}

// ActivePolicy is the active policy generated from CRD. There would be exactly
// one active policy at any given time.
var ActivePolicy policy

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=policies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=policies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	policyLogger := logrus.WithContext(ctx)

	var policy configv1beta1.Policy
	var resource = req.Name
	policyLogger.Infof("Reconciling Policy %s", resource)

	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if apierrors.IsNotFound(err) && resource == ActivePolicy.Name {
			policyLogger.Infof("delete event detected, removing policy %s", resource)
			ActivePolicy.deletePolicy(resource)
		} else {
			policyLogger.Error("failed to get Policy: ", err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if resource != constants.RatifyPolicy {
		errStr := fmt.Sprintf("metadata.name must be ratify-policy, got %s", resource)
		policyLogger.Error(errStr)
		writePolicyStatus(ctx, r, &policy, policyLogger, false, errStr)
		return ctrl.Result{}, nil
	}

	if err := policyAddOrReplace(policy.Spec); err != nil {
		policyLogger.Error("unable to create policy from policy crd: ", err)
		writePolicyStatus(ctx, r, &policy, policyLogger, false, err.Error())
		return ctrl.Result{}, err
	}

	writePolicyStatus(ctx, r, &policy, policyLogger, true, "")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.Policy{}).
		Complete(r)
}

func policyAddOrReplace(spec configv1beta1.PolicySpec) error {
	policyEnforcer, err := specToPolicyEnforcer(spec)
	if err != nil {
		return fmt.Errorf("failed to create policy enforcer: %w", err)
	}

	ActivePolicy.Name = spec.Type
	ActivePolicy.Enforcer = policyEnforcer
	return nil
}

func specToPolicyEnforcer(spec configv1beta1.PolicySpec) (policyprovider.PolicyProvider, error) {
	policyConfig, err := rawToPolicyConfig(spec.Parameters.Raw, spec.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy config: %w", err)
	}

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(policyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy provider: %w", err)
	}

	return policyEnforcer, nil
}

func rawToPolicyConfig(raw []byte, policyName string) (config.PoliciesConfig, error) {
	pluginConfig := config.PolicyPluginConfig{}

	if string(raw) == "" {
		return config.PoliciesConfig{}, fmt.Errorf("no policy parameters provided")
	}
	if err := json.Unmarshal(raw, &pluginConfig); err != nil {
		return config.PoliciesConfig{}, fmt.Errorf("unable to decode policy parameters.Raw: %s, err: %w", raw, err)
	}

	pluginConfig["name"] = policyName

	return config.PoliciesConfig{
		PolicyPlugin: pluginConfig,
	}, nil
}

func (p *policy) deletePolicy(resource string) {
	if p.Name == resource {
		p.Name = ""
		p.Enforcer = nil
	}
}

// IsEmpty returns true if there is no policy set up.
func (p *policy) IsEmpty() bool {
	return p.Name == "" && p.Enforcer == nil
}

func writePolicyStatus(ctx context.Context, r client.StatusClient, policy *configv1beta1.Policy, logger *logrus.Entry, isSuccess bool, errString string) {
	if isSuccess {
		updatePolicySuccessStatus(policy)
	} else {
		updatePolicyErrorStatus(policy, errString)
	}
	if statusErr := r.Status().Update(ctx, policy); statusErr != nil {
		logger.Error(statusErr, ", unable to update policy error status")
	}
}

func updatePolicySuccessStatus(policy *configv1beta1.Policy) {
	policy.Status.IsSuccess = true
	policy.Status.Error = ""
	policy.Status.BriefError = ""
}

func updatePolicyErrorStatus(policy *configv1beta1.Policy, errString string) {
	briefErr := errString
	if len(errString) > maxBriefErrLength {
		briefErr = fmt.Sprintf("%s...", errString[:maxBriefErrLength])
	}
	policy.Status.IsSuccess = false
	policy.Status.Error = errString
	policy.Status.BriefError = briefErr
}
