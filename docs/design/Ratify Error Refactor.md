Ratify Error Refactor
==
Author: Binbin Li (@binbin-li)

# Background
Ratify can generate various errors, either from its core workflow or from external plugins. Currently, these errors are simply returned with or without explanations, making it difficult for developers and users to understand the details and root cause of the errors. This results in slow debugging. To facilitate easier development and debugging, it is necessary to refactor the existing error handling system.

# Error Analysis
In Ratify, errors can originate from various components or sources. These include:

1. Plugins: Both built-in and external plugins can generate errors. Additionally, communication with upstream services, such as remote registries and Key Vaults, can also cause plugin errors.
2. Executor: Errors can occur while processing the executor workflow.
3. Servce/component init: During service start-up, many components are initialized, which can result in errors. Additionally, when users apply a new CR resource, errors may arise.

Currently users could check errors via 4 approaches.
1. Check the external data response from Ratify to Gatekeeper, include `verify` and `mutate` APIs.
2. Check the logs in the Ratify container.
3. If a new CR resource applied, check the CR description or details.
4. If Ratify is used as command line tool, errors would be printed out directly.

# Requirement
1. Provide more detailed information for most errors.
2. Make errors easier to locate within the codebase or documentation.
3. Make it easier to identify the component or source of an error.
4. For errors originating from plugins, support the identification of the specific plugin name.
5. Simplify the manipulation of error groups.
6. For errors originating from different cloud providers, include links to the corresponding provider’s website.


# Refactor Design
1. Wrap/convert most existing errors to custom errors.
2. For each plugin API that returns error, create a custom error.
3. Add error code and brief message to custom errors.
4. [optional] Add error details to custom errors to help users understand and fix errors. For the errors from cloud providers, attach links to corresponding documents.
5. Export custom errors so that users could check on Godoc.
6. Add component or source field to custom errors.
7. Add `PluginName` field to custom plugin errors.
8. For built-in plugins, add some fine-grained custom errors. For external plugins, add a general custom error.


## Error Definition
A custom error would be defined as below:
```go
type ErrorCode int
type Error struct {
	OriginalError error         `json:"originalError,omitempty"`
	Code          ErrorCode     `json:"code"`
	Message       string        `json:"message"`
	Detail        interface{}   `json:"detail,omitempty"`
	ComponentType ComponentType `json:"componentType,omitempty"`
	PluginName    string        `json:"pluginName,omitempty"`
	LinkToDoc     string        `json:"linkToDoc,omitempty"`
}
```
- OriginalError: the nested error that are wrapped by the current error. 
- Code: unique identifier of the error. It can be in the form of an integer or a string. e.g. 1001 or BAD_REQUEST
- Message: message describing the error.
- Details: detailed information about the error.
- ComponentType: the component that this error comes from. e.g. verifier
- PluginName: the name of the plugin. e.g. notary-verifier, cosign-verifier.
- LinkToDoc: a link to the documentation if this error is from a cloud provider.

Error functions:
```go
func (e Error) Is(target error) bool
func (e Error) ErrorCode() ErrorCode
func (e Error) Unwrap() error
func (e Error) Error() string
func (e Error) WithDetail(detail interface{}) Error
func (e Error) WithPluginName(pluginName string) Error
func (e Error) WithComponentType(componentType ComponentType) Error
func (e Error) WithError(err error) Error
func (e Error) WithLinkToDoc(link string) Error
func (e Error) IsEmpty() bool
```

The functions that prefixed with `With` adds a corresponding field to the error. e.g.
```go
err := Error{}.WithDetail("detail")
  .WithComponentType("verifier")
  .WithPluginName("notary-verifier")
  .WithLinkToDoc("link")
```

## Error classification
Most errors in Ratify can be divided into two categories. The first category includes errors related to plugins, whether external or built-in. The second category includes generic errors that occur within Ratify itself. 

### Plugin errors
Here are some common plugin errors:
#### Verifier errors
- ErrorCodeVerifyReferenceFailure
- ErrorCodeVerifySignatureFailure
- ErrorCodeSignatureNotFound
#### Referrer Store errors
- ErrorCodeListReferrersFailure
- ErrorCodeGetSubjectDescriptorFailure
- ErrorCodeGetReferenceManifestFailure
- ErrorCodeGetBlobContentFailure
- ErrorCodeReferrerStoreFailure
- ErrorCodeCreateRepositoryFailure
- ErrorCodeRepositoryOperationFailure
- ErrorCodeManifestInvalid
- ErrorCodeReferrersNotFound
#### Generic plugin errors
- ErrorCodePluginInitFailure
- ErrorCodePluginNotFound
- ErrorCodeDownloadPluginFailure
- ErrorCodeCertInvalid
- ErrorCodeProviderNotFound
- ErrorCodeKeyVaultOperationFailure

### Generic Ratify errors
Here are some common Ratify errors:
- ErrorCodeUnknown
- ErrorCodeExecutorFailure
- ErrorCodeBadRequest
- ErrorCodeReferenceInvalid
- ErrorCodeCacheNotSet
- ErrorCodeConfigInvalid
- ErrorCodeAuthDenied
- ErrorCodeEnvNotSet
- ErrorCodeClusterOperationFailure
- ErrorCodeHostNameInvalid
- ErrorCodeNoMatchingCredential
- ErrorCodeDataDecodingFailure
- ErrorCodeDataEncodingFailure

# Discussion
## Questions
1. Do we define too many errors? Should we remove some uncommon errors? `uncommon` means it's invoked less than 3 times.
  Discussed in PR review meeting. We'll keep errors but add a doc to describing all of them.
3. Is there a better way to compose a new error with different fields? The current way has to invoke `With` functions one by one, which seems a bit complex.