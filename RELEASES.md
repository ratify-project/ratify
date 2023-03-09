# Ratify Release Process

This document describes the versioning scheme and release processes for Ratify.

## Attribution

The specification release process was created using content and verbiage from the following specifications:

* [ORAS Artifact Specification Releases](https://github.com/oras-project/artifacts-spec/blob/main/RELEASES.md)
* [ORAS Developer Guide](https://github.com/oras-project/oras-www/blob/main/docs/CLI/5_developer_guide.md)
* [Mystikos Release Management](https://github.com/deislabs/mystikos/blob/main/doc/releasing.md)

## Versioning

The Ratify project follows [semantic versioning](https://semver.org/) beginning with version `v0.1.0`.  Pre-release versions may be specified with a dash after the patch version and the following specifiers (in the order of release readiness):

* `alpha1`, `alpha2`, etc.
* `beta1`, `beta2`, etc.
* `rc1`, `rc2`, `rc3`, etc.

Example pre-release versions include `v0.1.0-alpha1`, `v0.1.0-beta2`, `v0.1.0-rc3`.  Pre-release versions are not required and stages can be bypassed (i.e. an `alpha` release does not require a `beta` release).  Pre-releases must be in order and gaps are not allowed (i.e. the only releases that can follow `rc1` are the full release or `rc2`).

## Pre Release Activity

Most e2e-scenarios for cli, K8, and Azure are covered by the ratify e2e tests. Please refer to this [document](test/validation.md) for the current supported and unsupported tests. 

Please perform manual prerelease validations for the unsupported tests list [here](test/validation.md#unsupported-tests)

Validate that the format of the data returned for external data calls has not changed. If it has changed update the version in `httpserver/types.go` to reflect a change in the format and document the update.

Delete all dev images generated since the previous release under the `ratify-dev` and `ratify-crds-dev` packages. Each dev image tag is prefixed with `dev` followed by the date of creation (e.g a build generated on March 2, 2023 would be tagged `dev.20230302`).
## Git Release Flow

This section deals with the practical considerations of versioning in Git, this repo's version control system.  See the semantic versioning specification for the scope of changes allowed for each release type.

### Patch releases

When a patch release is required, the patch commits should be merged with the `main` branch when ready.  Then a new branch should be created with the patch version incremented and optional pre-release specifiers.  For example if the previous release was `v0.1.0`, the branch should be named `v0.1.1` and can optionally be suffixed with a pre-release (e.g. `v0.1.1-rc1`).  The limited nature of fixes in a patch release should mean pre-releases can often be omitted.

### Minor releases

When a minor release is required, the release commits should be merged with the `main` branch when ready.  Then a new branch should be created with the minor version incremented and optional pre-release specifiers.  For example if the previous release was `v0.1.1`, the branch should be named `v0.2.0` and can optionally be suffixed with a pre-release (e.g. `v0.2.0-beta1`).  Pre-releases will be more common will be more common with minor releases.

### Major releases

When a major release is required, the release commits should be merged with the `main` branch when ready.  Then a new branch should be created with the major version incremented and optional pre-release specifiers.  For example if the previous release was `v1.1.1`, the branch should be named `v2.0.0` and can optionally be suffixed with a pre-release (e.g. `v2.0.0-alpha1`).  Major versions will usually require multiple pre-release versions.

### Tag and Release

When the release branch is ready, a tag should be pushed with a name matching the branch name, e.g. `git tag v0.1.0-alpha1` and `git push --tags`.  This will trigger a [Goreleaser](https://goreleaser.com/) action that will build the binaries and creates a [GitHub release](https://help.github.com/articles/creating-releases/):

* The release will be marked as a draft to allow an final editing before publishing.
* The release notes and other fields can edited after the action completes.  The description can be in Markdown.
* The pre-release flag will be set for any release with a pre-release specifier.
* The pre-built binaries are built from commit at the head of the release branch.
  * The files are named `ratify_<major>-<minor>-<patch>_<OS>_<ARCH>` with `.zip` files for Windows and `.tar.gz` for all others.

### Weekly Dev Release

Ratify is configured to generate and publish dev build images based on schedule [here](https://github.com/deislabs/ratify/blob/main/.github/workflows/publish-package.yml#L8).  
1. The `ratify` image and `ratify-crds` image for dev builds exist as separate packages on Github [here](https://github.com/deislabs/ratify/pkgs/container/ratify-dev) and [here](https://github.com/deislabs/ratify/pkgs/container/ratify-crds-dev).
2. the `repository` `crdRepository` and `tag` fields must be updated in the helm chart to point to dev build instead of last released build. Please set the tag to be latest tag found at the corresponding `-dev` suffixed package. An example install command:
```
helm install ratify \
    ./charts/ratify --atomic \
    --namespace gatekeeper-system \
    --set image.repository=ghcr.io/deislabs/ratify-dev
    --set image.crdRepository=ghcr.io/deislabs/ratify-crds-dev
    --set image.tag=dev.20230307
    --set-file notaryCert=./test/testdata/notary.crt
```
NOTE: the tag field is the only value that will change when updating to newer dev build images