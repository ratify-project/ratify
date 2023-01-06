package types

import "time"

// This class is based on implementation from  azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/blob/master/pkg/provider/
const (
	// VaultObjectTypeCertificate certificate vault object type
	VaultObjectTypeCertificate = "cert"

	CertTypePem = "application/x-pem-file"
	CertTypePfx = "application/x-pkcs12"

	CertificateType = "CERTIFICATE"

	ObjectFormatPEM = "pem"
	ObjectFormatPFX = "pfx"

	ObjectEncodingHex    = "hex"
	ObjectEncodingBase64 = "base64"
	ObjectEncodingUtf8   = "utf-8"

	// pod identity NMI port
	PodIdentityNMIPort = "2579"

	CSIAttributePodName              = "csi.storage.k8s.io/pod.name"
	CSIAttributePodNamespace         = "csi.storage.k8s.io/pod.namespace"
	CSIAttributeServiceAccountTokens = "csi.storage.k8s.io/serviceAccount.tokens" // nolint

	// KeyVaultNameParameter is the name of the key vault name parameter
	KeyVaultNameParameter = "keyvaultName"
	// CloudNameParameter is the name of the cloud name parameter
	CloudNameParameter = "cloudName"
	// TenantIDParameter is the name of the tenant ID parameter
	// TODO(aramase): change this from tenantId to tenantID after v1.2 release
	// ref: https://github.com/Azure/secrets-store-csi-driver-provider-azure/issues/857
	TenantIDParameter = "tenantID"
	// ClientIDParameter is the name of the client ID parameter
	// This clientID is used for workload identity
	ClientIDParameter = "clientID"
	// ObjectsParameter is the name of the objects parameter
	ObjectsParameter = "objects"
)

// KeyVaultCertificate holds keyvault object related config
type KeyVaultCertificate struct {
	// the name of the Azure Key Vault objects
	CertificateName string `json:"certificateName" yaml:"certificateName"`
	// the filename the object will be written to
	CertificateAlias string `json:"certificateAlias" yaml:"certificateAlias"`
	// the version of the Azure Key Vault objects
	CertificateVersion string `json:"certificateVersion" yaml:"certificateVersion"`
	// The number of versions to load for this secret starting at the latest version
	CertificateVersionHistory int32 `json:"certificateVersionHistory" yaml:"certificateVersionHistory"`
}

// CertificateFile holds content and metadata of a keyvault secret file
type CertificateFile struct {
	Content []byte
	Path    string //we are not writing the certificate to disk ,but we might need to path of the cert since notary trust policy allows for named trust store,
	Version string
}

// StringArray holds a list of strings
type StringArray struct {
	Array []string `json:"array" yaml:"array"`
}

// KeyVaultObjectVersion holds the version id and when that version was
// created for a specific version of a secret from KeyVault
type KeyVaultObjectVersion struct {
	Version string
	Created time.Time
}

// KeyVaultObjectVersionList holds a list of KeyVaultObjectVersion
type KeyVaultObjectVersionList []KeyVaultObjectVersion

func (list KeyVaultObjectVersionList) Len() int {
	return len(list)
}

func (list KeyVaultObjectVersionList) Less(i, j int) bool {
	return list[i].Created.After(list[j].Created)
}

func (list KeyVaultObjectVersionList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
