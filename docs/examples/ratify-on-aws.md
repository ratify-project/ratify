# Ratify on AWS
This guide will explain how to get up and running with Ratify on AWS using EKS and ECR. This will involve setting up
necessary AWS resources, installing necessary components, and configuring them properly. Once everything is set up we
will walk through a simple scenario of verifying the signature on a container image at deployment time.

By the end of this guide you will have a public ECR repository, an EKS cluster with Gatekeeper and Ratify installed,
and have validated that only images signed with a particular key can be deployed.

This guide assumes you are starting from scratch, but portions of the guide can be skipped if you have an existing EKS
cluster or ECR repository.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Setting Up ECR](#set-up-ecr)
3. [Setting Up EKS](#set-up-eks)
4. [Prepare Container Image](#prepare-container-image)
5. [Configure Ratify](#configure-ratify)
6. [Deploy Ratify](#deploy-ratify)
7. [Deploy Container Image](#deploy-container-image)
8. [Cleaning Up](#cleaning-up)

## Prerequisites
There are a couple tools you will need locally to complete this guide:
- [awscli](https://aws.amazon.com/cli/): This is used to interact with AWS and provision necessary resources
- [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html): This is used to easily provision EKS clusters
- [kubectl](https://kubernetes.io/docs/tasks/tools/): This is used to interact with the EKS cluster we will create
- [helm](https://helm.sh/docs/intro/quickstart/): This is used to install ratify components into the EKS cluster
- [docker](https://www.docker.com/get-started): This is used to build the container image we will deploy in this guide
- [cosign](https://github.com/sigstore/cosign): This is used to sign the container image we will deploy in this guide
- [ratify](https://github.com/deislabs/ratify/releases): This is used to check images from ECR locally
- [jq](https://stedolan.github.io/jq/): This is used to capture variables from json returned by commands

If you have not done so already, configure awscli to interact with your AWS account by following these [instructions](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-prereqs.html).

## Set Up ECR
We need to provision a public container repository to make our container images and their associated artifacts
available to our EKS cluster. We will do this using awscli. For this guide we will be provisioning a public ECR
repository to keep things simple.

```shell
export REPO_NAME=ratifydemo
export REPO_URI=$(aws ecr-public create-repository --repository-name $REPO_NAME --region us-east-1 | jq -r ."repository"."repositoryUri" )
```

We will use the repository URI returned by the create command later to build and tag the images we create.

For more information on provisioning ECR repositories check the [documentation](https://docs.aws.amazon.com/AmazonECR/latest/public/public-getting-started.html).

## Set Up EKS
We will need to provision a Kubernetes cluster to deploy everything on. We will do this using the `eksctl` command line
utility. Before provisioning our EKS cluster we will need to create a key pair for the nodes:

```shell
aws ec2 create-key-pair --region us-east-1 --key-name ratifyDemo
```

Save the output to your local machine, then run the following to create the cluster:

```shell
eksctl create cluster \
  --name ratify-demo \
  --region us-east-1 \
  --zones us-east-1c,us-east-1d \
  --with-oidc \
  --ssh-access \
  --ssh-public-key ratifyDemo
```

The template will provision a basic EKS cluster with default settings.

Additional information on EKS deployment can be found in the EKS [documentation](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html).

## Prepare Container Image
For this guide we will create a basic container image we can use to simulate deployments of a service. We will start by
building the container image:

```shell
docker build -t $REPO_URI:v1 https://github.com/wabbit-networks/net-monitor.git#main
```

After the container is built we need to push it to the repository:

```shell
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin $REPO_URI

docker push $REPO_URI:v1
```

Once the container is built and pushed, we will use cosign to create a key and sign the container image:

```shell
cosign generate-key-pair

cosign sign --key cosign.key $REPO_URI:v1
```

Both the container image and the signature should now be in the public ECR repository. We can use cosign to verify the
signature and image are present and valid:

```shell
docker rmi $REPO_URI:v1

cosign verify --key cosign.pub $REPO_URI:v1
```

## Configure Ratify
### Ratify
We need to ensure that Ratify is properly configured to find signature artifacts for our container image. This is done
using a json configuration file. The Ratify configuration file for the guide is created and deployed by the helm chart,
but we can look at what will be generated below:

```json
{
    "store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "cosignEnabled": true,
                "localCachePath": "./local_oras_cache"
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.dev.cosign.simplesigning.v1+json": "any"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [        
          {
            "name": "cosign",
            "artifactTypes": "application/vnd.dev.cosign.simplesigning.v1+json",
            "key": "/usr/local/ratify-certs/cosign.pub"
          }
        ]        
    }
}
```

This configuration file does the following:
- Enables the built-in `oras` referrer store with cosign support which will retrieve the necessary manifests and
  signature artifacts from the container registry
- Enables the `cosign` verifier that will validate cosign signatures on container images

The configuration file and cosign public key will be mounted into the Ratify container via the helm chart.

### Gatekeeper
The Ratify container will perform the actual validation of images and their artifacts, but
[Gatekeeper](https://github.com/open-policy-agent/gatekeeper) is used as the policy controller for Kubernetes. The helm
chart for this guide has a basic Gatekeeper rego that checks for the string "false" in the results from the Ratify
container.

This rego is kept simple to demonstrate the capability of Ratify. More complex combinations of regos and Ratify
verifiers can be used to accomplish many types of checks. See the [Gatekeeper docs](https://open-policy-agent.github.io/gatekeeper/website/docs/)
for more information on rego authoring.

## Deploy Ratify
We first need to install Gatekeeper into the cluster. We will use the Gatekeeper helm chart with some customizations:

```shell
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set controllerManager.dnsPolicy=ClusterFirst,audit.dnsPolicy=ClusterFirst
```

Once Gatekeeper has been deployed into the cluster we can deploy Ratify using the provided helm chart:

```shell
export COSIGN_PUBLIC_KEY=$(cat cosign.pub)

helm install ratify charts/ratify \
    --set cosign.enabled=true \
    --set cosign.key=$COSIGN_PUBLIC_KEY

kubectl apply -f ./library/default/template.yaml
kubectl apply -f ./library/default/samples/constraint.yaml
```

We can then confirm all pods are running:

```shell
kubectl get po -A
```

We should see a ratify pod and some gatekeeper pods running

## Deploy Container Image
Now that the signed container image is in the registry and Ratify is installed into the EKS cluster we can deploy our
container image:

```shell
kubectl create ns demo

kubectl run demosigned -n demo --image $REPO_URI:v1
```

We should be able to see from the Ratify and Gatekeeper logs that the container signature was validated. The pod for
the container should also be running.

```shell
kubectl logs deployment/ratify
```

We can also test that an image without a valid signature is not able to run:

```shell
kubectl run demounsigned -n demo --image hello-world
```

The command should fail with an error and we should be able to see from the Ratify and Gatekeeper logs that the
signature validation failed.

```shell
kubectl logs deployment/ratify
```

## Cleaning Up
We can use awscli and eksctl to delete our ECR repository and EKS cluster:

```shell
aws ecr-public delete-repository --region us-east-1 --repository-name $REPO_NAME

eksctl delete cluster --region us-east-1 --name ratify-demo
```
