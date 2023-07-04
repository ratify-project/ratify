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
	"strings"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/deislabs/ratify/pkg/policyprovider/types"
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

type activePolicy struct {
	// The name of the policy.
	PolicyName string
	// The policy enforcer making a decision.
	PolicyEnforcer policyprovider.PolicyProvider
}

var (
	// PolicyEnforcer is the policy enforcer generated from CRD.
	ActivePolicy activePolicy
)

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
		if apierrors.IsNotFound(err) && resource == ActivePolicy.PolicyName {
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
		logrus.Error("unable to create policy provider: ", err)
		return err
	}

	ActivePolicy.PolicyName = policyName
	ActivePolicy.PolicyEnforcer = policyEnforcer
	return nil
}

func specToPolicyEnforcer(spec configv1beta1.PolicySpec, policyName string) (policyprovider.PolicyProvider, error) {
	if strings.ToLower(policyName) != types.RegoPolicy && strings.ToLower(policyName) != types.ConfigPolicy {
		return nil, fmt.Errorf("unknown policy type %s", policyName)
	}

	policyConfig, err := specToPolicyConfig(spec, policyName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy config: %w", err)
	}

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(policyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy provider: %w", err)
	}

	return policyEnforcer, nil
}

func specToPolicyConfig(spec configv1beta1.PolicySpec, policyName string) (config.PoliciesConfig, error) {
	pluginConfig := config.PolicyPluginConfig{}

	if string(spec.Parameters.Raw) == "" {
		return config.PoliciesConfig{}, fmt.Errorf("no policy parameters provided")
	}
	if err := json.Unmarshal(spec.Parameters.Raw, &pluginConfig); err != nil {
		logrus.Error("unable to decode policy parameters: ", err, " Parameters.Raw", spec.Parameters.Raw)
		return config.PoliciesConfig{}, err
	}

	pluginConfig["name"] = policyName

	return config.PoliciesConfig{
		PolicyPlugin: pluginConfig,
	}, nil
}

func (p *activePolicy) deletePolicy(resource string) {
	if p.PolicyName == resource {
		p.PolicyName = ""
		p.PolicyEnforcer = nil
	}
}

func (p *activePolicy) IsEmpty() bool {
	return p.PolicyName == "" && p.PolicyEnforcer == nil
}
