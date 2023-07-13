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

	if err := policyAddOrReplace(policy.Spec, resource); err != nil {
		policyLogger.Error("unable to create policy from policy crd: ", err)
		return ctrl.Result{}, err
	}

	// List all policies in the same namespace.
	policyList := &configv1beta1.PolicyList{}
	if err := r.List(ctx, policyList, client.InNamespace(req.Namespace)); err != nil {
		policyLogger.Error("failed to list Policies: ", err)
		return ctrl.Result{}, err
	}

	// Delete all policies except the current one.
	for _, item := range policyList.Items {
		item := item
		if item.Name != resource {
			policyLogger.Infof("Deleting policy %s", item.Name)
			err := r.Delete(ctx, &item)
			if err != nil {
				policyLogger.Error("failed to delete Policy: ", err)
				return ctrl.Result{}, err
			}
			policyLogger.Info("Deleted policy", "name", item.Name)
		}
	}

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.Policy{}).
		Complete(r)
}

func policyAddOrReplace(spec configv1beta1.PolicySpec, policyName string) error {
	policyEnforcer, err := specToPolicyEnforcer(spec, policyName)
	if err != nil {
		return fmt.Errorf("failed to create policy enforcer: %w", err)
	}

	ActivePolicy.Name = policyName
	ActivePolicy.Enforcer = policyEnforcer
	return nil
}

func specToPolicyEnforcer(spec configv1beta1.PolicySpec, policyName string) (policyprovider.PolicyProvider, error) {
	policyConfig, err := rawToPolicyConfig(spec.Parameters.Raw, policyName)
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
