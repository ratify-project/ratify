#!/bin/bash

echo "Starting Ratify on Azure\n"
echo "RESOURCE_GROUP: $RESOURCE_GROUP\n"
echo "CLUSTER_NAME: $CLUSTER_NAME\n"
echo "ENABLE_MUTATION: $ENABLE_MUTATION\n"
echo "ENABLE_CERT_ROTATION: $ENABLE_CERT_ROTATION\n"
echo "MANAGED_IDENTITY_CLIENT_ID: $MANAGED_IDENTITY_CLIENT_ID\n"

# Authenticate with managed identity
az login --identity --username $MANAGED_IDENTITY_CLIENT_ID

az aks get-credentials --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --overwrite-existing
