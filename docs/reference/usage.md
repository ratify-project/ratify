# Usage

This page documents useful flags and options supported by Ratify

## Environment variables

- `RATIFY_LOG_LEVEL`: configure the log level. Valid options are
  - `PANIC`
  - `FATAL`
  - `ERROR`
  - `WARNING`
  - `INFO` (default)
  - `DEBUG`
  - `TRACE`
- `RATIFY_CONFIG`: change the default Ratify configuration directory. Defaults to `~/.ratify`

## Feature flags

Ratify may roll out new features behind feature flags, which are activated by setting the corresponding environment variable `RATIFY_<FEATURE_NAME>=1`.
A value of `1` indicates the feature is active; any other value disables the flag.

- `RATIFY_DYNAMIC_PLUGINS`: (disabled) Enables Ratify to download plugins at runtime from an OCI registry by setting `source` on the plugin config
- `RATIFY_USE_REGO_POLICY`: (disabled) Enables Ratify to generate verification reports in the format designed for evaluation by OPA engine with Rego language. Users can either use the OPA engine embedded in Ratify or the one within Gatekeeper.
- `RATIFY_PASSTHROUGH_MODE`: (disabled) Enables Ratify to offload the policy decision-making to Gatekeeper. Ratify verify API will just return verification reports without the decision. `RATIFY_USE_REGO_POLICY` must be enabled to use this feature.