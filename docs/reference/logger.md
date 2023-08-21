# Logger

## Overview
Logging is a very important part of Ratify as it provides a way to debug and understand the behavior of the system. Ratify uses [Logrus](https://github.com/sirupsen/logrus) as the logging library.

## Usage
Ratify exposed a few options for users to configure the behavior of the logger. The logger can be configured through the helm chart values or the configuration file.

### Formatter
Users could set the formatter to be `text`, `json` or `logstash`. The default value is `text` if it's not set.

`json` and `text` are 2 commonly used formats for logging system. `logstash` is another format that's compatible with [Logstash](https://www.elastic.co/logstash).

### Request Headers
Ratify supports logging required headers from the external data request. Users could configure the logger to log the request headers by setting the `requestHeaders` field in the config. The `requestHeaders` field is a map of field name to header name in the request.

Fields that Ratify supports logging are:
- `traceIDHeaderName`: The trace ID header name in the request. Ratify will log the trace ID if it's set in the request header. Otherwise, it will generate a new trace ID and log it.

## Configuration
When running Ratify as a K8s add-on, users could configure the logger by setting helm chart values, e.g.
```yaml
logger:
  formatter: "text"
  requestHeaders:
    traceIDHeaderName: "x-trace-id"
```



