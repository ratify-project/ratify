# Ratify with Private Container Registries
The default content store that is configured with Ratify is a container registry that is queried using the  ```oras``` store extension. This document outlines options to authenticate with the private container registries for querying supply chain content like signatures. 

As of now, ```oras``` supports two ways to authenticate with the private registries. 
- Providing Username and Password for the registry
- Retrieving the registry credentials from Docker Config 

> In future, ```oras``` will provide abstractions for authentication that enables providing credentials through other options like k8s secret, credential provider etc. 

For CLI mode of Ratify, the authentication should work automatically by doing ```docker login``` to the private registry. The below section describes the steps to configure this authentication for the k8s deployment of Ratify.

## Authentication using Docker credentials

- Create a k8s secret by providing credentials on the command line. This secret should be in the same namespace that contains Ratify deployment. 

```bash
kubectl create secret docker-registry ratify-regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```
 
- Deploy Ratify using helm
```bash
helm install ratify charts/ratify --set registryCredsSecret=ratify-regcred
```
> This mode can be used to authenticate with a single registry. If authentication to multiple registries is needed, docker config file can be used as described below

## Authentication using Local Docker Config file
If Docker config file is used for the registry login process, the same file can be used to create a k8s secret. 
- Deploy Ratify by specifying the path to the Docker config file.

> Note: If you use a Docker credentials store, you won't see that auth entry but a credsStore entry with the name of the store as value. In such cases, this option cannot be used. 

```bash
helm install ratify charts/ratify --set-file dockerConfig=<path to the local Docker config file>
```

Both the above modes uses a k8s secret of type ```dockerconfigjson``` that is described in the [document](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)