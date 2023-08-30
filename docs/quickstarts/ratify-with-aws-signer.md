# Ratify with AWS Signer

This guide will explain how to get started with Ratify on AWS using EKS, ECR, and AWS Signer. This will involve setting up necessary AWS resources, installing necessary components, and configuring them properly. Once everything is set up we will walk through a simple scenario of verifying the signature on a container image at deployment time.

By the end of this guide you will have a public ECR repository, an EKS cluster with Gatekeeper and Ratify installed, and have validated that only images signed by a trusted AWS Signer SigningProfile can be deployed.

This guide assumes you are starting from scratch, but portions of the guide can be skipped if you have an existing EKS cluster, ECR repository, or AWS Signer resources.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Set up ECR](#set-up-ecr)
3. [Set up EKS](#set-up-eks)
4. [Prepare Container Image](#prepare-container-image)
5. [Sign Container Image](#sign-container-image)
6. [Deploy Gatekeeper](#deploy-gatekeeper)
7. [Configure IAM Permissions](#configure-iam-permissions)
8. [Deploy Ratify](#deploy-ratify)
9. [Deploy Container Image](#deploy-container-image)
10. [Cleaning Up](#cleaning-up)

## Prerequisites

There are a couple tools you will need locally to complete this guide:

- [awscli](https://aws.amazon.com/cli/): This is used to interact with AWS and provision necessary resources
- [eksctl](https://docs.aws.amazon.com/eks/latest/userguide/eksctl.html): This is used to easily provision EKS clusters
- [kubectl](https://kubernetes.io/docs/tasks/tools/): This is used to interact with the EKS cluster we will create
- [helm](https://helm.sh/docs/intro/quickstart/): This is used to install ratify components into the EKS cluster
- [docker](https://www.docker.com/get-started): This is used to build the container image we will deploy in this guide
- [ratify](https://github.com/deislabs/ratify/releases): This is used to check images from ECR locally
- [jq](https://stedolan.github.io/jq/): This is used to capture variables from json returned by commands
- [notation](https://github.com/notaryproject/notation): This is used to sign the container image we will deploy in this guide
- [AWS Signer notation plugin](https://docs.aws.amazon.com/signer/latest/developerguide/image-signing-prerequisites.html): this is required to use `notation` with AWS Signer resources

If you have not done so already, configure awscli to interact with your AWS account by following these [instructions](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-prereqs.html).

## Set Up ECR

We need to provision a public container repository to make our container images and their associated artifacts available to our EKS cluster. We will do this using awscli. For this guide we will be provisioning a public ECR repository to keep things simple.

```shell
export REPO_NAME=ratifydemo
export REPO_URI=$(aws ecr-public create-repository --repository-name $REPO_NAME --region us-east-1 | jq -r ."repository"."repositoryUri" )
```

We will use the repository URI returned by the create command later to build and tag the images we create.

For more information on provisioning ECR repositories check the [documentation](https://docs.aws.amazon.com/AmazonECR/latest/public/public-getting-started.html).

## Set up EKS

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

aws eks update-kubeconfig --name ratify-demo
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

## Sign Container Image

For this guide, we will sign the image using `notation` and AWS Signer resources. First, we will create a SigningProfile in AWS Signer and get the ARN:
```shell
aws signer put-signing-profile \
    --profile-name ratifyDemo \
    --platform-id Notation-OCI-SHA384-ECDSA

export PROFILE_ARN=$(aws signer get-signing-profile --profile-name ratifyDemo | jq .arn -r)
```

To use the SigningProfile in `notation`, we will add the profile as signing key:
```shell
notation key add \
    --plugin com.amazonaws.signer.notation.plugin \
    --id $PROFILE_ARN \
    --default ratifyDemo
```

After the profile has been added, we will use `notation` to sign the image with the SigningProfile:

```shell
notation sign $REPO_URI:v1
```

Both the container image and the signature should now be in the public ECR repository. We can also inspect the signature information using notation:

```shell
notation inspect $REPO_URI:v1
```

More information on signing can be found in the [AWS Signer](https://docs.aws.amazon.com/signer/latest/developerguide/Welcome.html) and [notation](https://github.com/notaryproject/notation) documentation.

## Deploy Gatekeeper

The Ratify container will perform the actual validation of images and their artifacts, but [Gatekeeper](https://github.com/open-policy-agent/gatekeeper) is used as the policy controller for Kubernetes.

We first need to install Gatekeeper into the cluster. We will use the Gatekeeper helm chart with some customizations:

```shell
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=5 \
    --set mutatingWebhookTimeoutSeconds=2
```

Next, we need to deploy a Gatekeeper policy and constraint. For this guide, we will use a sample policy and constraint that requires images to have at least one trusted signature.

```shell
kubectl apply -f https://raw.githubusercontent.com/deislabs/ratify/main/library/notation-validation/template.yml
kubectl apply -f https://raw.githubusercontent.com/deislabs/ratify/main/library/notation-validation/samples/constraint.yaml
```

More complex combinations of regos and Ratify verifiers can be used to accomplish many types of checks. See the [Gatekeeper docs](https://open-policy-agent.github.io/gatekeeper/website/docs/) for more information on rego authoring.

## Configure IAM Permissions

Before deploying Ratify, we need to configure permissions for Ratify to be able to make requests to AWS Signer. To do this we will use the IAM Roles for Service Accounts integration. First, we need to create an IAM policy that has AWS Signer permissions:
```shell
cat > signer_policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "signer:GetRevocationStatus"
            ],
            "Resource": "*"
        }
    ]
}
EOF

export POLICY_ARN=$(aws iam create-policy \
    --policy-name signerGetRevocationStatus \
    --policy-document file://signer_policy.json \
    | jq ."Policy"."Arn" -r)
```

Then, we will use `eksctl` to create a service account and role and attach the policies to the role:

```shell

eksctl create iamserviceaccount \
    --name ratify-admin \
    --namespace gatekeeper-system \
    --cluster ratify-demo \
    --attach-policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
    --attach-policy-arn $POLICY_ARN \
    --approve
```

We can validate that the service account was created by using `kubectl`:
```shell
kubectl -n gatekeeper-system get sa ratify-admin -oyaml
```

## Deploy Ratify

Now we can deploy Ratify to our cluster with the AWS Signer root as the notation verification certificate:
```shell
curl -sSLO https://d2hvyiie56hcat.cloudfront.net/aws-signer-notation-root.cert

helm install ratify \
    ratify/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notationCert=./aws-signer-notation-root.cert \
    --set featureFlags.RATIFY_EXPERIMENTAL_DYNAMIC_PLUGINS=true \
    --set serviceAccount.create=false \
    --set oras.authProviders.awsEcrBasicEnabled=true
```

After deploying Ratify, we will download the AWS Signer notation plugin to the Ratify pod using the [Dynamic Plugins feature](../reference/dynamic-plugins.md):

```shell
cat > aws-signer-plugin.yaml << EOF
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: aws-signer-plugin
spec:
  name: notation-com.amazonaws.signer.notation.plugin
  artifactTypes: application/vnd.oci.image.manifest.v1+json
  source:
    artifact: public.ecr.aws/aws-signer/notation-plugin:linux-amd64-latest
EOF

kubectl apply -f aws-signer-plugin.yaml
```

Finally, we will create a verifier that specifies the trust policy to use when verifying signatures. In this guide, we will use a trust policy that only trusts images signed by the SigningProfile we created earlier:

```shell
cat > notation-verifier.yaml << EOF
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-notation
spec:
  name: notation
  artifactTypes: application/vnd.cncf.notary.signature
  parameters:
    verificationCertStores:
      certs:
        - ratify-notation-inline-cert
    trustPolicyDoc:
      version: "1.0"
      trustPolicies:
        - name: default
          registryScopes:
            - "*"
          signatureVerification:
            level: strict
          trustStores:
            - signingAuthority:certs
          trustedIdentities:
            - $PROFILE_ARN
EOF

kubectl apply -f notation-verifier.yaml
```

More complex trust policies can be used to customize verification. See [notation documentation](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-policy) for more information on writing trust policies.

## Deploy Container Image

Now that the signed container image is in the registry and Ratify is installed into the EKS cluster we can deploy our
container image:

```shell
kubectl run demosigned --image $REPO_URI:v1
```

We should be able to see from the Ratify and Gatekeeper logs that the container signature was validated. The pod for the container should also be running.

```shell
kubectl logs -n gatekeeper-system deployment/ratify
```

We can also test that an image without a valid signature is not able to run:

```shell
kubectl run demounsigned --image hello-world
```

The command should fail with an error and we should be able to see from the Ratify and Gatekeeper logs that the signature validation failed.

## Cleaning Up

We can use awscli and eksctl to delete any resources created:

```shell
aws ecr-public delete-repository --region us-east-1 --repository-name $REPO_NAME

eksctl delete cluster --region us-east-1 --name ratify-demo

aws signer cancel-signing-profile --profile-name ratifyDemo
```

