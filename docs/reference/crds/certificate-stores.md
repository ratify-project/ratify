A `Certificate Store` resource defines an array of public certificates to fetch from a provider. 

View more CRD samples [here](../../../config/samples/). Each provider must specify the `name` of the certificate store.

```yml
provider: required, name of the certificate store provider
parameters: required, parameters specific to this certificate store provider
```
 
## AzureKeyVault Certificate Provider
See notary integration example [here](../../developer/verifier.md#section-6-built-in-verifiers)
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: CertificateStore
metadata:
  name: certstore-akv
spec:
  provider: azurekeyvault
  parameters:
    vaultURI: https://yourkeyvault.vault.azure.net/
    certificates:  |
      array:
        - |
          certificateName: yourCertName
          certificateVersion: yourCertVersion 
    tenantID:
    clientID: 
```

| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| vaultURI      | yes    |      URI of the azure key vault        |   ""            |
| certificateName      | yes    |    the name of the key vault object   |       ""        |
| certificateVersion   | no     |   provider will fetch latest version if empty   |     ""    |
| tenantID   | yes     |   tenantID of the workload identity that have read access to this key vault   |     ""    |
| clientID   | yes     |   clientID of the workload identity that have read access to this key vault   |     ""    |

### Limitation
Azure Key Vault integration currently only works for self signed certificate, we are following up on Azure Key Vault specific limitations so we could support certificate chains in the future, please use  [issue 695](https://github.com/deislabs/ratify/issues/695) for tracking. If you are working with a certificate chain, please specify the public root certificate value inline using the [inline certificate provider](certificate-stores.md#inline-certificate-provider). 

Please also ensure the certificate is in PEM format, PKCS12 format with nonexportable private keys can not be parsed due to limitation of Golang certificate library.

## Inline Certificate Provider
```
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: CertificateStore
metadata:
  name: certstore-inline
spec:
  provider: inline
  parameters:
    value: |
      -----BEGIN CERTIFICATE-----
      MIIDWDCCAkCgAwIBAgIBUTANBgkqhkiG9w0BAQsFADBaMQswCQYDVQQGEwJVUzEL
      MAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzANBgNVBAoTBk5vdGFyeTEb
      MBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMCAXDTIyMTIwMjA4MDg0NFoYDzIx
      MjIxMjAzMDgwODQ0WjBaMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNV
      BAcTB1NlYXR0bGUxDzANBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5l
      dHdvcmtzLmlvMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnoskJWB0
      ZsYcfbTvCYQMLqWaB/yN3Jf7Ryxvndrij83fWEQPBQJi8Mk8SpNqm2x9uP3gsQDc
      L/73a0p6/D+hza2jQQVhebe/oB0LJtUoD5LXlJ83UQdZETLMYAzeBNcBR4kMecrY
      CnE6yjHeiEWdAH+U7Mt39zJh+9lGIcbk0aUE5UOp8o3t5RWFDcl9hQ7QOXROwmpO
      thLUIiY/bcPpsg/2nH1nzFjqiBef3sgopFCTgtJ7qF8B83Xy/+hJ5vD29xsbSwuB
      3iLE7qLxu2NxdIa4oL0Y2QKMh/getjI0xnvwAmPkFiFbzC7LFdDfd6+gA5GpUXxL
      u6UmwucAgiljGQIDAQABoycwJTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYI
      KwYBBQUHAwMwDQYJKoZIhvcNAQELBQADggEBAFvRW/mGjnnMNFKJc/e3o/+yiJor
      dcrq/1UzyD7eNmOaASXz8rrrFT/6/TBXExPuB2OIf9OgRJFfPGLxmzCwVgaWQbK0
      VfTN4MQzRrSwPmNYsBAAwLxXbarYlMbm4DEmdJGyVikq08T2dZI51GC/YXEwzlnv
      ldN0dBflb/FKkY5rAp0JgpHLGKeStxFvB62noBjWfrm7ShCf9gkn1CjmgvP/sYK0
      pJgA1FHPd6EeB6yRBpLV4EJgQYUJoOpbHz+us62jKj5fAXsX052LPmk9ArmP0uJ1
      CJLNdj+aShCs4paSWOObDmIyXHwCx3MxCvYsFk/Wsnwura6jGC+cNsjzSx4=
      -----END CERTIFICATE-----

```

| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| value      | yes    |      public certificate content       |   ""            |