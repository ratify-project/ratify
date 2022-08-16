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
 [Test.bats](test/bats/test.bats) provides limited end to end test coverage, while we are working on improving our coverage, please perform [manual validations](test/ManualValidation.md) to ensure release quality.

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
