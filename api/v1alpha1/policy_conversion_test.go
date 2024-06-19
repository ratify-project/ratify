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

package v1alpha1

import (
	"reflect"
	"testing"

	unversioned "github.com/ratify-project/ratify/api/unversioned"
	"github.com/ratify-project/ratify/internal/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

var params = runtime.RawExtension{}

const testPolicyType = "testPolicyType"

func TestConvert_unversioned_PolicySpec_To_v1alpha1_PolicySpec(t *testing.T) {
	in := &unversioned.PolicySpec{
		Parameters: params,
	}
	out := &PolicySpec{}
	if err := Convert_unversioned_PolicySpec_To_v1alpha1_PolicySpec(in, out, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out.Parameters, in.Parameters) {
		t.Fatalf("expect parameters to be equal, but got different values")
	}
}

func TestConvert_unversioned_PolicyStatus_To_v1alpha1_PolicyStatus(t *testing.T) {
	if err := Convert_unversioned_PolicyStatus_To_v1alpha1_PolicyStatus(nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConvert_unversioned_Policy_To_v1alpha1_Policy(t *testing.T) {
	in := &unversioned.Policy{
		Spec: unversioned.PolicySpec{
			Type: testPolicyType,
		},
	}
	out := &Policy{}
	if err := Convert_unversioned_Policy_To_v1alpha1_Policy(in, out, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ObjectMeta.Name != in.Spec.Type {
		t.Fatalf("expect metadata.name to be %s, but got %s", in.Spec.Type, out.ObjectMeta.Name)
	}
}

func TestConvert_v1alpha1_Policy_To_unversioned_Policy(t *testing.T) {
	in := &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: testPolicyType,
		},
	}
	out := &unversioned.Policy{}
	if err := Convert_v1alpha1_Policy_To_unversioned_Policy(in, out, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ObjectMeta.Name != constants.RatifyPolicy {
		t.Fatalf("expect metadata.name to be %s, but got %s", constants.RatifyPolicy, out.ObjectMeta.Name)
	}
	if out.Spec.Type != in.ObjectMeta.Name {
		t.Fatalf("expect spec.type to be %s, but got %s", in.ObjectMeta.Name, out.Spec.Type)
	}
}
