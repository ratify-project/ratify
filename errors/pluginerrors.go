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

package errors

var (
	// errors related to Verifier plugins

	// ErrorCodeVerifyReferenceFailure is returned when verifier plugin fails to
	// to verify given reference.
	ErrorCodeVerifyReferenceFailure = Register("errcode", ErrorDescriptor{
		Value:       "VERIFY_REFERENCE_FAILURE",
		Message:     "verify reference failure",
		Description: `Verifier fails to verify the reference. Please check the error details for more information.`,
	})

	// ErrorCodeVerifyPluginFailure is returned when verifier plugin fails
	// to verify attached artifact.
	ErrorCodeVerifyPluginFailure = Register("errcode", ErrorDescriptor{
		Value:       "VERIFY_PLUGIN_FAILURE",
		Message:     "verify plugin failure",
		Description: "Verifier plugin failed to verify. Please check the error details from the verifier plugin and refer to plugin's documentation for more details.",
	})

	// ErrorCodeSignatureNotFound is returned when verifier cannot find a
	// signature.
	ErrorCodeSignatureNotFound = Register("errcode", ErrorDescriptor{
		Value:       "SIGNATURE_NOT_FOUND",
		Message:     "signature not found",
		Description: "No signature was found. Please validate the verifying artifact has attached any expected signatures.",
	})

	// errors related to Referrer Store plugins

	// ErrorCodeListReferrersFailure is returned when ListReferrers API fails.
	ErrorCodeListReferrersFailure = Register("errcode", ErrorDescriptor{
		Value:       "LIST_REFERRERS_FAILURE",
		Message:     "list referrers failure",
		Description: `Referrer store fails to list the referrers. Refer to https://ratify.dev/docs/reference/store#listreferrers for more details.`,
	})

	// ErrorCodeGetSubjectDescriptorFailure is returned when GetSubjectDescriptor
	// API fails.
	ErrorCodeGetSubjectDescriptorFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_SUBJECT_DESCRIPTOR_FAILURE",
		Message:     "get subject descriptor failure",
		Description: `Referrer store fails to get the subject descriptor. Refer to https://ratify.dev/docs/reference/store#getsubjectdescriptor for more details.`,
	})

	// ErrorCodeGetReferenceManifestFailure is returned when GetReferenceManifest
	// API fails.
	ErrorCodeGetReferenceManifestFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_REFERRER_MANIFEST_FAILURE",
		Message:     "get reference manifest failure",
		Description: `Referrer store fails to get the reference manifest. Refer to https://ratify.dev/docs/reference/store#getreferencemanifest for more details.`,
	})

	// ErrorCodeGetBlobContentFailure is returned when GetBlobContent API fails.
	ErrorCodeGetBlobContentFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_BLOB_CONTENT_FAILURE",
		Message:     "get blob content failure",
		Description: `Referrer store fails to get the blob content. Refer to https://ratify.dev/docs/reference/store#getblobcontent for more details.`,
	})

	// ErrorCodeReferrerStoreFailure is returned when a generic error happen in
	// ReferrerStore.
	ErrorCodeReferrerStoreFailure = Register("errcode", ErrorDescriptor{
		Value:       "REFERRER_STORE_FAILURE",
		Message:     "referrer store failure",
		Description: `Referrer store fails to perform an operation.`,
	})

	// ErrorCodeCreateRepositoryFailure is returned when Referrer Store fails to
	// create a repository object.
	ErrorCodeCreateRepositoryFailure = Register("errcode", ErrorDescriptor{
		Value:       "CREATE_REPOSITORY_FAILURE",
		Message:     "create repository failure",
		Description: "Failed to create repository. Please verify the repository config is configured correctly and check error details for more information.",
	})

	// ErrorCodeRepositoryOperationFailure is returned when a repository
	// operation fails.
	ErrorCodeRepositoryOperationFailure = Register("errcode", ErrorDescriptor{
		Value:       "REPOSITORY_OPERATION_FAILURE",
		Message:     "repository operation failure",
		Description: `The operation to the repository failed. Please check the error details for more information.`,
	})

	// ErrorCodeManifestInvalid is returned if fetched manifest is invalid.
	ErrorCodeManifestInvalid = Register("errcode", ErrorDescriptor{
		Value:       "MANIFEST_INVALID",
		Message:     "manifest invalid",
		Description: `The manifest is invalid. Please validate the manifest is correctly formatted.`,
	})

	// ErrorCodeNoVerifierReport is returned if there is no ReferrerStore set.
	ErrorCodeNoVerifierReport = Register("errcode", ErrorDescriptor{
		Value:       "NO_VERIFIER_REPORT",
		Message:     "no verifier report",
		Description: "No verifier report was generated. This might be due to various factors, such as lack of artifacts attached to the image, a misconfiguration in the Referrer Store preventing access to the registry, or the absence of appropriate verifiers corresponding to the referenced image artifacts.",
	})

	// Generic errors happen in plugins

	// ErrorCodePluginInitFailure is returned when executor or controller fails
	// to initialize a plugin.
	ErrorCodePluginInitFailure = Register("errcode", ErrorDescriptor{
		Value:       "PLUGIN_INIT_FAILURE",
		Message:     "plugin init failure",
		Description: "The plugin fails to be initialized. Please check error details and validate the plugin config is correctly provided.",
	})

	// ErrorCodePluginNotFound is returned when the executor cannot find the
	// required external plugin.
	ErrorCodePluginNotFound = Register("errcode", ErrorDescriptor{
		Value:       "PLUGIN_NOT_FOUND",
		Message:     "plugin not found",
		Description: "No plugin was found. Verify the required plugin is supported by Ratify and check the plugin name is entered correctly.",
	})

	// ErrorCodeDownloadPluginFailure is returned when executor fails to
	// download a required external plugin.
	ErrorCodeDownloadPluginFailure = Register("errcode", ErrorDescriptor{
		Value:       "DOWNLOAD_PLUGIN_FAILURE",
		Message:     "download plugin failure",
		Description: "Failed to download plugin. Please verify the provided plugin configuration is correct and check the error details for further investigation. Refer to https://ratify.dev/docs/reference/dynamic-plugins for more information.",
	})

	// ErrorCodeCertInvalid is returned when provided certificates are invalid.
	ErrorCodeCertInvalid = Register("errcode", ErrorDescriptor{
		Value:       "CERT_INVALID",
		Message:     "cert invalid",
		Description: "The certificate is invalid. Please verify the provided inline certificates or certificates fetched from key vault are in valid format. Refer to https://ratify.dev/docs/reference/crds/certificate-stores for more information.",
	})

	// ErrorCodeKeyInvalid is returned when provided key is invalid.
	// TODO: add website docs for this error code and update URL for error description
	ErrorCodeKeyInvalid = Register("errcode", ErrorDescriptor{
		Value:       "KEY_INVALID",
		Message:     "key invalid",
		Description: "The key is invalid. Please verify the provided inline key or key fetched from key vault is in valid format. Refer to [INPUT URL] for more information.",
	})

	// ErrorCodePolicyProviderNotFound is returned when a policy provider cannot
	// be found.
	ErrorCodePolicyProviderNotFound = Register("errcode", ErrorDescriptor{
		Value:       "POLICY_PROVIDER_NOT_FOUND",
		Message:     "policy provider not found",
		Description: "No provider was found. Please verify that the necessary policy provider has been registered and that the provider's name has been entered correctly.",
	})

	// ErrorCodeKeyVaultOperationFailure is returned when a key vault operation
	// fails.
	ErrorCodeKeyVaultOperationFailure = Register("errcode", ErrorDescriptor{
		Value:       "KEY_VAULT_OPERATION_FAILURE",
		Message:     "Key vault operation failed",
		Description: "Key vault operation failed. Please validate correct key vault configuration is provided or check the error details for further investigation.",
	})

	// ErrorCodeKeyManagementProviderFailure is returned when a key management provider operation fails.
	ErrorCodeKeyManagementProviderFailure = Register("errcode", ErrorDescriptor{
		Value:       "KEY_MANAGEMENT_PROVIDER_FAILURE",
		Message:     "Key management provider failure",
		Description: "Generic failure in key management provider. Please validate correct key management provider configuration is provided or check the error details for further investigation.",
	})
)
