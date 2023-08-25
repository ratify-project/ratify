# Store Plugin

## Overview

A referrer store in the framework is a component that can store and distribute OCI artifacts with the ability to discover reference types for a subject. It provides the following capabilities

- Retrieve list of referrers for an artifact
- Retrieve manifest for a referrer identified by the OCI reference
- Download the blobs of an artifact
- Retrieves properties of the subject manifest like descriptor to support verification.
- Retrieves the descriptor properties for a subject reference

This document proposes a generic plugin-based solution for integrating different stores into the framework.

### Terminology

The following terms will be used throughout the document

- **store** is the component that will be used to retrieve referrers for a subject and its related data like blobs and manifests.
- **framework** is the [verification framework](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both) that coordinates multiple verifiers with the underlying multiple stores
- **plugin** is a program that implements the store interface and provides the necessary data required for verification.

The following sections of the document aims to specify the interface between "framework" and "plugins"

> The key words "must", "must not", "required", "shall", "shall not", "should", "should not", "recommended", "may" and "optional" are used as specified in RFC 2119.

### Specification

 The store specification defines

- An interface that defines store capabilities
- The format of the configuration used to configure the store plugins.
- A protocol for the framework to make requests to the store plugins
- A process for executing plugins based on the provided configuration.
- Data types of the results returned by plugin to the framework
- Format of error response returned by plugin
- Version compatibility between framework and plugin

### Section1 : Store Configuration format

The framework can be configured with a set of  parameters that are used by both the framework and plugins. When a specific plugin is executed, its  corresponding parameters will be passed as an execution configuration to the plugin. The framework MAY support dynamic updates to the configuration as needed and hence it is recommended for the plugins to not consider this configuration as static and always use the config passed by the framework for execution.

#### Store Configuration format

The store configuration can be [TBD] YAML/JSON with the following properties

| Property | Type | IsRequired | Description |
| -------- | -------- | -------- | --------- |
| version     | string     | true     |The semantic version 2.0 of the store specification to which all configuration and data types conform. Currently it is 1.0.0|
| plugins     | array     | true     |The array of store plugins  and their configuration. This is a list of plugin configuration object described in the following section. |

##### Plugin Configuration objects

The following are the keys used to describe configuration of individual plugins.

| Property | Type | IsRequired | Description |
| -------- | -------- | -------- | --------- |
| name     | string     | true     |The name of the plugin that should match with plugin binary on disk. Must not contain characters disallowed in file paths for the system (e.g. / or \) |
| pluginBinDirs     | array     | false     |The list of paths to look for the plugin binary to execute. Default: the home path of the framework. |

Any other fields specified for a plugin other than the above mentioned are considered as opaque. The framework MUST preserve unknown fields and pass through these fields to the plugins at the time of execution. Plugins may define additional fields that they accept and may generate an error if called with unknown fields.

##### Example store configuration

```yml
stores:
  version: 1.0.0
  plugins:
  - name: ociregistry
    useHttp:  true
  - name: filesystem
    folderPath: "/home/user/artifacts"
```

### Section2 : Store Interface

The framework defines an interface for all the capabilities provided by a store.

An interace defined in ```golang```:

```go=
type ReferrerStore interface {
 Name() string
 ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (ListReferrersResult, error)
 // Used for small objects.
 GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error)
 GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error)
 GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error)
}

```

#### Name

The method is used to get the name of the store.

#### ListReferrers

This method is used to query all the referrers that are linked to a subject that is identified by a reference. The result will include the list of references and a continution token to support pagination.

#### GetBlobContent

[TBD] handle large objects through streaming?
This method is used to fetch the contents of a blob that is contained within the repository of the subject. The blob is identified by the given digest.

#### GetReferenceManifest

This method is used to fetch the contents of an [artifact manifest](https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md) that is identified by a reference descriptor. This reference descriptor is a referrer to the given subject.

#### GetSubjectDescriptor
This method is used to get the descriptor for a subject that includes digest, size and media type properties. 

### Section 3 : Plugin Based Store

