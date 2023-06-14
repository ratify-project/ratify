output "resource_group_name" {
  value = azurerm_resource_group.rg.name
}

output "aks_cluster_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

output "key_vault_uri" {
  value = azurerm_key_vault.kv.vault_uri
}

output "key_vault_name" {
  value = azurerm_key_vault.kv.name
}

output "acr_name" {
  value = azurerm_container_registry.registry.name
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