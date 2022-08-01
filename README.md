# Ratify

The project provides a framework to integrate scenarios that require
verification of reference artifacts and provides a set of interfaces
that can be consumed by various systems that can participate in
artifact ratification.

**WARNING:** This is experimental code. It is not considered production-grade
by its developers, nor is it "supported" software.

## Table of Contents
- [Community Meetings](#community-meetings)
- [Quick Start](#quick-start)
- [Documents](#documents)
- [Code of Conduct](#code-of-conduct)
- [Release Management](#release-management)
- [Licensing](#licensing)
- [Trademark](#trademark)

## Community meetings

- Agenda: https://hackmd.io/ABueHjizRz2iFQpWnQrnNA
- We hold a weekly Ratify community meeting with alternating times to accommodate more time zones.
Series #1 Tues 4-5pm
Series #2 Wed 1-2pm
Get Ratify Community Meeting Calendar [here](https://calendar.google.com/calendar/u/0?cid=OWJjdTF2M3ZiZGhubm1mNmJyMDhzc2swNTRAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
- We meet regularly to discuss and prioritize issues. The meeting may get cancelled due to holidays, all cancellation will be posted to meeting notes prior to the meeting.

## Quick Start

Try out ratify in Kuberenetes through Gatekeeper as the admission controller.

Prerequisite: Kubernetes v1.20 or higher

- Setup Gatekeeper with [external data](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata)

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=7
```

NOTE: `validatingWebhookTimeoutSeconds` increased from 3 to 7 so all Ratify operations complete in complex scenarios. See [discussion here](https://github.com/deislabs/ratify/issues/269) to remove this requirement. Kubernetes v1.20 or higher is REQUIRED to increase timeout.  

- Deploy ratify and a `demo` constraint on gatekeeper in the default namespace.

```bash
helm repo add ratify https://deislabs.github.io/ratify
helm install ratify \
    ratify/ratify --atomic

kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
```

Once the installation is completed, you can test the deployment of an image that is signed using Notary V2 solution.

- This will successfully create the pod `demo`

```bash=
kubectl run demo --image=ratify.azurecr.io/testimage:signed
kubectl get pods -w
```
You will see the "successful" created pod in a CrashLoopBackoff because the container by design executes a shell script displaying text and then crashes.

- Now deploy an unsigned image

```bash=
kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned
```

You will see a deny message from Gatekeeper denying the request to create it as the image doesn't have any signatures.

```bash=
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: ratify.azurecr.io/testimage:unsigned
```

You just validated the container images in your k8s cluster!

- Uninstall Ratify

```bash=
kubectl delete -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl delete -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
helm delete ratify
```

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

[microsoft-trademark]: https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks
