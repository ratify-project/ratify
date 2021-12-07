# Ratify

The project provides a framework to integrate scenarios that require
verification of reference artifacts and provides a set of interfaces
that can be consumed by various systems that can participate in
artifact ratification.

**WARNING:** This is experimental code. It is not considered production-grade
by its developers, nor is it "supported" software.

## Table of Contents

- [Quick Start](#quick-start)
- [Documents](#documents)
- [Code of Conduct](#code-of-conduct)
- [Release Management](#release-management)
- [Licensing](#licensing)
- [Trademark](#trademark)

## Quick Start

Try out ratify in Kuberenetes through Gatekeeper as the admission controller.

- Setup Gatekeeper with [external data](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata)

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set controllerManager.dnsPolicy=ClusterFirst,audit.dnsPolicy=ClusterFirst
```

- Deploy ratify and a `demo` constraint on gatekeeper

```bash
git clone https://github.com/deislabs/ratify.git
cd ratify

helm install ratify \
    ./charts/ratify --atomic

kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml
```

Once the installation is completed, you can test the deployment of an image that is signed using Notary V2 solution.

- Create the namespace `demo`

```bash=
kubectl create ns demo
```

- This will successfully create the pod `demo` in the namespace `demo`

```bash=
kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
```

- Now deploy an unsigned image

```bash=
kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned -n demo
```

You will see a deny message from Gatekeeper as the image doesn't have any signatures.

You just validated the container images in your k8s cluster!

- Uninstall Ratify

```bash=
kubectl delete -f ./charts/ratify-gatekeeper/templates/constraint.yaml
helm delete ratify
kubectl delete namespace demo
```

## Documents

The [docs](docs/README.md) folder contains the beginnings of a formal
specification for the Reference Artifact Verification framework and its plugin model.

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
