A `Certificate Store` resource defines an array of public certificates to fetch from a provider. 

View more CRD samples [here](../../../config/samples/). Each provider must specify the `name` of the certificate store.

```yml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: CertificateStore
metadata:
  name:  
spec:
  provider: # required, name of the certificate store provider
  parameters: # required, parameters specific to this certificate store provider
status: # supported in version >= config.ratify.deislabs.io/v1beta1
  error:            # error message if the operation failed
  issuccess:        # boolean that indicate if operation was successful
  lastfetchedtime:  # timestamp of last attempted certificate fetch operation
  properties: # provider specific properties of the fetched certificates. If the current certificate fetch operation fails, this property displays the properties of last successfully cached certificate
```

# Certificate Store Provider
## AzureKeyVault Certificate Provider
See notation integration example [here](../../reference/verifier.md#section-6-built-in-verifiers)
```yml
apiVersion: config.ratify.deislabs.io/v1beta1
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
status:
  issuccess:        true
  lastfetchedtime:  # time stamp of last fetch operation
  properties: 
    certificates:
      certificate Name:  yourCertName
      last Refreshed:    # time stamp of last successful cert fetch operation
      version:           yourCertVersion 
```

| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| vaultURI      | yes    |      URI of the azure key vault        |   ""            |
| certificateName      | yes    |    the name of the key vault object   |       ""        |
| certificateVersion   | no     |   provider will fetch latest version if empty   |     ""    |
| tenantID   | yes     |   tenantID of the workload identity that have read access to this key vault   |     ""    |
| clientID   | yes     |   clientID of the workload identity that have read access to this key vault   |     ""    |

Use command `kubectl get certificatestores.config.ratify.deislabs.io` to see a overview of `certificatestores` status.
Use command `kubectl get certificatestores.config.ratify.deislabs.io/certstore-akv` to see full details on each certificate.
### Limitation
Azure keyvault Certificates are built on top of keys and secrets. When a certificate is created, an addressable key and secret are also created with the same name. Ratify requires secret permissions to retrieve the public certificates for the entire certificate chain, please set private keys to Non-exportable at certificate creation time to avoid security risk. Learn more about non-exportable keys [here](https://learn.microsoft.com/en-us/azure/key-vault/certificates/how-to-export-certificate?tabs=azure-cli#exportable-and-non-exportable-keys)

Please also ensure the certificate is in PEM format, PKCS12 format with nonexportable private keys can not be parsed due to limitation of Golang certificate library.

Akv set up guide in ratify-on-azure [quick start](https://github.com/deislabs/ratify/blob/main/docs/quickstarts/ratify-on-azure.md#configure-access-policy-for-akv).

> Note: If you were unable to configure certificate policy, please consider specifying the public root certificate value inline using the [inline certificate provider](../../reference/crds/certificate-stores.md#inline-certificate-provider) to reduce risk of exposing private key.

## Inline Certificate Provider
```
apiVersion: config.ratify.deislabs.io/v1beta1
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

# Certificate Specification
The main use case of certificate store is for notation verifier in Ratify, so users must follow the [TrustStore specification](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-store) defined by notation.

In brief, users must provide CA certificates or self-signed signing certificates, which means leaf certificates are not allowed to be used. Whatever certificates are provided, Ratify would keep only CA certificates and self-signed certificates. Therefore, if only leaf certificates are provided, Ratify would fail the verification directly since there are no valid certificates.

For `AzureKeyVault Certificate Provider`, users must use the self-signed certificates due to some limitations on the SDK API. If users want to configure a certificate chain from root to leaf, it's recommended to use the `Inline Certificate Provider` instead.

# CRD Resource Create/Update 
During the CRD creating/updating process, some matters require attention. The CRD operation could be successful even though some invalid values or typos are provided. Examples:

1. Invalid certificate value is provided in inline certificate provider.
2. Invalid vaultUri/certificateName or typos within them are provided in AKV certificate provider.

However, those invalid values or typos would cause failures while parsing certificates and signature verification in Ratify.
So it's recommended to check the CRD status once the CRD operation is done.