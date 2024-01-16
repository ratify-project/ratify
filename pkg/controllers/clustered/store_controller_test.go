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

package clustered

import (
	"context"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/controllers/test"
	"github.com/deislabs/ratify/pkg/customresources/referrerstores"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	storeName = "storeName"
)

func TestStoreReconcile(t *testing.T) {
	tests := []struct {
		name               string
		store              *configv1beta1.ClusterStore
		req                *reconcile.Request
		expectedErr        bool
		expectedStoreCount int
	}{
		{
			name: "nonexistent store",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			store: &configv1beta1.ClusterStore{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      storeName,
				},
				Spec: configv1beta1.ClusterStoreSpec{
					Name: "oras",
				},
			},
			expectedErr:        false,
			expectedStoreCount: 0,
		},
		{
			name: "valid spec",
			store: &configv1beta1.ClusterStore{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      storeName,
				},
				Spec: configv1beta1.ClusterStoreSpec{
					Name: "oras",
				},
			},
			expectedErr:        false,
			expectedStoreCount: 1,
		},
		{
			name: "invalid parameters",
			store: &configv1beta1.ClusterStore{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      storeName,
				},
				Spec: configv1beta1.ClusterStoreSpec{
					Parameters: runtime.RawExtension{
						Raw: []byte("test"),
					},
				},
			},
			expectedErr:        true,
			expectedStoreCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStoreMap()
			scheme, _ := test.CreateScheme()
			client := fake.NewClientBuilder().WithScheme(scheme)
			client.WithObjects(tt.store)
			r := &StoreReconciler{
				Scheme: scheme,
				Client: client.Build(),
			}
			var req reconcile.Request
			if tt.req != nil {
				req = *tt.req
			} else {
				req = reconcile.Request{
					NamespacedName: test.KeyFor(tt.store),
				}
			}

			_, err := r.Reconcile(context.Background(), req)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("Reconcile() expected error %v, actual %v", tt.expectedErr, err)
			}
			if controllers.StoreMap.GetStoreCount() != tt.expectedStoreCount {
				t.Fatalf("Store map expected size %v, actual %v", tt.expectedStoreCount, controllers.StoreMap.GetStoreCount())
			}
		})
	}
}

func TestStoreSetupWithManager(t *testing.T) {
	scheme, err := test.CreateScheme()
	if err != nil {
		t.Fatalf("CreateScheme() expected no error, actual %v", err)
	}
	client := fake.NewClientBuilder().WithScheme(scheme)
	r := &StoreReconciler{
		Scheme: scheme,
		Client: client.Build(),
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		t.Fatalf("NewManager() expected no error, actual %v", err)
	}

	if err := r.SetupWithManager(mgr); err != nil {
		t.Fatalf("SetupWithManager() expected no error, actual %v", err)
	}
}

func TestStoreAdd_InvalidConfig(t *testing.T) {
	resetStoreMap()
	var testStoreSpec = configv1beta1.ClusterStoreSpec{
		Name: "oras",
		Parameters: runtime.RawExtension{
			Raw: []byte("test"),
		},
	}

	if err := storeAddOrReplace(testStoreSpec, "oras"); err == nil {
		t.Fatalf("storeAddOrReplace() expected error, actual %v", err)
	}
	if controllers.StoreMap.GetStoreCount() != 0 {
		t.Fatalf("Store map expected size 0, actual %v", controllers.StoreMap.GetStoreCount())
	}
}

func resetStoreMap() {
	controllers.StoreMap = referrerstores.NewActiveStores()
}
