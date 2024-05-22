# Breaking changes and deprecations

Breaking changes are defined as a change to any of the following that causes installation errors or 
unexpected runtime behavior after upgrading to the next stable minor version of Ratify:
- Verification API
- Verification result and schema
- Default configuration value
- User facing plugin interfaces

## Stability levels
- Generally available features should not be removed from that version or have its behavior significantly changed to avoid breaking existing users.
- Beta or pre-release versions may introduce breaking changes without deprecation notice.
- Alpha or experimental API versions may change in runtime behaviour without prior change and deprecation notice.

The following features are currently in experimental:
- [Dynamic plugin](https://ratify.dev/docs/reference/dynamic-plugins)
- [High Availability](https://ratify.dev/docs/quickstarts/ratify-high-availability)

## Process for applying breaking changes
- A deprecation notice must be posted as part of a release.
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

## Attribution

The specification release process was created using content and verbiage from the following specifications:
- [Kubernetes Deprecation Policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)
- [Dapr reference](https://docs.dapr.io/operations/support/breaking-changes-and-deprecations/)

