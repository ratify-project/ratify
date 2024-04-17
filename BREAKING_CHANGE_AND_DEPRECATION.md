# Breaking changes and deprecations

Breaking changes are defined as a change to any of the following that causes installation errors or 
unexpected runtime behavior after upgrading to the next stable minor version of Ratify(cli, plugin,):
- Verification result and schema
- Default configuration value
- User facing plugin interfaces

We do our best to maintain backwards compatibility, however breaking changes may be introduced between version for features that have not yet reached v1.0.0 or only available in RATIFY_EXPERIMENTAL flag.

### Features in Preview
|     Feature    |  GA Plan |
|:--------------:|-----:|
| CRD  |  |  
| Dynamic plugin |                          |         
|    HA            |                          |
 

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
 - Configuration values
 
### Announced deprecations
|     Feature    | Deprecation announcement | Removal |
|:--------------:|:------------------------:|:-------:|
| Config Policy  |                          |         |
| License Plugin |                          |         |
 
## Attribution

The specification release process was created using content and verbiage from the following specifications:

- [Dapr reference](https://docs.dapr.io/operations/support/breaking-changes-and-deprecations/)
