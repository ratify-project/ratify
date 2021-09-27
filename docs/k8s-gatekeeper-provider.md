## Hora with OPA Gatekeeper
This document outlines steps to run Hora as an external data provider to OPA Gatekeeper of Kubernetes. This assumes that OPA Gatekeeper has been deployed to Kubernetes with external data provider feature enabled.

### Setup
- Setup local Kubernetes cluster. You can use Kubernetes that comes with [Docker Desktop](https://docs.docker.com/desktop/kubernetes/). This will enable access of the local images that are built without the need of pushing them to a remote registry. 
- Update the ```./config/config.json``` to the certs folder something like ```/usr/local/wa-certs/wabbit-networks.crt```. This will be used to configure the volume mount when Hora is deployed as a k8s service.
- Build the container image for Hora
```bash
docker build -f httpserver/Dockerfile -t deislabs/hora:v1 .
```
- Update the secret value in the  [deployment](./deploy/hora.yaml) with base64 encode of the notary cert that will be used for signature verification.
- Deploy hora as service in k8s
```bash
kubectl apply -f deploy/hora.yaml
```
- Deploy hora as external data provider for gatekeeper

```bash
kubectl apply -f deploy/policy/provider.yaml
```

- Deploy gatekeeper template that invokes Hora as part of admission controller
```bash
kubectl apply -f deploy/policy/template.yaml
```

- Deploy gatekeeper constraint for the policy to be applied for all deployment & pod objects
```bash
kubectl apply -f deploy/policy/constraint.yaml
```

- Now OPA Gatekeeper has been configured to invoke Hora for image verification. 
- Follow the nv2 demo script to create signed and unsigned images using the same cert that is configured with Hora
- The deployment of signed image should be success. The deployment of unsigned image should be blocked by admission controller with error 
```
Error from server ([signed-image] Image registry.wabbit-networks.io:5000/net-monitor:unsigned@sha256:2179f005cd7701a236d3f574b6c5e92f1783b23552df73995d521a50632ced80 verification failed false): error when creating "unsigned.yaml": admission webhook "validation.gatekeeper.sh" denied the request: [signed-image] Image registry.wabbit-networks.io:5000/net-monitor:unsigned@sha256:2179f005cd7701a236d3f574b6c5e92f1783b23552df73995d521a50632ced80 verification failed false
```
