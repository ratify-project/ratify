# Errors in Ratify
Ratify can generate various errors, either from its core workflow or from external/internal plugins. To facilitate easier debugging, Ratify will throw an error with a `code` property in most cases. This code can be used to identify the error and provide a more meaningful error message.

## Error Codes
### Generic errors from core workflow
| Code | Description | Details | Trouble Shooting Guide |
| ---- | ----------- | ------- | ---------------------- |
| EXECUTOR_FAILURE | executor failure | Generic error returned when the executor fails to perform an operation. | Check the original error for more details |
| BAD_REQUEST | bad request | The request is invalid or malformed. Check the request body and headers for more details. | - |
| REFERENCE_INVALID | reference invalid | The reference is invalid. Check the reference for more details. | - |
| CACHE_NOT_SET | cache not set | The cache is not set successfully. Check the cache for more details. | - |
| CONFIG_INVALID | config invalid | The config is invalid. Check the config for more details. | - |
| AUTH_DENIED | auth denied | The authentication is denied. Check the auth config for more details. | - |
| ENV_NOT_SET | env not set | The required environment is not set. Please set it up properly. | - |
| CLUSTER_OPERATION_FAILURE | cluster operation failure | The cluster operation fails. Check the cluster for more details. | - |
| HOST_NAME_INVALID | host name invalid | The host name is invalid. Check the host name for more details. | - |
| NO_MATCHING_CREDENTIAL | no matching credential| No matching credential is found. Check the credential for more details. | - |
| DATA_DECODING_FAILURE | data decoding failure | Failed to decode data. Check the data for more details. | - |
| DATA_ENCODING_FAILURE | data encoding failure | Failed to encode data. Check the data for more details. | - |

### Errors from external/built-in plugins

| Code | Description | Details | Trouble Shooting Guide |
| ---- | ----------- | ------- | ---------------------- |
| VERIFY_REFERENCE_FAILURE | verify reference failure | Generic error returned when the verifier fails to verify the reference. | Check the original error for more details |
| VERIFY_SIGNATURE_FAILURE | verify signature failure | Verifier failed to verify signature. | Check the signature for more details, or check the corresponding verifier spec. |
| SIGNATURE_NOT_FOUND | signature not found | No signature was found. Check the artifact for more details | - |
| LIST_REFERRERS_FAILURE | list referrers failure | Generic error returned when the referrer store fails to list the referrers. | Check the original error for more details. |
| GET_SUBJECT_DESCRIPTOR_FAILURE | get subject descriptor failure | Generic error returned when the referrer store fails to get the subject descriptor. | Check the original error for more details. |
| GET_REFERRER_MANIFEST_FAILURE | get reference manifest failure | Generic error returned when the referrer store fails to get the reference manifest. | Check the original error for more details. |
| GET_BLOB_CONTENT_FAILURE | get blob content failure | Generic error returned when the referrer store fails to get the blob content. | Check the original error for more details. |
| REFERRER_STORE_FAILURE | referrer store failure | Generic error returned when the referrer store fails to perform an operation. | Check the original error for more details. |
| CREATE_REPOSITORY_FAILURE | create repository failure | Failed to create repository. Check the repository config for more details. | - |
| REPOSITORY_OPERATION_FAILURE | repository operation failure | The repository operation fails. Check the repository for more details. | Please check the remote repository configuration and client configuration. |
| MANIFEST_INVALID | manifest invalid | The manifest is invalid. Check the manifest for more details. | - |
| REFERRERS_NOT_FOUND | referrers not found | No referrers are found. Check the artifact is configured correctly. | Please check if fetching artifact exists in the remote registry. |
| PLUGIN_INIT_FAILURE | plugin init failure | The plugin fails to initialize. Check the plugin config for more details. | - |
| PLUGIN_NOT_FOUND | plugin not found | No plugin was found. Check the plugin config for more details. | - |
| DOWNLOAD_PLUGIN_FAILURE | download plugin failure | Failed to download plugin. Check if the provided source is correct. | - |
| CERT_INVALID | cert invalid | The certificate is invalid. Check the certificate for more details. | - |
| PROVIDER_NOT_FOUND | provider not found | No provider was found. Check the provider config for more details. | - |
| KEY_VAULT_OPERATION_FAILURE | Key vault operation failed | Key vault operation failed. Check the key vault for more details. | - | 