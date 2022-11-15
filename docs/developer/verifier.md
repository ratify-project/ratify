# Verifier Specification

## Overview

A reference verifier in the framework is a component that is responsible to verify a specific artifact type(s) using the provided configuration. It provides the following capabilities

- Given an reference type artifact, determine if the verifier supports its verification like ```CanVerify```
- Verifies the reference artifact and returns the result of verification.

This document proposes a generic plugin-based solution for integrating different verifiers into the framework.

### Terminology

The following terms will be used throughout the document

- **verifier** is the component that verifies a reference for an artifact. This could be a signature verifier or SBOM  verifier for example.
- **referrer store** is the component that will be used to retrieve referrers for a subject and its related data like blobs and manifests.
- **framework** is the [verification framework](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both) that coordinates multiple verifier
- **plugin** is a program that implements the verifier interface and provides the verification result to the framework upon invocation.

The following sections of the document aims to specify the interface between "framework" and "plugins"

> The key words "must", "must not", "required", "shall", "shall not", "should", "should not", "recommended", "may" and "optional" are used as specified in RFC 2119.

### Specification

 The verifier specification defines

- An interface that defines verifier capabilities
- The format of the configuration used to configure the verifier plugins.
- A protocol for the framework to make requests to the verifier plugins
- A process for executing plugins based on the provided configuration.
- A process for plugins to access the referrer store to fetch additional data required for verification
- Data types of the results returned by plugin to the framework
- Format of error response returned by plugin
- Version compatibility between framework and plugin

### Section1 : Verifier Configuration format

The framework can be configured with a set of  parameters that are used by both the framework and plugins. When a specific plugin is executed, its  corresponding parameters will be passed as an execution configuration to the plugin. The framework MAY support dynamic updates to the configuration as needed and hence it is recommended for the plugins to not consider this configuration as static and always use the config passed by the framework for execution.

#### Verifier Configuration format

The verifier configuration is [TBD] YAML/JSON  with the following properties

| Property | Type | IsRequired | Description |
| -------- | -------- | -------- | --------- |
| version     | string     | true     |The semantic version 2.0 of the verifier specification to which all configuration and data types conform. Currently it is 1.0.0|
| plugins     | array     | true     |The array of verifier plugins  and their configuration. This is a list of plugin configuration object described in the following section. |

##### Plugin Configuration objects

The following are the keys used to describe configuration of  individual plugins.

| Property | Type | IsRequired | Description |
| -------- | -------- | -------- | --------- |
| name     | string     | true     |The name of the plugin that should match with plugin binary on disk. Must not contain characters disallowed in file paths for the system (e.g. / or \) |
| pluginBinDirs     | array     | false     |The list of paths to look for the plugin binary to execute. Default: the home path of the framework. |
| artifactTypes     | array     | true     |The list of artifact types for which this verifier plugin has to be executed. [TBD] May change to `matchingLabels` |

Any other fields specified for a plugin other than the above mentioned are considered as opaque. The framework MUST preserve unknown fields and pass through these fields to the plugins at the time of execution. Plugins may define additional fields that they accept and may generate an error if called with unknown fields.

##### Example verifier configuration

```yml
verifiers:
  version: 1.0.0
  plugins:
  - name: nv2verifier
    artifactTypes: application/vnd.cncf.notary.v2
    verificationCerts:
    - "/home/user/.notary/keys/wabbit-networks.crt"
  - name: sbom
    artifactTypes: application/x.example.sbom.v0
```

### Section2 : Verifier Interface

The framework defines an interface for all the capabilities provided by the verifier.

An interace defined in ```golang```:

```go=
type ReferenceVerifier interface {
 Name() string
 CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool
 Verify(ctx context.Context,
  subjectReference common.Reference,
  referenceDescriptor ocispecs.ReferenceDescriptor,
  referrerStore *referrerStore.Store
  ) (VerifierResult, error)
}

```

#### Name

The method is used to get the name of the verifier.

#### CanVerify

The framework will invoke this method of the verifier to determine if it supports verification of a given artifact reference.

#### Verify

If verifier acknowledges its support for a reference type, the framework will invoke this method on the verifier to trigger the verification of the artifact reference. In addition to the artifact reference that has to be verified, the framework MUST include the associated referrer store and the framework's execution engine as part of the invocation. This will enable the verifier to query additional data from the store as needed.

### Section 3 : Plugin Based Verifier

