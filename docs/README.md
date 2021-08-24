# Overview

This specification outlines a workflow of artifact verification which can be utilized by artifact consumers and provides a framework for the Notary project to be consumed by clients that consume these artifacts.

> This is a DRAFT specification and is a guidance for the prototype.

The HORA framework enables composition of various verifiers and referer stores to build verification framework into projects like GateKeeper. See [data-flow digram](#data-flow)

## Artifact references

An artifact is defined by a manifest and identified by a reference as per the OCI `{DNS/IP}/{Repository}:[name|digest]`. 
A reference artifact is linked to another artifact through discriptor and enables forming a chain of artifacts as shown below.

![artifact-hierarcy](./artifact-heirarchy.svg)

## Referrer Provider

A referrers provider provides a capability to query one or more underlying referrer stores to obtain a list of possible refererence artifacts of a particular artifact type.

### Capabilities

Referrers SHOULD support returning a collection of referrences from an underlying store.

> Update definitions once prototype is complete. This might evolve to just an interafce tos tart with

- `GET` ->  `localhost:5050/myrepository/{digest}/referrers?artifactType={notary.v2.signature}`
        > This should mirror the HTTP APIs for referrers but also provide support over a file driver
- Referrers SHOULD support obtaining the content for consumers requested through a descriptor that can be defined through a reference
- `GET` -> `localhost:5050/myrepository/blobs/{digest}`
- Server MUST reject verification requests for an un-registered `artifactType`

### Referrers Store

The Referreres Store provides an ability to compose multiple refer providers.

- Multiple refer providers MAY be registered through configuration.
- Configuration SHOULD ensure that there is atleast ONE referrer provider in the configuration.
- Order of query of the provider MAY be governed by order of registration or through configuration.
- Selection of a particular referrers store  is defined below.

> `MatchingLables` might become the construct that we chose to detemine which stores to query for references. 

### Reference Provider Policy

1. Referrer store may iterate through providers in ordinal sequence as per definition in configuration or registration
2. Reference artifacts may be returned from multiple provider
3. Artifacts references MAY be restricted to a set set of providers. Order or restriction does not ensure order of return of the references.

## Verfier Specification

A verifier provides the implementation for verification of a given artifact type.

- Notary Verifier supports `artifactType` of `notary.v2.signature`
- A verifier will provide `CanVerify` that will be called with an artifact type. All verifiers that respond true to CanVerify will be invoked in the order that are registered in the config.

## Executor

The executor is responsible for running the different verifiers .Multiple verifiers can be chained and executed one after the other.

- Order of execution MAY be prescribed by the sequence of verifiers that are registered.
- Verifiers MAY be registered with the executor dynamically at runtime or at configuration time.
- The executor MAY invoke the verifiers  synchronously.

- Should exception hanlder be a different handler or just another verifier (TBD)

### Policy Specification

The executor policy determines the chained outcome of execution of the multiple Verifiers through the Executor.

- This is scoped to a single reference store
- The policy engine should provide capability to determine if an artifact may `Continue on Errors` 
- The engine SHOULD support `Ignore failures` or `break glass` scenarios when one or more verifiers fail on verification.  (TBD)
- The policy engine MUST ensure that atleast one verifier per artifact type is successful.  
- Policy configuration should be registered before verification can be served.

## Data Flow

![workflow](./data-flow.svg)

## Composing Verifiers

The framework enables composition of multiple verifiers and referer stores so that artifacts of different types can be chained together. 

THe verification follows a delegation model where each verifier is responsible for verification of a type of artifact. 

**Nested verification** MAY be required by an artifact verifier for hierachical verification. For e.g. when verifying an SBOM, we first need to ensure that the SBOM signature is validated first or other nested artifacts.  

```yaml
IMAGE
└── SBOM
    └── SIGNATURE
```

> This currenlty needs to be defined in detail since there are considerations that need to be made the exit criteria for the tree-walk.

Consider the following composition registering a Signature and SBOM verifier. 

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
    nestedReferences: application/vnd.cncf.notary.v2
```

The `nv2verifier` binds to `cncf.notary.v2` and `sbom` verifier similarly binds to `x.example.sbom`. 

When `hora` encounters a signature of type `cncf.notary.v2` then the nv2 verifier configuration settings can be used to determine things like what keys are expected to be used. Scoping rules through `matchinglabels` can be used to determine the exact matching criteria of the policy that needs to be applied. 
> Matching Labels is currently being defined and how do we enforce how the SBOM verifier can indicate that the nested verifier needs to be executed first.
