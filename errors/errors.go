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

const (
	AKVLink                   = "https://learn.microsoft.com/en-us/azure/key-vault/general/overview"
	AzureWorkloadIdentityLink = "https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview"
	AzureManagedIdentityLink  = "https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview"
	NotationSpecLink          = "https://github.com/notaryproject/notaryproject/tree/main/specs"
	OrasLink                  = "https://oras.land/"
)

var (
	// ErrorCodeUnknown is a generic error that can be used as a last
	// resort if there is no situation-specific error message that can be used
	ErrorCodeUnknown = Register("errcode", ErrorDescriptor{
		Value:       "UNKNOWN",
		Message:     "unknown error",
		Description: `Generic error returned when the error does not have an API classification.`,
	})

	// ErrorCodeExecutorFailure is returned when a generic error happen in
	// executor.
	ErrorCodeExecutorFailure = Register("errcode", ErrorDescriptor{
		Value:       "EXECUTOR_FAILURE",
		Message:     "executor failure",
		Description: `Generic error returned when the executor fails to perform an operation.`,
	})

	// ErrorCodeBadRequest is returned if the request is not valid.
	ErrorCodeBadRequest = Register("errcode", ErrorDescriptor{
		Value:       "BAD_REQUEST",
		Message:     `bad request`,
		Description: `The request is invalid or malformed. Check the request body and headers for more details.`,
	})

	// ErrorCodeReferenceInvalid is returned if provided reference is invalid.
	ErrorCodeReferenceInvalid = Register("errcode", ErrorDescriptor{
		Value:       "REFERENCE_INVALID",
		Message:     "reference invalid",
		Description: `The reference is invalid. Check the reference for more details.`,
	})

	// ErrorCodeCacheNotSet is returned if cache is not set successfully.
	ErrorCodeCacheNotSet = Register("errcode", ErrorDescriptor{
		Value:       "CACHE_NOT_SET",
		Message:     "cache not set",
		Description: `The cache is not set successfully. Check the cache for more details.`,
	})

	// ErrorCodeConfigInvalid is returned if provided configuration is invalid.
	ErrorCodeConfigInvalid = Register("errcode", ErrorDescriptor{
		Value:       "CONFIG_INVALID",
		Message:     "config invalid",
		Description: `The config is invalid. Check the config for more details.`,
	})

	// ErrorCodeAuthDenied is returned if authentication is denied.
	ErrorCodeAuthDenied = Register("errcode", ErrorDescriptor{
		Value:       "AUTH_DENIED",
		Message:     "auth denied",
		Description: `The authentication is denied. Check the auth config for more details.`,
	})

	// ErrorCodeEnvNotSet is returned if some environment variable is not set.
	ErrorCodeEnvNotSet = Register("errcode", ErrorDescriptor{
		Value:       "ENV_NOT_SET",
		Message:     "env not set",
		Description: `The required environment is not set. Please set it up properly.`,
	})

	// ErrorCodeClusterOperationFailure is returned if a Kubernetes cluster
	// operation fails.
	ErrorCodeClusterOperationFailure = Register("errcode", ErrorDescriptor{
		Value:       "CLUSTER_OPERATION_FAILURE",
		Message:     "cluster operation failure",
		Description: "The cluster operation fails. Check the cluster for more details.",
	})

	// ErrorCodeHostNameInvalid is returned when provided hostName is invalid.
	ErrorCodeHostNameInvalid = Register("errcode", ErrorDescriptor{
		Value:       "HOST_NAME_INVALID",
		Message:     "host name invalid",
		Description: "The host name is invalid. Check the host name for more details.",
	})

	// ErrorCodeNoMatchingCredential is returned if authProvider cannot find
	// matching credentials.
	ErrorCodeNoMatchingCredential = Register("errcode", ErrorDescriptor{
		Value:       "NO_MATCHING_CREDENTIAL",
		Message:     "no matching credential",
		Description: "No matching credential is found. Check the credential for more details.",
	})

	// ErrorCodeDataDecodingFailure is returned when it fails to decode data.
	ErrorCodeDataDecodingFailure = Register("errcode", ErrorDescriptor{
		Value:       "DATA_DECODING_FAILURE",
		Message:     "data decoding failure",
		Description: "Failed to decode data. Check the data for more details.",
	})

	// ErrorCodeDataEncodingFailure is returned when it fails to encode data.
	ErrorCodeDataEncodingFailure = Register("errcode", ErrorDescriptor{
		Value:       "DATA_ENCODING_FAILURE",
		Message:     "data encoding failure",
		Description: "Failed to encode data. Check the data for more details.",
	})
)
