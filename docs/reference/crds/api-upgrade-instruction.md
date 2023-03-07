It's normal and inevitable to move CRD APIs to newer versions as the project 
moves to a more stable stage. This doc lists the steps for the first version 
bump-up from `v1alpha1` to `v1beta1`, which can be referenced when we need to have 
new API versions in the future.

## Upgrade steps from `v1alpha1` to `v1beta1`
1.  Controller-runtime models conversion between versions in terms of a 
"hub and spoke" model. In Ratify, we mark the `unversioned` as the hub version, 
other versions(`v1alpha1` and `v1beta1`) as spoke versions. New versions would 
be added as spoke versions.

2. Create new API version by `kubebuilder` command:
```bash
kubebuilder create api --group config.ratify.deislabs.io --version v1beta1 --kind <kind>
```
kind could be `Store`, `Verifier` and `CertificateStore` respectively.

3. Copy over existing types from `v1alpha1` to `v1beta1`.

4. Create an `unversioned` API by manually copying the existing types from `v1alpha1` to 
`unversioned` as `kubebuilder` doesn't support `unversioned` as version value.

5. In each spoke version package, add marker `+k8s:conversion-gen` directive 
pointing to the hub(`unversioned`) version, which must be in `doc.go`. Example:
```go
// +k8s:conversion-gen=github.com/deislabs/ratify/api/unversioned
package v1alpha1
```

6. In hub(`unversioned`) version package, create `doc.go` and add marker `+kubebuilder:object:generate=true` so that the object generator can use it. Example:
```go
package unversioned
// +kubebuilder:object:generate=true
```

7. In spoke version packages, add a `localSchemeBuilder = runtime.NewSchemeBuilder(SchemeBuilder.AddToScheme)` in `groupversion_info.go` so the auto-generated code 
could compile.

8. In hub(`unversioned`) version package, add marker `+kubebuilder:skip` to each 
API and remove all other markers so that skip kubebuilder processing it.

9. Mark `v1beta1` as the storage version by adding marker `+kubebuilder:storageversion` 
to the root types. Example:
```go
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:storageversion
// Store is the Schema for the stores API
type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreSpec   `json:"spec,omitempty"`
	Status StoreStatus `json:"status,omitempty"`
}
```

10. In the outdated spoke version package, add marker `+kubebuilder:deprecatedversion:warning=<msg>` to the root type of each API. Example:
```go
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:deprecatedversion:warning="v1alpha1 of the Store API has been deprecated. Please migrate to v1beta1."
// Store is the Schema for the stores API
type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   StoreSpec   `json:"spec,omitempty"`
	Status StoreStatus `json:"status,omitempty"`
}
```

11. Run `make manifests generate` to generate CRD objects, DeepCopy methods and conversion methods. `zz_generated.conversion.go` and `zz_generated.deepcopy.go` 
would be generated in each spoke version package.

## References
[Kubebuilder tutorial on multi-version API](https://book.kubebuilder.io/multiversion-tutorial/api-changes.html)

[Guide for API conversions](https://cluster-api-ibmcloud.sigs.k8s.io/developer/conversion.html)

[Approach to using conversion-gen with multi-versioning](https://github.com/kubernetes-sigs/kubebuilder/issues/1529#issuecomment-656359330)

[Example PR following same steps](https://github.com/Azure/eraser/pull/544)