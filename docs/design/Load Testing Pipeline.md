# Ratify Performance Load Testing for Azure
Author: Akash Singhal (@akashsinghal)

Ratify's security guarantees rely on having a clear understanding of the performance limitations. Previous iterations of performance testing have largely been adhoc and lack any real automation. This document outlines a performance pipeline and framework for Ratify performance testing. For now the pipeline, will target the Azure e2e scenario; however, it can be adapted for other cloud providers in the future. 

## Current State
The previous performance testing work can be found in the repository [here](https://github.com/anlandu/ratify-perf). This project largely contains loose K8s templates (pods, jobs, deployments) to mimic various loads. There is a small Go binary to generate k8s resouce templates based on the # of containers, # of pods. There is also a script to generate images and signatures based on type of resources, # of resources, registry name, # of signatures. This script pushes all the generated artifacts to a target registry. It is very rudimentary/manual and thus cannot be relied upon in a standardized framework.

## Goals
* Create an Azure Devops Pipeline for Ratify performance tests
* Create a `csscgen` tool to generate k8s resource templates + necessary images + referrers and push to registry. 
* Collect metrics from Ratify's prometheus endpoint and collate
* Collect diagnostic logs from Ratify's pods
* Run Ratify in various configurations to benchmark performance

## Measurements
* Ratify pod memory usage
* Ratify pod cpu usage
* \# of 429s from ACR
* \# of system failures from Ratify due to timeout
* Request duration for both verification and mutation

## Test Dimensions
* Ratify installation type (regular vs HA)
* K8s workload type (deployment vs Job)
    - \# of deployments / # of jobs
    - \# of replicas (pod count)
    - \# of images per resource
    - \# of signatures per image
* ACR
    - Region Size (This affects the request per second rate)
        - We will be testing XS, S, M, L regions
            - eastus2: L
            - central us (Availability Zone enabled): M
            - eastus2 (Availability Zone enabled): S
            - eastus2euap: XS
        - Our AKS cluster will be in eastus2. To minimize region latency from ACR to AKS, we have tried to pick nearby/same regions.

## Test Scenarios
- Singular Deployment (S)
    - 1 Deployment
    - 200 replicas
    - 10 containers (10 unique images)
    - 5 signatures per image
    - Nodes Required ~ 11
- Singular Deployment (M)
    - 1 Deployment
    - 2,000 replicas
    - 10 containers (10 unique images)
    - 5 signatures per image
    - Nodes Required ~ 62
- Singular Deployment (L)
    - 1 Deployment
    - 10,000 replicas
    - 10 containers (10 unique images)
    - 5 signatures per image
    - Nodes Required ~ 307
- Singular Deployment (XL)
    - 1 Deployment
    - 20,000 replicas
    - 10 containers (10 unique images)
    - 5 signatures per image
    - Nodes Required ~ 702
- Sharded Jobs: Same sizes as above but split parallel containers across 5 different jobs

## Azure Dev Ops Pipeline

The pipeline will be set up using an ADO Pipeline to leverage some existing gatekeeper templates for perf testing.
### Prerequisites (needs to be done once per subscription)
* Set correct environment variables that are configurable
    - Each configurable aspect of the pipeline will surface an environment variable that can be configured at runtime
* Set subscription
* Assign Identity and Roles
    * Upstream ADO runner identity assigned Owner access to target subscription
        * Owner required for role assignment creation for access to centralized AKV and ACR that are persistent between runs

    * Create resource group
        * Create premium ACRs for regions of different sizes
        * Create AKV
* Setup AKV & Push Artifacts to ACR
    * Setup Notation Azure signing following directions [here](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-tutorial-sign-build-push)
    * docker login to ACR
    * Use `csscgen` project's` genArtifacts.sh` script to push artifacts to registry (see below)
        * currently only support notation but can be modified to generate arbitrary types of artifacts in future

### Pipeline Execution
* Runner login using identity
* Update the az cli version and aks-preview cli version
* Generate load test resource group (this is where Workload identity and AKS cluster will live)
* Generate AKS Cluster, generate Ratify MI
    * AKS
        * options: (these are configured via env variables)
            * node pool size
            * workload identity enabled
            * region
            * kubernetes version
            * max pods per node
            * vm sku
        * OIDC & workload identity enable the AKS cluster
        * Attach ACR to AKS cluster
    * AKV
        * Grant get secret role to MI
    * MI
        * Create Ratify MI
        * Create federated credential based on Ratify MI and service account
        * Grant ACRPull role to MI
* Install Gatekeeper
    * options:
        * replicas
        * audit interval
        * emit events
        * enable external data
        * audit cache
        * GK version
    * Apply GK config for excluding namespaces
        * exclude: kube-\*, gatekeeper-\*, cluster-loader, monitoring
* Apply constraint template and constraint
* Generate resource templates:
    * Create cluster-loader's config template by substituting env variables in the tepmlate file
    * Clone `csscgen` repo and build the binary
        * use `csscgen genk8s` to create deployment/job template based on registry host name, number of replicas, number of images, number of signatures
* Install Ratify
    * workload identity enabled
    * AKV cert settings
    * replicas?
    * HA mode?
    * Ratify version
* Use cluster loader (see below for more info)
* Scrape metrics and download logs
    * port forward to localhost the prometheus endpoint for ratify
    * parse scraped metrics to look for
    * upload ratify + GK logs as artifacts
* Tear down cluster and delete resource group for test

## `csscgen` tool
https://github.com/akashsinghal/csscgen

This repository combines the initial work done in the first round of perf testing for ratify. It creates a new CLI tool called `csscgen` that can be used to generate k8s resource templates for deployments/jobs based on the # of containers, # of replicas, # of referrers. It adds list of containers and populates each with the registry reference that is generated when the same tool is used to generate the artifacts to push to registry.

This tool is not related to Ratify and thus could be used in the future for other use cases.

The `genArtifacts.sh` script takes in the registry host name, # of images, & # of referrers and generates unique container image and referrers attached to it. These are then pushed to the registry. The naming of the images follow same convention used by the `cssgen genk8s` cli tool to generate the accompanying templates. TODO: in the future we can incorporate script into the `csscgen` tool.

## Cluster Loader v2

https://github.com/kubernetes/perf-tests/tree/master/clusterloader2

This tool helps with generating and then loading arbitrary resources into a target cluster. It generates new namespaces and the specified copies of the target resource at different rates (QPS). It collects logs from the api server and watches for resource creation success/failure. Finally, a report is generated with resource consumption across the cluster during the cluster-loader process.

## ACR Testing

The load ACR can handle is dependent on the size of the registry's region. Size refers to the number of compute resources ACR service has allocated for that region. As a result, smaller regions MAY result in 429 http response status for requests from ratify due to the burst nature on initial resource deployment. 

We are working with ACR to determine if it is safe to perform large load tests on smaller regions. We need to guarantee we don't affect customer workloads in the region.


