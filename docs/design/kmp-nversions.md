# nVersionCount support for Key Management Provider

Author: Josh Duffney (@duffney)

Tracked issues in scope:

- https://github.com/ratify-project/ratify/issues/1751

Proposal ref:

- https://github.com/ratify-project/ratify/blob/dev/docs/proposals/Automated-Certificate-and-Key-Updates.md

## Problem Statement

In version 1.3.0 and earlier, Ratify does not support the nVersionCount parameter for Key Management Provider (KMP) resources. This means that when a certificate or key is rotated, Ratify updates the cache with the new version and removes the previous one, which may not suit all use cases.

For instance, if a user needs to retain the last three versions of a certificate or key in the cache, Ratify cannot meet this requirement without manually adjusting the KMP resource for each new version.

By supporting nVersionCount, Ratify would allow users to specify how many versions of a certificate or key should be kept in the cache, eliminating the need for manual updates to the KMP resource.

## Proposed Solution

To address this challenge, this proposal suggests adding support for the `maxVersionCount` parameter to the KMP resource in Ratify. This parameter will allow users to specify the number of versions of a certificate or key that should be retained in the cache.

When a new version of a certificate or key is created, Ratify will check the `maxVersionCount` parameter to determine how many versions should be retained in the cache. If the number of versions exceeds the specified count, Ratify will remove the oldest version from the cache.

If a version is disabled, Ratify will remove it from the cache. This ensures that disabled versions are not retained in the cache, reducing the risk of using compromised keys or certificates being passed to the verifiers.

## Implementation Details

## Dev Work Items

## Open Questions

- If a version is disabled, should it be removed from the cache or retained based on the nVersionCount and marked as inactive\disabled?
  - Retaining the disabled version would require changing the KMP data structure to hold a list of versions and their status.
- If a version is disabled, does that count towards the nVersionCount? For example, if nVersionCount is set to 3 and one of the versions is disabled, should Ratify retain the last three active versions or the last three versions, regardless of their status?
- What does the KMP status look like when multiple versions are retained in the cache?
- Should the inlined provider support nVersionCount?

## Future Considerations
