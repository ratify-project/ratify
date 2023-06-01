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

- `RATIFY_CERT_ROTATION`: (enabled) Enables Ratify to rotate TLS certificates automatically when they are about to expire. See [cert-controller](https://github.com/open-policy-agent/cert-controller) for more details on the implementation. The cert-controller checks the validation of certificates every 12 hours and generates a new certificate that is valid for 10 years 90 days before the old certificate expires.