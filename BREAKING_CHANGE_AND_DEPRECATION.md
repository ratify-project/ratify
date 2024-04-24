# Breaking changes and deprecations

Breaking changes are defined as a change to any of the following that causes installation errors or 
unexpected runtime behavior after upgrading to the next stable minor version of Ratify:
- Verification result and schema
- Default configuration value
- User facing plugin interfaces

## Stability levels
- Generally available (GA) API and plugins may introduce breaking changes 2 releases after deprecation notice.
- Beta or pre-release API versions may introduce breaking changes 1 release after deprecation notice.
- Alpha or experimental API versions may change in runtime behavoiur without prior change and deprecation notice.

The following features are currently in experimental:
- Dynamic plugin
- HA

## Process for applying breaking changes

- A deprecation notice must be posted as part of a release.
- The breaking changes are applied two (2) releases after the release in which the deprecation was announced.
For example, feature X is announced to be deprecated in the 1.0.0 release notes and will then be removed in 1.2.0
- The PR containing the breaking changes should contain a "!" to indicate this is breaking change. E.g. feat! , fix!
- Breaking changes should be listed before new features in the release notes
- Create a issue to help customer to mitigate the change
 
## Deprecations
Deprecations can apply to:
 - CRDs
 - Experimental Features
 - Plugins
 - User facing Interfaces
 - Configuration Values
 - Cli Usage and Configuration
 - Verification Schema
 
## Process for deprecation

 - A deprecation notice must be posted as part of the release notes.
 - Documentation should mark the feature as deprecated and redirect user to the alternative implementation.

### Announced deprecations
|     Feature    | Deprecation announcement | Planned Removal |
|:--------------:|:------------------------:|:-------:|
| Config Policy  |                          |   v1.5      |
| License Plugin |                          |   v2      |
 
## Attribution

The specification release process was created using content and verbiage from the following specifications:
- [Kubernetes Deprecation Policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)
- [Dapr reference](https://docs.dapr.io/operations/support/breaking-changes-and-deprecations/)

