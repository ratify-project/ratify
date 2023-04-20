# Instrumentation

This page outlines current instrumentation support in Ratify. It also contains guides for installing metrics providers and supported dashboards.

## Metrics Supported

Metrics Types:

- Counter: Value that accumulates over time. (e.g request count, signatures verified)
- Gauge: Point-in-time value of a continuous data stream (e.g file system size, speed, pressure)
- Histogram: Aggregation of counters where each bin is bounded from the min value (0) to the upper bin boundary. For example if we had bin boundaries [0, 1, 2, 3, 4, 5] and the measured value is 3.5, then the resulting histogram would be [0, 0, 0, 1, 1].

|             Name              | Type      |     Unit     | Attributes                                                                                                                                           |                                                                                                   Description                                                                                                    |
|:-----------------------------:| --------- |:------------:| ---------------------------------------------------------------------------------------------------------------------------------------------------- |:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|  ratify_verification_request  | Histogram | milliseconds | N/A                                                                                                                                                  | Duration of a single request to the `/verify` endpoint. Histogram bins:   `[0, 10, 30, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1400, 1600, 1800, 2000, 2300, 2600, 4000, 4400, 4900]` |
|    ratify_mutation_request    | Histogram | milliseconds | N/A                                                                                                                                                  |                    Duration of a single request to the `/mutate` endpoint. Histogram bins: `[0, 10, 30, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1400, 1600, 1800]`                    |
|   ratify_verifier_duration    | Histogram | milliseconds | `verifier`: name of the verifier <br/> `subject`: full subject reference <br/> `success`: verifier result <br/> `error`: if operation returned error |                             Duration of a single verifier's execution for a single referrer artifact. Histogram bins: `[0, 10, 50, 100, 200, 300, 400, 600, 800, 1100, 1500, 2000]`                              |
|   ratify_system_error_count   | Counter   |     N/A      | `error`: error message                                                                                                                               |                                                                                    Count of errors emitted   by http handlers                                                                                    |
| ratify_registry_request_count | Counter   |     N/A      | `status_code`: registry request status code <br/> `registry_host`: registry host name                                                                |                                                                                        Count of requests made to registry                                                                                        |
|    ratify_blob_cache_count    | Counter   |     N/A      | `hit`: boolean cache hit                                                                                                                             |                                                                                        Count of ORAS blob cache hit/miss                                                                                         |                                                                                                                                                                   |

### Azure Metrics

| Name                            | Type      | Unit         | Attributes                                                      | Description                                                     |
| ------------------------------- | --------- | ------------ | --------------------------------------------------------------- | --------------------------------------------------------------- |
| ratify_aad_exchange_duration    | Histogram | milliseconds | `resource_type`: resource scope (ACR vs AKV)                    | Duration of federated JWT exchange for AAD resource scope token |
| ratify_acr_exchange_duration    | Histogram | milliseconds | `repository`: full ACR repository name for token exchange scope | Duration of  exchange of AAD token for ACR refresh token        |
| ratify_akv_certificate_duration | Histogram | milliseconds | `certificate_name`: name of AKV certificate object              | Duration of AKV certificate fetch operation                     |

## Metrics Providers Supported

### Prometheus

Prometheus is a very popular metrics, monitoring, and database tool. It is time series based and is widely used in K8s. Prometheus metrics collection is pull-based. A prometheus instance is installed in the cluster. This instance is responsible for periodically "scraping" the metrics from the `/metrics` endpoint for each resource. Ratify's prometheus metrics provider implementation  creates a new http server which binds the `/metrics` endpoint to a prometheus handler. Once the OpenTelemetry Prometheus exporter is created, the handler takes care of publishing the correct Prometheus-formatted metrics. 


Scraping is achieved via annotations on the resources that expose metrics on the `/metrics` endpoint. There are two parts to this:
1. How does prometheus know which resources to consider and what annotations to look for?
The prometheus config is responsible for this. The configuration contains instructions on which annotations to look for to determine if a resource is eligible for scraping and which resources it should check. Depending on the prometheus configuration that exists on the cluster, users may need to append an additional scrape configuration (instruction below on how to do this). 
For ratify, the metrics will be published by the Ratify pod. 
2. How do we tell Prometheus to look for Ratify metrics specifically from the `/metrics` endpoint on the Ratify pod? The Ratify helm chart contains annotations to the on Deployment spec to indicate scraping and which port the metrics server is running on the pod:
    ```
    annotations:
        prometheus.io/scrape: 'true'
        prometheus.io/port: '8888'
    ```

### Prometheus and Grafana Setup

This quick start provides instructions on installing and using Prometheus with Grafana to visualize the current supported metrics. This guide assumes neither is installed on the K8s cluster. (Warning: This guide is purely informational and should not be considered a production-grade solution for installing Prometheus and Grafana on cluster. Please consult the corresponding project guidelines for more scenario-specific information)

Prior to installing Ratify on the cluster:

1. Create a new namespace, if not yet created, for all instrumentation
    ```
    kubectl create namespace monitoring
    ```
1. Apply an additional scrap configuration to Prometheus via a secret:
    ```
    kubectl apply -f instrumentation/additional-scrape-configs.yaml -n monitoring
    ```
    - Note: if the [scrape config](../../instrumentation/prometheus-additional.yaml) stored is updated and the secret needs to be regenerated. Run this command prior:
        ```
        kubectl create secret generic additional-scrape-configs --from-file=instrumentation/prometheus-additional.yaml --dry-run=client -oyaml > instrumentation/additional-scrape-configs.yaml
        ```
1. Add the prometheus-community helm repository
    ```
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    ```
1. Install the `prometheus-community/kube-prometheus-stack` helm chart. This will install the standard kubernetes metrics including Grafana. It will also expose Prometheus and Grafana on an external IP and load a Ratify specific dashboard.
    ```
    helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --atomic\
        --set prometheus.prometheusSpec.additionalScrapeConfigsSecret.enabled=true \
        --set prometheus.prometheusSpec.additionalScrapeConfigsSecret.name=additional-scrape-configs \
        --set prometheus.prometheusSpec.additionalScrapeConfigsSecret.key=prometheus-additional.yaml \
        --set prometheus.service.type=LoadBalancer \
        --set grafana.service.type=LoadBalancer \
        --set grafana.sidecar.dashboards.enabled=true
    ```
1. Apply ConfigMap with Ratify Dashboard to monitoring namespace
    ```
    kubectl apply -f instrumentation/grafana_configMap.yaml -n monitoring
    ```
1. Find the Grafana service external IP
    ```
    kubectl get svc prometheus-grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
   ```
1. Navigate to Grafana instance by using external IP and port 80 in browser
1. Login using default credentials
    - username: admin
    - password: prom-operator
1. View Ratify dashboard under "General/Ratify"
    - Note: Metrics will not appear until after Ratify has been installed and an external data request is processed
    
### Adding a new metric provider

>Note: There is only support for a single metric exporter at a time

Additional metrics exporters can be added as metrics backends.
The exporter must be initialized in the `InitMetricsExporter` method as a new exporter type. Each exporter type is determined from the `metricsBackend` parameter. During metric exporter initialization a new metric reader must be instantiated and assigned to the static global variable `MetricsReader`. Once registered, the global `MetricsReader` is used to initialize OpenTelemetry meter and instruments.