The framework MUST provide a reference implementation of the verifier interface using the [plugin architecture](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both#Plugin-architecture) It will execute the configured plugins to implement the methods of the interface.

The interface method ```CanVerify``` can be implemented by the framework using the verifier configuration without executing the plugin. It can use ```artifactTypes``` key (or ```matchingLabels```) to determine the support of a verifier plugin for a given artifact reference.

[Nested verification](./executor.md#nested-references-configuration-object) can also be handled by the framework's executor engine by supplying the ```artifactType``` as ```true``` in the executor's ```nestedReferences``` config.

The rest of the sections of the document defines the protocol for executing the plugins to implement the ```Verify``` method of the verifier interface.

### Section 4: Plugin Execution Protocol

The protocol is based on the execution of binaries invoked by the framework. The framework passes parameters to the plugin via environment variables and configuration. The configuration is supplied via ```stdin```. The plugin returns the result on ```stdout``` on success, or an error on ```stderr``` if the verification fails. Configuration and results are encoded using JSON format.
There are two types of inputs that are passed to the plugin. They are parameters which define invokation specific settings and the other is configuration that includes verifier and store configuration settings.

#### Parameters

Execution parameters are passed to the plugins via OS environment variables. The parameters that are passed to a verifier are defined below

- **RATIFY_VERIFIER_COMMAND** indicates the operation to be executed. Currently the only operation that is desired is ```VERIFY```
- **RATIFY_VERIFIER_SUBJECT** is the artifact under verification usually identified by a reference as per the OCI `{DNS/IP}/{Repository}:[name|digest]`
- **RATIFY_VERIFIER_VERSION** is the version of the specification used between the framework and plugin. This value is taken from the ```version``` field of the verifier configuration.

#### Execution Configuration

When a plugin is registered using the [configuration](https://hackmd.io/ReuaHddWSDq4gaKyUofa7w?both#Section1--Verifier-Configuration-format), the framework interprets the configuration per plugin and transforms it to a format that is expected by the plugin. This section describes the transformations made  by the framework  before the configuration is passed to the plugin.

The execution configuration for a plugin invocation is encoded in JSON. It will contain the plugin configuration that is provided by the user, primarily unchanged except for the specified additions

The execution configuration provided by the framework will contain the following fields.

- ```config``` : A JSON object representing the [plugin configuration](https://hackmd.io/ReuaHddWSDq4gaKyUofa7w?both#Plugin-Configuration-objects) provided as part of registration with the framework and passed unchanged.
- ```storeConfig``` : A JSON object representing the configuration of the [store](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both#Referrer-Store-Configuration) that will be used to create a store plugin to fetch additional data needed for verification
- ```referenceDesc``` : A JSON object that has the [descriptor](https://github.com/oras-project/artifacts-spec/blob/main/descriptor.md) properties of a reference type.

### Section 5: Result Types

Plugins can return either a **Success** or **Error** result type.

#### Success

The output of the verification process will be returned by the plugin. It MUST output a JSON object with the following properties upon successful  ```VERIFY``` operation

- ```isSuccess``` (bool) Indicates if the artifact is verified successfully or not.
- ```results```: (list of strings) A list of strings that describe the outcomes of the verification process.
- ```name```: (string) The name of the verifier plugin which matches with the name provided as part of the registration.

#### Error

Plugins should output a JSON object with the following properties if they encounter an error

- ```code```: A numeric error code as described below
- ```msg```: A short message describing the error
- ```details```: More details describing the error.

[TODO] Add the error codes after the implementation

### Section 5: Plugin Implementation

The framework MAY provide libraries that can provide skeletons for writing plugins. These libraries can scaffold the parameter and configuration parsing and transformation and can define methods that the plugin writers can override for the implementation. These libraries also should catch any exceptions retruned from the plugins and return a proper error result to the framework. A simple CLI for example ```ratify plugin verifier add myverifier``` to create a stub for a plugin using these libraries MAY be provided by the framework.

### Appendix : Examples

A example protocol sequence for a ```verify``` operation is given below

1. The following framework configuration will be used for the example ([TBD] need to decide if it is JSON or YAML)

```yml
stores:
  version: 1.0.0
  plugins:
  - name: ociregistry
    useHttp: true
verifiers:
  version: 1.0.0
  plugins:
  - name: nv2verifier
    artifactTypes: application/vnd.cncf.notary.v2
    verificationCerts:
    - "/home/user/.notary/keys/wabbit-networks.crt"
  - name: sbom
    artifactTypes: application/x.example.sbom.v0
executor:
  cache: false
policy:
  type: opa
  policy: |
    package ratify.rules
        
        verify_artifact{
            regex.match(".+.azurecr.io$", input.subject)
        }
```

2. The following descriptor fetched from the store ```ociregistry``` will be used as the reference for a subject ```registry.wabbit-networks.io:5000/net-monitor:signed@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb``` and will be verified

```json=
{
  "mediaType": "application/vnd.cncf.oras.artifact.manifest.v1+json",
  "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
  "size": 7682,
  "artifactType": "application/vnd.cncf.notary.v2"
}
```

3. The framework uses the ```artifactTypes``` property to match the verifier plugin for the above reference type. In this case, the verifier with the name ```nv2verifier``` supports its verification.
4. The framework calls the plugin ```nv2verifier``` with the following environment variables

- **RATIFY_VERIFIER_COMMAND** : ```VERIFY```
- **RATIFY_VERIFIER_SUBJECT**: ```registry.wabbit-networks.io:5000/net-monitor:signed@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb```
- **RATIFY_VERIFIER_VERSION**: ```1.0.0```

5. It calls the plugin with the following JSON execution configuration

```json=
{
"config" : {
    "name" : "nv2verifier",
    "artifactTypes": "[application/vnd.cncf.notary.v2]",
    "verificationCerts": ["/home/user/.notary/keys/wabbit-networks.crt"]    
},
"storeConfig": {
    "name": "ociregistry",
    "useHttp": true
},
"referenceDesc": {
  "mediaType": "application/vnd.cncf.oras.artifact.manifest.v1+json",
  "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
  "size": 7682,
  "artifactType": "application/vnd.cncf.notary.v2"
}
}
```

6. The  ```nv2verifier``` verifies the artifact using the provided configuration and returns a following JSON result

```json=
{
      "isSuccess": true,
      "name": "nv2verifier",
      "results": [
        "signature verification success"
      ]
    }
```
