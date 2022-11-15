# Ratify Dev Container

## Introduction

This repository is set up for development with [GitHub Codespaces](https://docs.github.com/en/codespaces/setting-up-your-project-for-codespaces/introduction-to-dev-containers) and VS Code [Dev Containers](https://code.visualstudio.com/docs/remote/containers).

## Features

- VS Code debug configuration
- One-time setup of Ratify config and certs
  - Initial configuration matches the quick start in the README
  - `.ratify` is symlinked to the Ratify config dir for easy access
- Use `.http` files under `.devcontainer` to execute sample HTTP requests

## Included Utilities

- Docker-in-Docker (with Docker Compose v2 support)
- protobuf compiler
- Helm
- KinD (Kubernetes in Docker)
- kubectl
- Bats (bash automated testing system)
- Notation
- Azure CLI
