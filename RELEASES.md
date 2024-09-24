# Ratify Release Process

This document describes the versioning scheme and release processes for Ratify.

## Attribution

The specification release process was created using content and verbiage from the following specifications:

* [ORAS Artifact Specification Releases](https://github.com/oras-project/artifacts-spec/blob/main/RELEASES.md)
* [ORAS Developer Guide](https://github.com/oras-project/oras-www/blob/main/docs/CLI/5_developer_guide.md)
* [Mystikos Release Management](https://github.com/deislabs/mystikos/blob/main/doc/releasing.md)
* [Gatekeeper Release Management](https://github.com/open-policy-agent/gatekeeper/blob/8f5201f0f48d50cc14153d100172689f03aa5f39/docs/Release_Management.md)

## Versioning

The Ratify project follows [semantic versioning](https://semver.org/) beginning with version `v0.1.0`.  Pre-release versions may be specified with a dash after the patch version and the following specifiers (in the order of release readiness):

* `alpha1`, `alpha2`, etc.
* `beta1`, `beta2`, etc.
* `rc1`, `rc2`, `rc3`, etc.

Example pre-release versions include `v0.1.0-alpha1`, `v0.1.0-beta2`, `v0.1.0-rc3`.  Pre-release versions are not required and stages can be bypassed (i.e. an `alpha` release does not require a `beta` release).  Pre-releases must be in order and gaps are not allowed (i.e. the only releases that can follow `rc1` are the full release or `rc2`).

## Pre Release Activity

1. Most e2e-scenarios for cli, K8s, and Azure are covered by the Ratify e2e tests. Please refer to this [document](test/validation.md) for the current supported and unsupported tests. Please perform manual prerelease validations for the unsupported tests list [here](test/validation.md#unsupported-tests)

2. If the format of the data returned for [external data calls](docs/reference/verification-result-version.md) has changed, validate change is also reflected in [`httpserver/types.go`](httpserver/types.go).

3. Delete all dev images generated since the previous release under the `ratify-dev` and `ratify-crds-dev` [packages](https://github.com/orgs/ratify-project/packages?repo_name=ratify). Each dev image tag is prefixed with `dev` followed by the date of creation and then the abbreviated 7 character commit SHA (e.g a build generated on March 8, 2023 from main branch with commit SHA `4cf98388ef33c587ef86b82e05cb0f7de2da2ea8` would be tagged `dev.20230308.4cf9838`). The most recent images are also tagged with a rolling tag `latest`.

4. Delete all dev helm charts since the previous release under the `ratify-chart-dev/ratify` [packages](https://github.com/orgs/ratify-project/packages?repo_name=ratify). Each helm chart is published with a semantic version compatible tag `0-dev` followed by the date of creation and then the abbreviated 7 character commit SHA (e.g a chart generated on March 8, 2023 from main branch with commit SHA `4cf98388ef33c587ef86b82e05cb0f7de2da2ea8` would be tagged `0-dev.20230308.4cf9838`). The most recent dev chart is also tagged with the rolling tag `0-dev`.

5. Copy contents from [`dev.helmfile.yaml`](dev.helmfile.yaml) to [`helmfile.yaml`](helmfile.yaml) & [`dev.high-availability.helmfile.yaml`](dev.high-availability.helmfile.yaml) to [`high-availability.helmfile.yaml`](high-availability.helmfile.yaml). You MUST update/remove values marked by comments in the files. The `dev` prefixed helmfiles are treated as staging files that are up to date with new changes on main branch. The primary `helmfile.yaml` and `high-availability.helmfile.yaml` MUST stay pinned to the current release since they are used by the quickstarts. Update `dev.helmfile.yaml` & `dev.high-availability.helmfile.yaml` ratify chart version to new release version.

## Git Release Flow

This section deals with the practical considerations of versioning in Git, this repo's version control system.  See the semantic versioning specification for the scope of changes allowed for each release type.

All releases will be of the form _vX.Y.Z_ where X is the major version, Y is the minor version and Z is the patch version.

### Patch releases

Applicable fixes, including security fixes, may be backported to supported releases, depending on severity and feasibility. Patch release are cut from branch `release-X.Y`. Commits can be cherry-picked from `main`, changes should be merged into latest supported minor release-X.Y branches once required PR requirements are met.

### Minor releases

When a minor release is required, the release commits should be merged with the `main` branch when ready.

* Alpha and Beta releases will be cut from the main branch.
* For RC and stable releases, a new branch `release-X.Y` will be created from `main`.
* Required changes for the minor release should be PRed to the `dev` branch, the change will then be cherry picked to `release-X.Y` from `main`.

### Major releases

When a major release is required, the release commits should be merged with the `main` branch when ready.  Major versions will usually require multiple pre-release versions. Similar to minor releases, the new branch should be created for the RC and stable release.

### Tag and Release

**X.Y.Z** refers to the version (git tag) of Ratify that is released.

1. Prepare the release with a [PR](https://github.com/ratify-project/ratify/pull/1801/files) to update the chart value.
2. When the `release-X.Y` branch is ready, a tag **X.Y.Z** should be pushed. e.g. `git tag v1.1.1` and `git push --tags`. This will trigger a [Goreleaser](https://goreleaser.com/) action that will build the binaries and creates a [GitHub release](https://help.github.com/articles/creating-releases/):

* The release will be marked as a draft to allow an final editing before publishing.
* The release notes and other fields can edited after the action completes.  The description can be in Markdown.
* The pre-release flag will be set for any release with a pre-release specifier.
* The pre-built binaries are built from commit at the head of the release branch.
  * The files are named `ratify_<major>-<minor>-<patch>_<OS>_<ARCH>` with `.zip` files for Windows and `.tar.gz` for all others.

## Supported Releases

Applicable fixes, including security fixes, may be cherry-picked into the release branch, depending on severity and feasibility. Patch releases are cut from that branch as needed.

We expect to "support" n (current). "Support" means we expect users to be running that version in production. For example, when v1.2 comes out, v1.1 will no longer be supported for patches, and we encourage users to upgrade to a supported version as soon as possible.

## Supported Kubernetes and Gatekeeper Versions

Ratify is assumed to be compatible with [GateKeeper Supported Versions](https://github.com/open-policy-agent/gatekeeper/blob/master/docs/Release_Management.md#supported-releases) and the [current Kubernetes Supported Versions](https://kubernetes.io/releases/patch-releases/#detailed-release-history-for-active-branches) per [Kubernetes Supported Versions policy](https://kubernetes.io/releases/version-skew-policy/).

For example, if Gatekeeper _supported_ versions are v3.13 and v3.14, and Kubernetes _supported_ versions are v1.28, v1.29, then current version of Ratify (v1.2) are assumed to be compatible with all supported Kubernetes versions (v1.28, v1.29) and Gatekeeper version(v3.13, v3.14).

## Post Release Activity

After a successful release, please prepare a [PR](https://github.com/ratify-project/ratify/pull/1805/files) to update the chart value in `dev` branch. After PR gets merged, manually trigger [quick start action](.github/quick-start.yml) to validate the quick start test is passing. Validate in the run logs that the version of ratify matches the latest released version.

### Weekly Dev Release

#### Publishing Guidelines

* Ratify is configured to generate and publish dev build images based on the schedule [here](https://github.com/ratify-project/ratify/blob/main/.github/workflows/publish-package.yml#L8).
* Contributors MUST select the `Helm Chart Change` option under the `Type of Change` section if there is ANY update to the helm chart that is required for proposed changes in PR.
* Maintainers MUST manually trigger the "Publish Package" workflow after merging any PR that indicates `Helm Chart Change`
  * Go to the `Actions` tab for the Ratify repository
  * Select `publish-ghcr` option from list of workflows on left pane
  * Select the `Run workflow` drop down on the right side above the list of action runs
  * Choose `Branch: main`
  * Select `Run workflow`
* Process to Request an off-schedule dev build be published
  * Submit a new feature request issue prefixed with `[Dev Build Request]`
  * In the the `What this PR does / why we need it` section, briefly explain why an off schedule build is needed
  * Once issue is created, post in the `#ratify` slack channel and tag the maintainers
  * Maintainers should acknowledge request by approving/denying request as a follow up comment

#### How to use a dev build

1. The `ratify` image and `ratify-crds` image for dev builds exist as separate packages on Github [here](https://github.com/ratify-project/ratify/pkgs/container/ratify-dev) and [here](https://github.com/ratify-project/ratify/pkgs/container/ratify-crds-dev).
2. the `repository` `crdRepository` and `tag` fields must be updated in the helm chart to point to dev build instead of last released build. Please set the tag to be latest tag found at the corresponding `-dev` suffixed package. An example install command scaffold:

```bash
helm install ratify \
    ./charts/ratify --atomic \
    --namespace gatekeeper-system \
    --set image.repository=ghcr.io/ratify-project/ratify-dev
    --set image.crdRepository=ghcr.io/ratify-project/ratify-crds-dev
    --set image.tag=dev.<YYYYMMDD>.<ABBREVIATED_GIT_HASH_COMMIT>
    --set-file notationCerts[0]=./test/testdata/notation.crt
```

NOTE: the tag field is the only value that will change when updating to newer dev build images