The framework MUST provide a reference implementation of the store interface using the [plugin architecture](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both#Plugin-architecture) It will execute the configured plugins to implement the methods of the interface.

The rest of the sections of the document defines the protocol for executing the plugins to implement the different methods of the store interface.

### Section 4: Plugin Execution Protocol

The protocol is based on the execution of binaries invoked by the framework. The framework passes parameters to the plugin via environment variables and configuration. The configuration is supplied via ```stdin```. The plugin returns the result on ```stdout``` on success, or an error on ```stderr``` if the verification fails. Configuration and results are encoded using JSON format.
There are two types of inputs that are passed to the plugin. They are parameters which define invokation specific settings and the other is configuration that includes verifier and store configuration settings.

#### Parameters

Execution parameters are passed to the plugins via OS environment variables. The parameters that are passed to a store are defined below

- **RATIFY_STORE_COMMAND** indicates the operation to be executed. Currently they include ```LISTREFERRERS```, ```GETBLOB```, ```GETREFMANIFEST```, ```GETSUBJECTDESCRIPTOR```
- **RATIFY_STORE_SUBJECT** is the artifact under verification usually identified by a reference as per the OCI `{DNS/IP}/{Repository}:[name|digest]`
- **RATIFY_STORE_VERSION** is the version of the specification used between the framework and plugin. This value is taken from the ```version``` field of the store configuration.
- **RATIFY_STORE_ARGS**: Extra arguments passed in by the framework at invocation time. They are key-value pairs separated by semicolons; for example, "digest=sha256:sdfdsdss;nextToken=123;artifactTypes:type1,type2". If the command doesn't need any arguments, this can be empty

##### Operations & Parameters

The store specification defines 3 operations ```LISTREFERRERS```, ```GETBLOB```, ```GETREFMANIFEST``` The operation type is passed to the plugin via the **RATIFY_STORE_COMMAND** environemnt variable.

**```LISTREFERRERS```**: Get the list of referrers to the given subject. The arguments that are passed to this operation as part of *RATIFY_STORE_ARGS* are

- ```nextToken``` : (string) The continuation token obtained from the previous ```LISTREFERRERS``` oepration
- ```artifactTypes```: (string) Comma separated list of artifact types that are used for filtering the referrers.

**```GETBLOB```**: The arguments that are passed to this oepration as part of *RATIFY_STORE_ARGS* are

- ```digest``` : (string) The digest of the blob that has to be retrieved.

**```GETREFMANIFEST```**: The arguments that are passed to this oepration as part of *RATIFY_STORE_ARGS* are

- ```digest``` : (string) The digest of the artifact manifest that has to be retrieved.

#### Execution Configuration

When a plugin is registered using the [configuration](https://hackmd.io/wkAGhJgQSCaX4ln9CrXo_w#Plugin-Configuration-objects), the framework interprets the configuration per plugin and transforms it to a format that is expected by the plugin. This section describes the transformations made  by the framework  before the configuration is passed to the plugin.

The execution configuration for a plugin invocation is encoded in JSON. It will contain the plugin configuration that is provided by the user, primarily unchanged except for the specified additions

The execution configuration provided by the framework will contain the following fields.

- ```config``` : A JSON object representing the [plugin configuration](https://hackmd.io/wkAGhJgQSCaX4ln9CrXo_w#Plugin-Configuration-objects) provided as part of registration with the framework and passed unchanged.

### Section 5: Operations & Result Types

Plugins can return either a **Success** or **Error** result type.

#### Success

The store specification defines 3 operations ```LISTREFERRERS```, ```GETBLOB```, ```GETREFMANIFEST``` The success result types for each of these operations are defined below

**```LISTREFERRERS```**: Get the list of referrers to the given subject. The result JSON returned for this operation has the following properties

- ```referrers```: (list) A list of referrers to the given subject identified by their [descriptors](https://github.com/oras-project/artifacts-spec/blob/main/descriptor.md)
- ```nextToken``` : (string) The continuation token that SHOULD be used to get the next page of referrers.

**```GETBLOB```**: The content of the blob is returned as byte array via ```stdout```

**```GETREFMANIFEST```**: The content of the reference manifest is returned as byte array via ```stdout``

**```GETSUBJECTDESCRIPTOR```**: The contents of the descriptor for the subject is returned as a byte array via ```stdout```
#### Error

Plugins should output a JSON object with the following properties if they encounter an error

- ```code```: A numeric error code as described below
- ```msg```: A short message describing the error
- ```details```: More details describing the error.

[TODO] Add the error codes after the implementation

### Section 6: Plugin Implementation

The framework MAY provide libraries that can provide skeletons for writing plugins. These libraries can scaffold the parameter and configuration parsing and transformation and can define methods that the plugin writers can override for the implementation. These libraries also should catch any exceptions retruned from the plugins and return a proper error result to the framework. A simple CLI for example ```ratify plugin store add mystore``` to create a stub for a plugin using these libraries MAY be provided by the framework.

### Appendix : Examples

A example protocol sequence for a ```LISTREFERRERS``` operation is given below

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
  - name: notation
    artifactTypes: application/vnd.cncf.notary.signature
    verificationCerts:
    - "/home/user/.notary/keys/wabbit-networks.crt"
  - name: sbom
    artifactTypes: application/x.example.sbom.v0
    nestedReferences: application/vnd.cncf.notary.signature
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

2. An example subject ```registry.wabbit-networks.io:5000/net-monitor:signed@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb``` will be used to query for its referrers of type ```application/vnd.cncf.notary.signature```.
3. The framework calls the plugin ```ociregistry``` with the following environment variables

- **RATIFY_STORE_COMMAND** : ```LISTREFERRERS```
- **RATIFY_STORE_SUBJECT**: ```registry.wabbit-networks.io:5000/net-monitor:signed@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb```
- **RATIFY_STORE_VERSION**: ```1.0.0```
- **RATIFY_STORE_ARGS** : ```artifactTypes:application/vnd.cncf.notary.signature```

5. It calls the plugin with the following JSON execution configuration

```json=
{
    "config": {
        "name": "ociregistry",
        "useHttp": true
    }    
}
```

6. The  ```ociregistry``` queries the registry using the provided configuration and returns a following JSON result

```json=
{
    "referrers" : [
        {
          "mediaType": "application/vnd.cncf.oras.artifact.manifest.v1+json",
          "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
          "size": 7682,
          "artifactType": "application/vnd.cncf.notary.signature"
        }
    ],
    "nextToken":""
}
```
