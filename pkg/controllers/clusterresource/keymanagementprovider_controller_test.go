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

package clusterresource

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ratify-project/ratify/pkg/keymanagementprovider/refresh"
	test "github.com/ratify-project/ratify/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKeyManagementProviderReconciler_ReconcileWithConfig(t *testing.T) {
	tests := []struct {
		name              string
		refresherType     string
		createConfigError bool
		refreshError      bool
		expectedError     bool
	}{
		{
			name:              "Successful Reconcile",
			refresherType:     "mockRefresher",
			createConfigError: false,
			refreshError:      false,
			expectedError:     false,
		},
		{
			name:              "Refresher Error",
			refresherType:     "mockRefresher",
			createConfigError: false,
			refreshError:      true,
			expectedError:     true,
		},
		{
			name:              "Invalid Refresher Type",
			refresherType:     "invalidRefresher",
			createConfigError: true,
			refreshError:      false,
			expectedError:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := ctrl.Request{
				NamespacedName: client.ObjectKey{
					Name:      "fake-name",
					Namespace: "fake-namespace",
				},
			}
			scheme, _ := test.CreateScheme()
			client := fake.NewClientBuilder().WithScheme(scheme).Build()

			r := &KeyManagementProviderReconciler{
				Client: client,
				Scheme: runtime.NewScheme(),
			}

			refresherConfig := map[string]interface{}{
				"type":              tt.refresherType,
				"client":            client,
				"request":           req,
				"createConfigError": tt.createConfigError,
				"refreshError":      tt.refreshError,
				"shouldError":       tt.expectedError,
			}

			_, err := r.ReconcileWithConfig(context.TODO(), refresherConfig)
			if tt.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
func TestKeyManagementProviderReconciler_Reconcile(t *testing.T) {
	req := ctrl.Request{
		NamespacedName: client.ObjectKey{
			Name:      "fake-name",
			Namespace: "fake-namespace",
		},
	}

	// Create a fake client and scheme
	scheme, _ := test.CreateScheme()
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &KeyManagementProviderReconciler{
		Client: client,
		Scheme: runtime.NewScheme(),
	}

	// Call the Reconcile method
	result, err := r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check the result
	expectedResult := ctrl.Result{}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
}

type MockRefresher struct {
	Results           ctrl.Result
	CreateConfigError bool
	RefreshError      bool
	ShouldError       bool
}

func (mr *MockRefresher) Refresh(_ context.Context) error {
	if mr.RefreshError {
		return errors.New("refresh error")
	}
	return nil
}

func (mr *MockRefresher) GetResult() interface{} {
	return ctrl.Result{}
}

func (mr *MockRefresher) Create(config map[string]interface{}) (refresh.Refresher, error) {
	createConfigError := config["createConfigError"].(bool)
	refreshError := config["refreshError"].(bool)
	if createConfigError {
		return nil, errors.New("create error")
	}
	return &MockRefresher{
		CreateConfigError: createConfigError,
		RefreshError:      refreshError,
	}, nil
}

func init() {
	refresh.Register("mockRefresher", &MockRefresher{})
}
