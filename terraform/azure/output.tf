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
