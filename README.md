<div align="center">
<img src="logo.svg" width="200">
</div>

# Ratify

Is a verification engine as a binary executable and on Kubernetes which enables verification of artifact security metadata and admits for deployment only those that comply with policies you create.

**WARNING:** This is not considered production-grade code
by its developers, nor is it "supported" software.

[![Go Report Card](https://goreportcard.com/badge/github.com/deislabs/ratify)](https://goreportcard.com/report/github.com/deislabs/ratify)
[![build-pr](https://github.com/deislabs/ratify/actions/workflows/build-pr.yml/badge.svg)](https://github.com/deislabs/ratify/actions/workflows/build-pr.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/deislabs/ratify/badge)](https://api.securityscorecards.dev/projects/github.com/deislabs/ratify)

## Table of Contents

- [Community Meetings](#community-meetings)
- [Quick Start](#quick-start)
- [Documents](#documents)
- [Code of Conduct](#code-of-conduct)
- [Release Management](#release-management)
- [Licensing](#licensing)
- [Trademark](#trademark)

## Community meetings

- Agenda: <https://hackmd.io/ABueHjizRz2iFQpWnQrnNA>
- We hold a weekly Ratify community meeting on Weds 4:30-5:30pm (Pacific Time)   
Get Ratify Community Meeting Calendar [here](https://calendar.google.com/calendar/u/0?cid=OWJjdTF2M3ZiZGhubm1mNmJyMDhzc2swNTRAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
- We meet regularly to discuss and prioritize issues. The meeting may get cancelled due to holidays, all cancellation will be posted to meeting notes prior to the meeting.
- Reach out on Slack at [cloud-native.slack.com#ratify](https://cloud-native.slack.com/archives/C03T3PEKVA9). If you're not already a member of cloud-native slack channel, first add [yourself here](https://communityinviter.com/apps/cloud-native/cncf).

## Pull Request Review Series
- We hold a weekly Ratify Pull Request Review Series on Mondays 5-6 pm PST.  
- People are able to use this time to walk through any Pull Requests and seek feedback from others in the Community.  If there are no PR to review, the meeting will be cancelled during that week.
- Reach out on Slack if you want to reserve a session for review or during our weekly community meetings.

## Quick Start

Try out ratify in Kubernetes through Gatekeeper as the admission controller.

Prerequisites:
- Kubernetes v1.20 or higher
- OPA Gatekeeper v3.10 or higher  

### Step 1: Setup Gatekeeper with [external data](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata)

> NOTE: If you have added Helm repository for Gatekeeper and Ratify, you can update them by executing `helm repo update` before installation.

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=5 \
    --set mutatingWebhookTimeoutSeconds=2
```

> NOTE: `validatingWebhookTimeoutSeconds` and `mutationWebhookTimeoutSeconds` increased from 3 to 5 and 1 to 2 respectively, so all Ratify operations complete in complex scenarios. See [discussion here](https://github.com/deislabs/ratify/issues/269) to remove this requirement. Kubernetes v1.20 or higher is REQUIRED to increase timeout. Timeout is configurable in helm chart under `provider.timeout` section.   

### Step 2: Deploy ratify on gatekeeper in the default namespace.

- Option 1: Install the last released version of Ratify

Note: if the crt/key/cabundle are NOT set under `provider.tls` in values.yaml, helm would generate a CA certificate and server key/certificate for you.

```bash
helm repo add ratify https://deislabs.github.io/ratify
# download the notary verification certificate
curl -sSLO https://raw.githubusercontent.com/deislabs/ratify/main/test/testdata/notary.crt
helm install ratify \
    ratify/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notaryCert=./notary.crt \
    --set featureFlags.RATIFY_CERT_ROTATION=true
```

- Option 2: Install ratify with charts from your local branch.  
Note: Latest chart in main may not be compatible with the last released version of ratify image, learn more about weekly dev builds [here](RELEASES.md/#weekly-dev-release) 
```bash
git clone https://github.com/deislabs/ratify.git
cd ratify
helm install ratify \
    ./charts/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notaryCert=./test/testdata/notary.crt \
    --set featureFlags.RATIFY_CERT_ROTATION=true
```

### Step 3: See Ratify in action

- Deploy a `demo` constraint.
```
kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
```

Once the installation is completed, you can test the deployment of an image that is signed using Notary V2 solution.

- This will successfully create the pod `demo`

```bash
kubectl run demo --image=wabbitnetworks.azurecr.io/test/notary-image:signed
kubectl get pods demo
```

Optionally you can see the output of the pod logs via: `kubectl logs demo`

- Now deploy an unsigned image

```bash
kubectl run demo1 --image=wabbitnetworks.azurecr.io/test/notary-image:unsigned
```

You will see a deny message from Gatekeeper denying the request to create it as the image doesn't have any signatures.

```bash
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: wabbitnetworks.azurecr.io/test/net-monitor:unsigned
```

You just validated the container images in your k8s cluster!

### Step 4: Uninstall Ratify
Notes: Helm does NOT support upgrading CRDs, so uninstalling Ratify will require you to delete the CRDs manually. Otherwise, you might fail to install CRDs of newer versions when installing Ratify.
```bash
kubectl delete -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl delete -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
helm delete ratify --namespace gatekeeper-system
kubectl delete crd stores.config.ratify.deislabs.io verifiers.config.ratify.deislabs.io certificatestores.config.ratify.deislabs.io
```

### Notes

If the image reference provided resolves to an OCI Index or a Docker Manifest List, validation will occur ONLY at the index or manifest list level. Ratify currently does NOT support image validation based on automatic platform selection. For more information, [see this issue](https://github.com/deislabs/ratify/issues/101).

## Documents

The [docs](docs/README.md) folder contains the beginnings of a formal
specification for the Reference Artifact Verification framework and its plugin model.

Meeting notes for weekly project syncs can be found [here](https://hackmd.io/ABueHjizRz2iFQpWnQrnNA?both)

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/).

For more information see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

## Release Management

The Ratify release process is defined in [RELEASES.md](./RELEASES.md).

## Licensing

This project is released under the [Apache-2.0 License](./LICENSE).

## Trademark

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines][microsoft-trademark]. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party's policies.

[microsoft-trademark]: https://www.microsoft.com/legal/intellectualproperty/trademarks
