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
	NotationTsgLink    = "https://ratify.dev/docs/troubleshoot/verifier/notation"
	OrasLink           = "https://oras.land/"
	AuthProviderLink   = "https://ratify.dev/docs/reference/oras-auth-provider"
	PolicyProviderLink = "https://ratify.dev/docs/reference/providers"
	PolicyCRDLink      = "https://ratify.dev/docs/reference/crds/policies"
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
		Description: `Generic error returned when the executor fails to perform an operation. Please check the error details for more information.`,
	})

	// ErrorCodeBadRequest is returned if the request is not valid.
	ErrorCodeBadRequest = Register("errcode", ErrorDescriptor{
		Value:       "BAD_REQUEST",
		Message:     `bad request`,
		Description: `The request is invalid or malformed. Check the request body and headers for more details.`,
	})

	// ErrorCodeReferenceInvalid is returned if provided image reference is invalid.
	ErrorCodeReferenceInvalid = Register("errcode", ErrorDescriptor{
		Value:       "REFERENCE_INVALID",
		Message:     "reference invalid",
		Description: `Ratify failed to parse the given reference. Please verify the reference is in the correct format following docker convention: https://docs.docker.com/engine/reference/commandline/images/`,
	})

	// ErrorCodeCacheNotSet is returned if cache is not set successfully.
	ErrorCodeCacheNotSet = Register("errcode", ErrorDescriptor{
		Value:       "CACHE_NOT_SET",
		Message:     "cache not set",
		Description: `The cache is not set successfully. Check the error details.`,
	})

	// ErrorCodeConfigInvalid is returned if provided configuration is invalid.
	ErrorCodeConfigInvalid = Register("errcode", ErrorDescriptor{
		Value:       "CONFIG_INVALID",
		Message:     "config invalid",
		Description: `The config is invalid. Please validate your config.`,
	})

	// ErrorCodeAuthDenied is returned if authentication is denied.
	ErrorCodeAuthDenied = Register("errcode", ErrorDescriptor{
		Value:       "AUTH_DENIED",
		Message:     "auth denied",
		Description: `The authentication to required resource is denied. Please validate the credentials or configuration and check the detailed error.`,
	})

	// ErrorCodeEnvNotSet is returned if some environment variable is not set.
	ErrorCodeEnvNotSet = Register("errcode", ErrorDescriptor{
		Value:       "ENV_NOT_SET",
		Message:     "env not set",
		Description: `The required environment is not set. Please set it up properly.`,
	})

	// ErrorCodeGetClusterResourceFailure is returned if Ratify failed to get
	// required resources from the cluster.
	ErrorCodeGetClusterResourceFailure = Register("errcode", ErrorDescriptor{
		Value:       "GET_CLUSTER_RESOURCE_FAILURE",
		Message:     "get cluster resource failure",
		Description: "Ratify failed to get required resources from the cluster. Please validate if the resources exists in the cluster and the access is correctly assigned.",
	})

	// ErrorCodeHostNameInvalid is returned when parsed hostName is invalid.
	ErrorCodeHostNameInvalid = Register("errcode", ErrorDescriptor{
		Value:       "HOST_NAME_INVALID",
		Message:     "host name invalid",
		Description: "The registry hostname of given image or artifact is invalid. Please verify the registry hostname to ensure it can be correctly parsed",
	})

	// ErrorCodeNoMatchingCredential is returned if authProvider cannot find
	// matching credentials.
	ErrorCodeNoMatchingCredential = Register("errcode", ErrorDescriptor{
		Value:       "NO_MATCHING_CREDENTIAL",
		Message:     "no matching credential",
		Description: "No matching credential is found. Please verify the credentials is set up in K8s Secret.",
	})

	// ErrorCodeDataDecodingFailure is returned when it fails to decode data.
	ErrorCodeDataDecodingFailure = Register("errcode", ErrorDescriptor{
		Value:       "DATA_DECODING_FAILURE",
		Message:     "data decoding failure",
		Description: "Failed to decode data. Please verify the decoding data.",
	})

	// ErrorCodeDataEncodingFailure is returned when it fails to encode data.
	ErrorCodeDataEncodingFailure = Register("errcode", ErrorDescriptor{
		Value:       "DATA_ENCODING_FAILURE",
		Message:     "data encoding failure",
		Description: "Failed to encode data. Please verify the encoding data.",
	})

	// ErrorCodeKeyManagementConflict is returned when key management provider and certificate store are configured together.
	ErrorCodeKeyManagementConflict = Register("errcode", ErrorDescriptor{
		Value:       "KEY_MANAGEMENT_CONFLICT",
		Message:     "key management provider and certificate store should not be configured together",
		Description: "Key management provider and certificate store should not be configured together. Please migrate to key management provider and delete certificate store.",
	})
)
