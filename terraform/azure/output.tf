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

output "key_vault_uri" {
  value = azurerm_key_vault.kv.vault_uri
}

output "ratify_certificate_name" {
  value = azurerm_key_vault_certificate.ratify-cert.name
}

output "tenant_id" {
  value = data.azurerm_client_config.current.tenant_id
}

output "workload_identity_client_id" {
  value = azurerm_user_assigned_identity.identity.client_id
}
