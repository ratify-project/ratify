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

package namespaced

import (
	"context"
	"fmt"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/controllers/test"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/deislabs/ratify/pkg/controllers"
	cs "github.com/deislabs/ratify/pkg/customresources/certificatestores"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const certStoreName = "certStoreName"

func TestUpdateErrorStatus(t *testing.T) {
	var parametersString = "{\"certs\":{\"name\":\"certName\"}}"
	var certStatus = []byte(parametersString)

	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: true,
		Properties: runtime.RawExtension{
			Raw: certStatus,
		},
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}
	expectedErr := "it's a long error from unit test"
	lastFetchedTime := metav1.Now()
	updateErrorStatus(&certStore, expectedErr, &lastFetchedTime)

	if certStore.Status.IsSuccess != false {
		t.Fatalf("Unexpected error, expected isSuccess to be false , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != expectedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedErr, certStore.Status.Error)
	}
	expectedBriedErr := fmt.Sprintf("%s...", expectedErr[:30])
	if certStore.Status.BriefError != expectedBriedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedBriedErr, certStore.Status.Error)
	}

	//make sure properties of last cached cert was not overridden
	if len(certStore.Status.Properties.Raw) == 0 {
		t.Fatalf("Unexpected properties,  expected %+v, got %+v", parametersString, string(certStore.Status.Properties.Raw))
	}
}

func TestUpdateSuccessStatus(t *testing.T) {
	certStatus := certificateprovider.CertificatesStatus{}
	properties := map[string]string{}
	properties["CertName"] = "wabbit"
	properties["Version"] = "ABC"

	certStatus["Certificates"] = properties

	lastFetchedTime := metav1.Now()

	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}

	updateSuccessStatus(&certStore, &lastFetchedTime, certStatus)

	if certStore.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", certStore.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(certStore.Status.Properties.Raw) == 0 {
		t.Fatalf("Properties should not be empty")
	}
}

func TestUpdateSuccessStatus_emptyProperties(t *testing.T) {
	lastFetchedTime := metav1.Now()
	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}

	updateSuccessStatus(&certStore, &lastFetchedTime, nil)

	if certStore.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", certStore.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(certStore.Status.Properties.Raw) != 0 {
		t.Fatalf("Properties should be empty")
	}
}

func TestCertStoreSetupWithManager(t *testing.T) {
	scheme, err := test.CreateScheme()
	if err != nil {
		t.Fatalf("CreateScheme() expected no error, actual %v", err)
	}
	client := fake.NewClientBuilder().WithScheme(scheme)
	r := &CertificateStoreReconciler{
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

func TestCertStoreReconcile(t *testing.T) {
	tests := []struct {
		name               string
		certStore          *configv1beta1.CertificateStore
		req                *reconcile.Request
		expectedErr        bool
		expectedCertsCount int
	}{
		{
			name: "nonexistent cert store",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			certStore: &configv1beta1.CertificateStore{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      certStoreName,
				},
			},
			expectedErr:        false,
			expectedCertsCount: 0,
		},
		{
			name: "empty cert store",
			certStore: &configv1beta1.CertificateStore{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      certStoreName,
				},
				Spec: configv1beta1.CertificateStoreSpec{},
			},
			expectedErr:        true,
			expectedCertsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCertStoreMap()
			scheme, err := test.CreateScheme()
			if err != nil {
				t.Fatalf("CreateScheme() expected no error, actual %v", err)
			}
			client := fake.NewClientBuilder().WithScheme(scheme)
			client.WithObjects(tt.certStore)
			r := &CertificateStoreReconciler{
				Scheme: scheme,
				Client: client.Build(),
			}
			var req reconcile.Request
			if tt.req != nil {
				req = *tt.req
			} else {
				req = reconcile.Request{
					NamespacedName: test.KeyFor(tt.certStore),
				}
			}

			_, err = r.Reconcile(context.Background(), req)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("Reconcile() expected error %v, actual %v", tt.expectedErr, err)
			}
			certs := controllers.CertificatesMap.GetCertStores(testNamespace)
			if len(certs) != tt.expectedCertsCount {
				t.Fatalf("Cert map expected size %v, actual %v", tt.expectedCertsCount, len(certs))
			}
		})
	}
}

func resetCertStoreMap() {
	controllers.CertificatesMap = cs.NewActiveCertStores()
}
