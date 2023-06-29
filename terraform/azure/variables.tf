variable "registry_name" {
  type = string
  default = "myregistry"
}

variable "key_vault_name" {
  type = string
  default = "mykeyvault"
}

variable "resource_group_name" {
  type = string
  default = "myresourcegroup"
}

variable "location" {
  type = string
  default = "eastus"
}

variable "identity_name" {
  type = string
  default = "myidentity"
}

variable "cluster_name" {
  type = string
  default = "mycluster"
}

variable "tags" {
  type = map(string)
}

variable "ratify_namespace" {
  type = string
  default = "gatekeeper-system"
}

variable "ratify_cert_name" {
  type = string
  default = "ratify"
}