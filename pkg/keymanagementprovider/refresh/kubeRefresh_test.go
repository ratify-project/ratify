package refresh

//TODO blank import of provider for intis

import (
	"context"
	"testing"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/inline"
	test "github.com/ratify-project/ratify/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//TODO: move helper tests for controller to here
func TestKubeRefresher_Refresh(t *testing.T) {
	tests := []struct {
		name    string
		provider *configv1beta1.KeyManagementProvider
		request ctrl.Request
		wantErr bool
	}{
		{
			name: "valid params",
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "" ,
					Name:      "kmpName",
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type:       "inline",
					Parameters: runtime.RawExtension{
						Raw: []byte(`{"type": "inline", "contentType": "certificate", "value": "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIQXy2VqtlhSkiZKAGhsnkjbDANBgkqhkiG9w0BAQsFADBvMRswGQYDVQQD\nExJyYXRpZnkuZXhhbXBsZS5jb20xDzANBgNVBAsTBk15IE9yZzETMBEGA1UEChMKTXkgQ29tcGFu\neTEQMA4GA1UEBxMHUmVkbW9uZDELMAkGA1UECBMCV0ExCzAJBgNVBAYTAlVTMB4XDTIzMDIwMTIy\nNDUwMFoXDTI0MDIwMTIyNTUwMFowbzEbMBkGA1UEAxMScmF0aWZ5LmV4YW1wbGUuY29tMQ8wDQYD\nVQQLEwZNeSBPcmcxEzARBgNVBAoTCk15IENvbXBhbnkxEDAOBgNVBAcTB1JlZG1vbmQxCzAJBgNV\nBAgTAldBMQswCQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL10bM81\npPAyuraORABsOGS8M76Bi7Guwa3JlM1g2D8CuzSfSTaaT6apy9GsccxUvXd5cmiP1ffna5z+EFmc\nizFQh2aq9kWKWXDvKFXzpQuhyqD1HeVlRlF+V0AfZPvGt3VwUUjNycoUU44ctCWmcUQP/KShZev3\n6SOsJ9q7KLjxxQLsUc4mg55eZUThu8mGB8jugtjsnLUYvIWfHhyjVpGrGVrdkDMoMn+u33scOmrt\nsBljvq9WVo4T/VrTDuiOYlAJFMUae2Ptvo0go8XTN3OjLblKeiK4C+jMn9Dk33oGIT9pmX0vrDJV\nX56w/2SejC1AxCPchHaMuhlwMpftBGkCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeAMAkGA1UdEwQC\nMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAU0eaKkZj+MS9jCp9Dg1zdv3v/aKww\nHQYDVR0OBBYEFNHmipGY/jEvYwqfQ4Nc3b97/2isMA0GCSqGSIb3DQEBCwUAA4IBAQBNDcmSBizF\nmpJlD8EgNcUCy5tz7W3+AAhEbA3vsHP4D/UyV3UgcESx+L+Nye5uDYtTVm3lQejs3erN2BjW+ds+\nXFnpU/pVimd0aYv6mJfOieRILBF4XFomjhrJOLI55oVwLN/AgX6kuC3CJY2NMyJKlTao9oZgpHhs\nLlxB/r0n9JnUoN0Gq93oc1+OLFjPI7gNuPXYOP1N46oKgEmAEmNkP1etFrEjFRgsdIFHksrmlOlD\nIed9RcQ087VLjmuymLgqMTFX34Q3j7XgN2ENwBSnkHotE9CcuGRW+NuiOeJalL8DBmFXXWwHTKLQ\nPp5g6m1yZXylLJaFLKz7tdMmO355\n-----END CERTIFICATE-----\n"}`),
					},
				},
			},
			request: ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: "",
					Name:      "kmpName",
				},
			},
			wantErr: false,
		},
		{
			name: "nonexistent KMP",
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "" ,
					Name:      "kmpName",
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type:       "inline",
				},
			},
			request: ctrl.Request{
				NamespacedName: client.ObjectKey{
					Name:      "nonexistent",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid params",
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "" ,
					Name:      "kmpName",
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type:       "inline",
				},
			},
			request: ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: "",
					Name:      "kmpName",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, _ := test.CreateScheme()
			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.provider).Build()
			kr := &KubeRefresher{
				Client:  client,
				Request: tt.request,
			}
			if err := kr.Refresh(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("KubeRefresher.Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}