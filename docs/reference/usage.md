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

## Optional features

Ratify may roll out new features behind feature flags, which are activated by setting the corresponding environment variable `RATIFY_<FEATURE_NAME>=1`.
A value of `1` indicates the feature is active; any other value disables the flag.
