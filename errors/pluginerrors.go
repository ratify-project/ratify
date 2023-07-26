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
		Description: `Generic error returned when the verifier fails to verify the reference.`,
	})

	// ErrorCodeVerifySignatureFailure is returned when upstream package fails
	// to verify attached signature.
	ErrorCodeVerifySignatureFailure = Register("errcode", ErrorDescriptor{
		Value:       "VERIFY_SIGNATURE_FAILURE",
		Message:     "verify signature failure",
		Description: "Verifier failed to verify signature. Check the signature for more details",
	})

	// ErrorCodeSignatureNotFound is returned when verifier cannot find a
	// signature.
	ErrorCodeSignatureNotFound = Register("errcode", ErrorDescriptor{
		Value:       "SIGNATURE_NOT_FOUND",
		Message:     "signature not found",
		Description: "No signature was found. Check the artifact for more details",
	})

	// errors related to Referrer Store plugins

	// ErrorCodeListReferrersFailure is returned when ListReferrers API fails.
	ErrorCodeListReferrersFailure = Register("errcode", ErrorDescriptor{
		Value:       "LIST_REFERRERS_FAILURE",
		Message:     "list referrers failure",
		Description: `Generic error returned when the referrer store fails to list the referrers.`,
	})

	// ErrorCodeGetSubjectDescriptorFailure is returned when GetSubjectDescriptor
	// API fails.
	ErrorCodeGetSubjectDescriptorFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_SUBJECT_DESCRIPTOR_FAILURE",
		Message:     "get subject descriptor failure",
		Description: `Generic error returned when the referrer store fails to get the subject descriptor.`,
	})

	// ErrorCodeGetReferenceManifestFailure is returned when GetReferenceManifest
	// API fails.
	ErrorCodeGetReferenceManifestFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_REFERRER_MANIFEST_FAILURE",
		Message:     "get reference manifest failure",
		Description: `Generic error returned when the referrer store fails to get the reference manifest.`,
	})

	// ErrorCodeGetBlobContentFailure is returned when GetBlobContent API fails.
	ErrorCodeGetBlobContentFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_BLOB_CONTENT_FAILURE",
		Message:     "get blob content failure",
		Description: `Generic error returned when the referrer store fails to get the blob content.`,
	})

	// ErrorCodeReferrerStoreFailure is returned when a generic error happen in
	// ReferrerStore.
	ErrorCodeReferrerStoreFailure = Register("errcode", ErrorDescriptor{
		Value:       "REFERRER_STORE_FAILURE",
		Message:     "referrer store failure",
		Description: `Generic error returned when the referrer store fails to perform an operation.`,
	})

	// ErrorCodeCreateRepositoryFailure is returned when Referrer Store fails to
	// create a repository object.
	ErrorCodeCreateRepositoryFailure = Register("errcode", ErrorDescriptor{
		Value:       "CREATE_REPOSITORY_FAILURE",
		Message:     "create repository failure",
		Description: "Failed to create repository. Check the repository config for more details.",
	})

	// ErrorCodeRepositoryOperationFailure is returned when a repository
	// operation fails.
	ErrorCodeRepositoryOperationFailure = Register("errcode", ErrorDescriptor{
		Value:       "REPOSITORY_OPERATION_FAILURE",
		Message:     "repository operation failure",
		Description: `The repository operation fails. Check the repository for more details.`,
	})

	// ErrorCodeManifestInvalid is returned if fetched manifest is invalid.
	ErrorCodeManifestInvalid = Register("errcode", ErrorDescriptor{
		Value:       "MANIFEST_INVALID",
		Message:     "manifest invalid",
		Description: `The manifest is invalid. Check the manifest for more details.`,
	})

	// ErrorCodeReferrersNotFound is returned if there is no ReferrerStore set.
	ErrorCodeReferrersNotFound = Register("errcode", ErrorDescriptor{
		Value:       "REFERRERS_NOT_FOUND",
		Message:     "referrers not found",
		Description: "No referrers are found. Check the artifact is configured correctly.",
	})

	// Generic errors happen in plugins

	// ErrorCodePluginInitFailure is returned when it fails to initialize a plugin.
	ErrorCodePluginInitFailure = Register("errcode", ErrorDescriptor{
		Value:       "PLUGIN_INIT_FAILURE",
		Message:     "plugin init failure",
		Description: "The plugin fails to initialize. Check the plugin config for more details.",
	})

	// ErrorCodePluginNotFound is returned when it cannot find the required plugin.
	ErrorCodePluginNotFound = Register("errcode", ErrorDescriptor{
		Value:       "PLUGIN_NOT_FOUND",
		Message:     "plugin not found",
		Description: "No plugin was found. Check the plugin config for more details.",
	})

	// ErrorCodeDownloadPluginFailure is returned when it fails to download a
	// plugin.
	ErrorCodeDownloadPluginFailure = Register("errcode", ErrorDescriptor{
		Value:       "DOWNLOAD_PLUGIN_FAILURE",
		Message:     "download plugin failure",
		Description: "Failed to download plugin. Check if the provided source is correct.",
	})

	// ErrorCodeCertInvalid is returned when provided certificates are invalid.
	ErrorCodeCertInvalid = Register("errcode", ErrorDescriptor{
		Value:       "CERT_INVALID",
		Message:     "cert invalid",
		Description: "The certificate is invalid. Check the certificate for more details.",
	})

	// ErrorCodeProviderNotFound is returned when a required provider cannot be
	// found.
	ErrorCodeProviderNotFound = Register("errcode", ErrorDescriptor{
		Value:       "PROVIDER_NOT_FOUND",
		Message:     "provider not found",
		Description: "No provider was found. Check the provider config for more details.",
	})

	ErrorCodeKeyVaultOperationFailure = Register("errcode", ErrorDescriptor{
		Value:       "KEY_VAULT_OPERATION_FAILURE",
		Message:     "Key vault operation failed",
		Description: "Key vault operation failed. Check the key vault for more details.",
	})
)
