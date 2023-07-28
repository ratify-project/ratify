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

- [Quick Start](#quick-start)
- [Community Meetings](#community-meetings)
- [Pull Request Review Series](#pull-request-review-series)
- [Documents](#documents)
- [Code of Conduct](#code-of-conduct)
- [Release Management](#release-management)
- [Licensing](#licensing)
- [Trademark](#trademark)

## Quick Start

Try out ratify in Kubernetes through Gatekeeper as the admission controller.

Prerequisites:
- Kubernetes v1.20 or higher
- OPA Gatekeeper v3.10 or higher
- [helmfile](https://helmfile.readthedocs.io/en/latest/#installation) v0.14 or higher (See [ratify-quickstart-manual.md](docs/quickstarts/ratify-quickstart-manual.md) for manual install steps)

### Step 1: Install Gatekeeper, Ratify, and Constraints

```bash
curl -L https://raw.githubusercontent.com/deislabs/ratify/main/helmfile.yaml | helmfile sync -f - 
```

### Step 2: See Ratify in action

Once the installation is completed, you can test the deployment of an image that is signed using Notary V2 solution.

- This will successfully create the pod `demo`

```bash
kubectl run demo --image=ghcr.io/deislabs/ratify/notary-image:signed
kubectl get pods demo
```

Optionally you can see the output of the pod logs via: `kubectl logs demo`

- Now deploy an unsigned image

```bash
kubectl run demo1 --image=ghcr.io/deislabs/ratify/notary-image:unsigned
```

You will see a deny message from Gatekeeper denying the request to create it as the image doesn't have any signatures.

```bash
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: wabbitnetworks.azurecr.io/test/net-monitor:unsigned
```

You just validated the container images in your k8s cluster!

### Step 4: Uninstall
```bash
curl -L https://raw.githubusercontent.com/deislabs/ratify/main/helmfile.yaml | helmfile destroy --skip-charts -f - 
```

### Notes

If the image reference provided resolves to an OCI Index or a Docker Manifest List, validation will occur ONLY at the index or manifest list level. Ratify currently does NOT support image validation based on automatic platform selection. For more information, [see this issue](https://github.com/deislabs/ratify/issues/101).

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
