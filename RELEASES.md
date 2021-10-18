# Hora Release Process

This document describes the versioning scheme and release processes for Hora.

## Attribution

The specification release process was created using content and verbiage from the following specifications:

   * [ORAS Artifact Specification Releases](https://github.com/oras-project/artifacts-spec/blob/main/RELEASES.md)
   * [ORAS Developer Guide](https://github.com/oras-project/oras-www/blob/main/docs/CLI/5_developer_guide.md)
   * [Mystikos Release Management](https://github.com/deislabs/mystikos/blob/main/doc/releasing.md)


## Versioning

The Hora project follows [semantic versioning](https://semver.org/) beginning with version `v0.1.0`.  Pre-release versions may be specified with a dash after the patch version and the following specifiers (in the order of release readiness):
* `alpha`
* `beta`
* `rc1`, `rc2`, `rc3`, etc.

Example pre-release versions include `v0.1.0-alpha`, `v0.1.0-beta`, `v0.1.0-rc1`.  Pre-release versions are not required and stages can be bypassed (i.e. an `alpha` release does not require a `beta` release).  `rc` releases must be in order and gaps are not allowed (i.e. the only releases that can follow `rc1` are the full release or `rc2`).

## Git Release Flow

This section deals with the practical considerations of versioning in Git, this repo's version control system.  See the semantic versioning specification for the scope of changes allowed for each release type.

### Patch releases

When a patch release is required, the patch commits should be merged with the `main` branch when ready.  Then a new branch should be created with the patch version incremented and optional pre-release specifiers.  For example if the previous release was `v0.1.0`, the branch could be named `v0.1.1` or `v0.1.1-rc1`.  The limited nature of fixes in a patch release should mean pre-releases can often be omitted.

### Minor releases

When a minor release is required, the release commits should be merged with the `main` branch when ready.  Then a new branch should be created with the minor version incremented and optional pre-release specifiers.  For example if the previous release was `v0.1.1`, the branch could be named `v0.2.0` or `v0.2.0-rc1`.  Pre-releases will be more common will be more common with minor releases.

### Major releases

When a major release is required, the release commits should be merged with the `main` branch when ready.  Then a new branch should be created with the major version incremented and optional pre-release specifiers.  For example if the previous release was `v1.1.1`, the branch could be named `v2.0.0` or `v2.0.0-alpha`.  Major versions will usually require multiple pre-release versions.

### Tag and Release

When the release branch is ready, at tag should be pushed with a name matching the branch name, e.g. `git tag v0.1.0-alpha` and `git push origin v0.1.0-alpha`.  This allows the creation of a release in GitHub (these steps are currently manual but are expected to be automated.)  See [GitHub Release](https://help.github.com/articles/creating-releases/).
* All releases before `v1.0.0` are pre-release.  After `v1.0.0` only versions with pre-release versions are pre-release.
* Find the correct commit ids and use the following to generate release notes:  `git log --no-merges --pretty=format:'- %s %H (%aN)' previous_release_commit..release_commit`
* Add release notes as the description field.  This field supports Markdown.
* Add pre-built binaries as tarball built from commit hash at the head of the release branch.
    * The file should be named `hora-<major>-<minor>-<patch>-<ARCH>.tar.gz`