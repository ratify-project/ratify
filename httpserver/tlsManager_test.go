/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpserver

import (
	"os"
	"path/filepath"
	"testing"
)

const firstCertFileName string = "firstCert.crt"
const firstKeyFileName string = "firstKey.key"
const firstCACertFileName string = "firstCACert.crt"
const firstCertificate string = `-----BEGIN CERTIFICATE-----
MIIDAjCCAeqgAwIBAgIUAbsAmO7kx6m5NgjlMLTvdHNqZtAwDQYJKoZIhvcNAQEL
BQAwKjEPMA0GA1UECgwGUmF0aWZ5MRcwFQYDVQQDDA5SYXRpZnkgUm9vdCBDQTAe
Fw0yMzA1MTgxOTQ5MzVaFw0yNDA1MTcxOTQ5MzVaMCMxITAfBgNVBAMMGHJhdGlm
eS5nYXRla2VlcGVyLXN5c3RlbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBANDh5yHzotQdPEZtI6S6Grq/85h29Q03iUbH5Bmc8MDKwl0QqX0UY1D1Ue4f
aFNJ+odD2yHdS6t1rCjEw/xdaaqVN82/d0/qBVoIfeUUBSKmrOQcaXIOzHidO2ci
++jPCRWApCay6k32t9G4NDZi99fwgybI19KNA1tmpQEU6pPRi37vljupplBSGFme
eSFPkmbXa/ECX0BjpAfs4if+8pzgIHqWvVT0Wl3SFLhXYfCQERupv/d+DGbTlTif
lWanDmZWLSSlPxgcgjjzQICU0Kj8WW7zxTSejwYCq7L+FL4XcWUurgqSxFNj41qT
cEeDFc1iWewLWCDA7D2wGjH43vUCAwEAAaMnMCUwIwYDVR0RBBwwGoIYcmF0aWZ5
LmdhdGVrZWVwZXItc3lzdGVtMA0GCSqGSIb3DQEBCwUAA4IBAQBRffIo/nY0JfNr
4IA3AI9fHkGKmwt2/tqEUe8paL+h42MWxa7FrSBMwcqOyhD+mrPDqC4CDqVOq5YY
ymzCuCbHsJqTohioPakFl80XTi03VZgMiySrWOwLtfBKSYxVhHI7jrR4VSLipxMu
Zu5gYBGyM/LEl2Mqg4+US5IMt9DsL83J8nMweIFv8Zsckz/Y4Gm9VGOAfKb2GUan
8etfPff+KRO5Gg4Pl/kwmOU+4vDpn+1PVyPzn3CVIDplcqcpeqjdV9YDBrlz0Ge+
Qh/RHtx6eyYD5PcLjlv9DO2SslR2kWYV0cJjTrlNGzscLrKSQj3vEJIJ5NKGvcvH
kxxpbB05
-----END CERTIFICATE-----`
const firstKey string = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDQ4ech86LUHTxG
bSOkuhq6v/OYdvUNN4lGx+QZnPDAysJdEKl9FGNQ9VHuH2hTSfqHQ9sh3Uurdawo
xMP8XWmqlTfNv3dP6gVaCH3lFAUipqzkHGlyDsx4nTtnIvvozwkVgKQmsupN9rfR
uDQ2YvfX8IMmyNfSjQNbZqUBFOqT0Yt+75Y7qaZQUhhZnnkhT5Jm12vxAl9AY6QH
7OIn/vKc4CB6lr1U9Fpd0hS4V2HwkBEbqb/3fgxm05U4n5Vmpw5mVi0kpT8YHII4
80CAlNCo/Flu88U0no8GAquy/hS+F3FlLq4KksRTY+Nak3BHgxXNYlnsC1ggwOw9
sBox+N71AgMBAAECggEBAJAhYF/wAgnExoN96VtPwwPbUVWBt6NQD9wUL5Nw1Drj
bWvUBG83MzR4ofjiGRVndYQCUWEzlnQP9SQIaYdoWXIIFoJUvBobS2gNdfkscEKx
qZiY9jVqerI7I/MNk67XtNfudNXzHHOBauM97GEetw98eLK5YRp6jLdzwyVU7mvh
qjlSO0PsYWhrRU4slxNsQMPy4aOetcavJFoidCZZbAt+LBVRrLDNRUf5rX99s+s4
8yp+fGgj6NwxSoDaps6cAYWmLCYk3c7fPbeLokP3Qnmi5cn74qkJ5Y39XdnSSyep
sQl7vDTl67DC4ujc2aVGIB3iasiq7qCF98phesdaQEECgYEA928v0ii4csJN/gqE
d6Bks755BZPpQz//0aiSVJ8MC7zng8BDQ407dHgHWofNxLsdHNkW9FpGL+R/wLSu
fj+1r7kX6iOwWjsJbAETF+hJfWc63Vo/dyeQo+azO1EPtTLlcrpum6rR3HO7fvh2
sTdPtsxdsE308fOqK4P46qRqs1ECgYEA2B0PttCaF6WP8z1Dm5e2ea92zqBwvaJ3
oTRRucKu2nXqRPj5GK2yzmWl1NGCaS8O9FqRK9OQFuK5PSjmN6gXHI6kH3BiOhHw
AG+ENlef6pV+E2KqPF8xjqqEUfKN2hEe2FRBfduFam0AIppUB5Y2Ebl0EagQXk9v
3i2YsyyWIGUCgYEAyXHwUP1uDaA7txQA/RPMaLot9WiShHnaYGsJl3NVb0kAg7dI
C/sz6ILAGehukjh0X0Qu+Al3Ew7JI672UTq1RLdAzRL5RLzD0vadAN3Q1xPwTL5o
5S2FCKuOSECatT8Wpu05l+reqMhgYeMPXwBVGdIQhLUzMrjaVks/oGjzpcECgYBt
P7Oz7RwYnB97DRtiSn16YlMi/URA+SKUoYg26c3OrhExsNLrwNNFN2lvfkH4vktH
B4mfqCGNECwoWMaYmCamzwz0v7FIPc0fy0AA4Kb8xXmofxYj0tOQlW6ypnVDKah4
H4/D+fcl59hLpcyY0TygFSoxys4LfwjEPjSVTxLNaQKBgQCnXKDWjqtgDwmvceQC
DH9dxDSdp/Jrz5GELXJuFzepxtGJS51CpPBV+2YOSyaSzEc9q6Usd4mSO6DCb2rM
A5p2U+ZMyVpdVxM+n2t8UxbqU8DK5Cko+6wR3q/HUXXNOtGYUrSQhQmW4y8ySyE6
7PGdPM31LcIuIjfPTG15FbNBTg==
-----END PRIVATE KEY-----
`
const firstCACert string = `-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIUODI+rSGQO2RuEyMftSqojjxcaRQwDQYJKoZIhvcNAQEL
BQAwKjEPMA0GA1UECgwGUmF0aWZ5MRcwFQYDVQQDDA5SYXRpZnkgUm9vdCBDQTAe
Fw0yMzA1MTgxOTQ5MzVaFw0yMzA1MTkxOTQ5MzVaMCoxDzANBgNVBAoMBlJhdGlm
eTEXMBUGA1UEAwwOUmF0aWZ5IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQDO43Y+WkYsH1jKoHtzdQMii2hdrXEzltXCLrM/P1OhERZyFSlZ
QubAwzNG41gT4/cBCRTafC45hw3aTKSCY7tphgGn6qXOelMLKu9yf5I/awukylHQ
F6+H+T40tWlRLT+bXFZtByIacZmoGoWzbZRmGBKoParNZCH3vfNvb7jYM4Ehb+JA
vPEKvz0Iu7l9jh0qgS6v9wx0GBz7vDFhTQHoVFXRcY9Vb9FdnNHEx+Lc7UoW0/kJ
LKTYqLsYJXZRqfmnAocpqxjaDODBzs9+k6ulrrqoZjyZmkbVbaFSLLvIVAeHzuwL
0gQbz8OEIP+uqH+7VhGb6vbKknAMwja+EbHDAgMBAAGjUzBRMB0GA1UdDgQWBBSD
TaR4b0ju72jRo1RMffIHt3fIcDAfBgNVHSMEGDAWgBSDTaR4b0ju72jRo1RMffIH
t3fIcDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBGVtPpetE8
Fe9aAN/ZLmaIj14FfvAy7LOSv1AIDNFgNLTbo25z+IH7cQf1Ws967BqlaW79gg8y
k9en7wfLBy3G/teU00bOy2eYtPIsVwLgrpIbVU6VYeUg1i86/pk/tNzOiDRj06qO
Dztv/HJIGGcLwj1T2yiiFnizgRlwmUSVi1BDu4c34A0P9OGoZPSZ97/hhHauK81o
WqlYtBTNTFx4eisM277wPKV8iHNGqULVGtADGmU0WWsWPDmHytmBaE5tF/NifRrF
1oSTxUCtw319hyHUFX5xHq2REEN+tbxTUbVsgXVTrZmnuSxN2WFAfvuPVKxQ9RD9
ve39ZodlqIAz
-----END CERTIFICATE-----
`

func TestNewTLSCertWatcher_Expected(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-certs")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certFileName := filepath.Join(tmpDir, firstCertFileName)
	if err = os.WriteFile(certFileName, []byte(firstCertificate), 0600); err != nil {
		t.Fatalf("cert file creation failed %v", err)
	}
	keyFileName := filepath.Join(tmpDir, firstKeyFileName)
	if err = os.WriteFile(keyFileName, []byte(firstKey), 0600); err != nil {
		t.Fatalf("cert key creation failed %v", err)
	}
	caFileName := filepath.Join(tmpDir, firstCACertFileName)
	if err = os.WriteFile(caFileName, []byte(firstCACert), 0600); err != nil {
		t.Fatalf("ca cert file creation failed %v", err)
	}
	var cw *TLSCertWatcher
	_, err = NewTLSCertWatcher("", "", "")
	if err == nil {
		t.Errorf("Expected error, got %v", err)
	}
	cw, err = NewTLSCertWatcher(certFileName, keyFileName, caFileName)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cw.ratifyServerCertPath != certFileName {
		t.Errorf("Expected %s, got %s", certFileName, cw.ratifyServerCertPath)
	}
	if cw.ratifyServerKeyPath != keyFileName {
		t.Errorf("Expected %s, got %s", keyFileName, cw.ratifyServerKeyPath)
	}
	if cw.clientCACertPath != caFileName {
		t.Errorf("Expected %s, got %s", caFileName, cw.clientCACertPath)
	}

	if cw.ratifyServerCert == nil {
		t.Errorf("Expected ratifyServerCert to be set")
	}
	if cw.clientCACert == nil {
		t.Errorf("Expected clientCACert to be set")
	}
	if cw.watcher == nil {
		t.Errorf("Expected watcher to be set")
	}
}

func TestReadCertificates_Expected(t *testing.T) {
	// setup temp dir with test certs/key
	tmpDir, err := os.MkdirTemp("", "test-certs")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certFileName := filepath.Join(tmpDir, firstCertFileName)
	if err = os.WriteFile(certFileName, []byte(firstCertificate), 0600); err != nil {
		t.Fatalf("cert file creation failed %v", err)
	}
	keyFileName := filepath.Join(tmpDir, firstKeyFileName)
	if err = os.WriteFile(keyFileName, []byte(firstKey), 0600); err != nil {
		t.Fatalf("cert key creation failed %v", err)
	}
	caFileName := filepath.Join(tmpDir, firstCACertFileName)
	if err = os.WriteFile(caFileName, []byte(firstCACert), 0600); err != nil {
		t.Fatalf("ca cert file creation failed %v", err)
	}

	// test with empty cert/key paths
	var cw *TLSCertWatcher
	cw = &TLSCertWatcher{}
	if err := cw.ReadCertificates(); err == nil {
		t.Errorf("Expected error, got %v", err)
	}

	// test with invalid ca cert path
	cw.ratifyServerCertPath = "invalid"
	cw.clientCACertPath = "invalid"
	cw.ratifyServerKeyPath = "invalid"
	if err := cw.ReadCertificates(); err == nil {
		t.Errorf("Expected error, got %v", err)
	}

	// test with invalid cert/key paths
	cw.clientCACertPath = caFileName
	if err := cw.ReadCertificates(); err == nil {
		t.Errorf("Expected error, got %v", err)
	}
}
