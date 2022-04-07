# Policy Providers

Ratify implements an extensible policy provider interface allowing for different policy providers to be created and registered. The policy provider to be used is determined by the policy plugin specified in the `policies` section of the configuration. 

Currently, Ratify supports a Configuration based Policy Provider named `configPolicy`.

## Sample Config

```
...
"policies": {
    "version": "1.0.0",
    "policyPlugin": {
        "name": "configPolicy",
        "artifactVerificationPolicies": {
            "application/vnd.cncf.notary.v2.signature": "all"
        }
    }
},
...
```

- The `name` field is REQUIRED and MUST match the name of the registered policy provider
- `artifactVerificationPolicies`: map of artifact type to policy; each entry in the map's policy must be satisfied for Ratify to return true.
    - `any`: policy that REQUIRES at least one artifact of specified type to verify to `true` 
    - `all`: policy that REQUIRES all artifacts of specified type to verify to `true`

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).