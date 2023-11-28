# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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