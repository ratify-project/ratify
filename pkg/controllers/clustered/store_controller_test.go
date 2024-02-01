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
	"os"
	"strings"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/customresources/referrerstores"
	test "github.com/deislabs/ratify/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	storeName = "storeName"
	orasName  = "oras"
)

func TestStoreReconcile(t *testing.T) {
	dirPath, err := test.CreatePlugin(orasName)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

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
					Name: orasName,
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
					Name:    orasName,
					Address: dirPath,
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

func TestStoreAdd_InvalidConfig(t *testing.T) {
	resetStoreMap()
	var testStoreSpec = configv1beta1.ClusterStoreSpec{
		Name: orasName,
		Parameters: runtime.RawExtension{
			Raw: []byte("test"),
		},
	}

	if err := storeAddOrReplace(testStoreSpec, orasName); err == nil {
		t.Fatalf("storeAddOrReplace() expected error, actual %v", err)
	}
	if controllers.StoreMap.GetStoreCount() != 0 {
		t.Fatalf("Store map expected size 0, actual %v", controllers.StoreMap.GetStoreCount())
	}
}

func TestWriteStoreStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	testCases := []struct {
		name       string
		isSuccess  bool
		store      *configv1beta1.ClusterStore
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			store:      &configv1beta1.ClusterStore{},
			reconciler: &mockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			store:      &configv1beta1.ClusterStore{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &mockStatusClient{},
		},
		{
			name:      "status update failed",
			isSuccess: true,
			store:     &configv1beta1.ClusterStore{},
			reconciler: &mockStatusClient{
				updateFailed: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writeStoreStatus(context.Background(), tc.reconciler, tc.store, logger, tc.isSuccess, tc.errString)
		})
	}
}

func TestStoreAddOrReplace_PluginNotFound(t *testing.T) {
	resetStoreMap()
	var resource = "invalidplugin"
	expectedMsg := "plugin not found"
	var spec = getInvalidStoreSpec()
	err := storeAddOrReplace(spec, resource)

	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("TestStoreAddOrReplace_PluginNotFound expected msg: '%v', actual %v", expectedMsg, err.Error())
	}
}

func resetStoreMap() {
	controllers.StoreMap = referrerstores.NewActiveStores()
}

func getInvalidStoreSpec() configv1beta1.ClusterStoreSpec {
	return configv1beta1.ClusterStoreSpec{
		Name:    "pluginnotfound",
		Address: "test/path",
	}
}
