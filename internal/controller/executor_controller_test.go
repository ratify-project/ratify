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

// revive:disable:dot-imports
package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv2alpha1 "github.com/notaryproject/ratify/v2/api/v2alpha1"
)

var _ = Describe("Executor Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "",
		}
		executor := &configv2alpha1.Executor{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Executor")
			err := k8sClient.Get(ctx, typeNamespacedName, executor)
			if err != nil && errors.IsNotFound(err) {
				resource := &configv2alpha1.Executor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "",
					},
					Spec: configv2alpha1.ExecutorSpec{
						Scopes: []string{"scope1", "scope2"},
						Verifiers: []*configv2alpha1.VerifierOptions{
							{
								Name: "notation-1",
								Type: "notation",
							},
						},
						Stores: []*configv2alpha1.StoreOptions{
							{
								Type: "registryStore",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &configv2alpha1.Executor{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Executor")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ExecutorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			updatedExecutor := &configv2alpha1.Executor{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedExecutor)).To(Succeed())
			Expect(updatedExecutor.Status.Succeeded).To(BeTrue())
		})
	})
})
