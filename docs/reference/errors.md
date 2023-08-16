# Errors in Ratify
Ratify can generate various errors, either from its core workflow or from external/internal plugins. To facilitate easier debugging, Ratify will throw an error with a `code` property in most cases. This code can be used to identify the error and provide a more meaningful error message.

## Error Codes
### Errors from core workflow
#### System errors
| Code | Description | Trouble Shooting Guide |
| ---- | ----------- | ---------------------- |
| EXECUTOR_FAILURE | executor failure | Generic error returned when the executor fails to perform an operation. |
| BAD_REQUEST | bad request | The request is invalid or malformed. Check the request body and headers for more details. |
| AUTH_DENIED | auth denied | The authentication to required resource is denied. Please validate the credentials or configuration and check the detailed error. For auth configuration in oras store, please refer to [oras-auth-provider](oras-auth-provider.md).|
| GET_CLUSTER_RESOURCE_FAILURE | get cluster resource failure | Ratify failed to get required resources from the cluster. Please validate if the resources exists in the cluster and the access is correctly assigned. |
| DATA_DECODING_FAILURE | data decoding failure | Failed to decode data. Please verify the decoding data. |
| DATA_ENCODING_FAILURE | data encoding failure | Failed to encode data. Please verify the encoding data. |

#### Configuration errors
| Code | Description | Trouble Shooting Guide |
| ---- | ----------- | ---------------------- |
| REFERENCE_INVALID | reference invalid | Ratify failed to parse the given reference. Please verify the reference is in the correct format following docker convention: https://docs.docker.com/engine/reference/commandline/images/ |
| CACHE_NOT_SET | cache not set | The cache is not set successfully, please validate cache was created successfully during Ratify initialization. |
| CONFIG_INVALID | config invalid | The config is invalid. Please validate your config. |
| ENV_NOT_SET | env not set | The required environment is not set. Please set it up properly. |
| HOST_NAME_INVALID | host name invalid | The registry hostname of given image or artifact is invalid. Please verify the registry hostname to ensure it can be correctly parsed |
| NO_MATCHING_CREDENTIAL | no matching credential| Please verify the credentials is set up in K8s Secret. Please refer to [oras-auth-provider](./oras-auth-provider.md#3-kubernetes-secrets) for more details. |


### Errors from external/built-in plugins

#### Generic plugin errors
| Code | Description | Trouble Shooting Guide |
| ---- | ----------- | ---------------------- |
| PLUGIN_INIT_FAILURE | plugin init failure | The plugin fails to be initialized. Please check error details and validate the plugin config is correctly provided. |
| PLUGIN_NOT_FOUND | plugin not found | No plugin was found. Verify the required plugin is supported by Ratify and check the plugin name is entered correctly. |
| DOWNLOAD_PLUGIN_FAILURE | download plugin failure | Failed to download plugin. Please verify the provided plugin configuration is correct and check the error details for further investigation. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/dynamic-plugins.md for more information. |
| CERT_INVALID | cert invalid | The certificate is invalid. Please verify the provided inline certificates or certificates fetched from key vault are in valid format. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/crds/certificate-stores.md for more information. |
| PROVIDER_NOT_FOUND | provider not found | No provider was found. Please verify that the necessary policy provider has been registered and that the provider's name has been entered correctly. |

#### Verifier errors
| Code | Description | Trouble Shooting Guide |
| ---- | ----------- | ---------------------- |
| VERIFY_REFERENCE_FAILURE | verify reference failure | Generic error returned when the verifier fails to verify the reference. Please check the error details for more information. |
| VERIFY_SIGNATURE_FAILURE | verify signature failure | Verifier failed to verify signature. Please check the error details from the verifier plugin and refer to plugin's documentation for more details. |
| SIGNATURE_NOT_FOUND | signature not found | No signature was found. Please validate the image has signatures attached. |

#### Certificate Store errors

| Code | Description  | Trouble Shooting Guide |
| ---- | -----------  | ---------------------- |
| KEY_VAULT_OPERATION_FAILURE | key vault operation failed | Please validate correct key vault configuration is provided or check the error details for further investigation. Please review steps in [ratify-on-azure](../quickstarts/ratify-on-azure.md) for configuring Key Vault. | 

#### Referrer Store errors
| Code | Description  | Trouble Shooting Guide |
| ---- | -----------  | ---------------------- |
| LIST_REFERRERS_FAILURE | list referrers failure | Referrer store fails to list the referrers. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/store.md#listreferrers for more details. |
| GET_SUBJECT_DESCRIPTOR_FAILURE | get subject descriptor failure | Referrer store fails to get the subject descriptor. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/store.md#getsubjectdescriptor for more details. |
| GET_REFERRER_MANIFEST_FAILURE | get reference manifest failure | Referrer store fails to get the reference manifest. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/store.md#getreferencemanifest for more details. |
| GET_BLOB_CONTENT_FAILURE | get blob content failure | Referrer store fails to get the blob content. | Check the original error for more details. |
| REFERRER_STORE_FAILURE | referrer store failure | Referrer store fails to get the blob content. Refer to https://github.com/deislabs/ratify/blob/main/docs/reference/store.md#getblobcontent for more details. |
| CREATE_REPOSITORY_FAILURE | create repository failure | Failed to create repository. Please verify the repository config is configured correctly and check error details for more information. |
| REPOSITORY_OPERATION_FAILURE | repository operation failure | The operation to the repository failed. Please check the error details for more information. |
| MANIFEST_INVALID | manifest invalid | The manifest is invalid. Please validate the manifest is correctly formatted. |
| REFERRERS_NOT_FOUND | referrers not found | No referrers are found. Please verify the subject has attached expected artifacts and refer to https://github.com/deislabs/ratify/blob/main/docs/reference/store.md to investigate Referrer Store configuration. |

