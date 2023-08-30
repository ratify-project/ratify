# Creating Plugins

Ratify supports plugins for [stores](../reference/store.md) and [verifiers](../reference/verifier.md) so that users can add capabilities that aren't included in the core application. This document is a guide on creating your own plugins.

## Overview

At a high level, Ratify will execute plugins as child processes, passing environment variables and JSON over STDIN. A plugin will perform its work and return JSON over STDOUT. There is no restriction for what programming language must be used for a plugin executable. For convenience, plugin skeletons have been created for programs written in Go which handle inputs and provide simple functions for authors to implement.

## Verifier

The easiest way to get started is to use [deislabs/ratify-verifier-plugin](https://github.com/deislabs/ratify-verifier-plugin), which is a template repository that stubs out a working verifier plugin.

## Store

There is a sample store plugin located at [plugins/referrerstore/sample/sample.go](https://github.com/deislabs/ratify/blob/main/plugins/referrerstore/sample/sample.go)

To build the plugin and place it into your plugins dir:

```shell
cd plugins/referrerstore/sample/
CGO_ENABLED=0 go build -o ~/.ratify/plugins/sample .
```

To use the plugin, add the following in your `config.json`:

```json
"store": {
  "version": "1.0.0",
  "plugins": [
    {
      "name": "oras"
    },
    {
      "name": "sample"
    }
  ]
},
```
