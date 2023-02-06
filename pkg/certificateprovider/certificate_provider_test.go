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

package certificateprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCertificates(t *testing.T) {
	cases := []struct {
		desc          string
		pemString     string
		expectedErr   bool
		expectedCerts int
	}{
		{
			desc:        "empty string",
			expectedErr: false,
		},
		{
			desc:        "invalid certificate",
			pemString:   "-----BEGIN CERTIFICATE-----\nbaddata\n-----END CERTIFICATE-----\n",
			expectedErr: true,
		},
		{
			desc:          "single certificate",
			pemString:     "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIQXy2VqtlhSkiZKAGhsnkjbDANBgkqhkiG9w0BAQsFADBvMRswGQYDVQQD\nExJyYXRpZnkuZXhhbXBsZS5jb20xDzANBgNVBAsTBk15IE9yZzETMBEGA1UEChMKTXkgQ29tcGFu\neTEQMA4GA1UEBxMHUmVkbW9uZDELMAkGA1UECBMCV0ExCzAJBgNVBAYTAlVTMB4XDTIzMDIwMTIy\nNDUwMFoXDTI0MDIwMTIyNTUwMFowbzEbMBkGA1UEAxMScmF0aWZ5LmV4YW1wbGUuY29tMQ8wDQYD\nVQQLEwZNeSBPcmcxEzARBgNVBAoTCk15IENvbXBhbnkxEDAOBgNVBAcTB1JlZG1vbmQxCzAJBgNV\nBAgTAldBMQswCQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL10bM81\npPAyuraORABsOGS8M76Bi7Guwa3JlM1g2D8CuzSfSTaaT6apy9GsccxUvXd5cmiP1ffna5z+EFmc\nizFQh2aq9kWKWXDvKFXzpQuhyqD1HeVlRlF+V0AfZPvGt3VwUUjNycoUU44ctCWmcUQP/KShZev3\n6SOsJ9q7KLjxxQLsUc4mg55eZUThu8mGB8jugtjsnLUYvIWfHhyjVpGrGVrdkDMoMn+u33scOmrt\nsBljvq9WVo4T/VrTDuiOYlAJFMUae2Ptvo0go8XTN3OjLblKeiK4C+jMn9Dk33oGIT9pmX0vrDJV\nX56w/2SejC1AxCPchHaMuhlwMpftBGkCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeAMAkGA1UdEwQC\nMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAU0eaKkZj+MS9jCp9Dg1zdv3v/aKww\nHQYDVR0OBBYEFNHmipGY/jEvYwqfQ4Nc3b97/2isMA0GCSqGSIb3DQEBCwUAA4IBAQBNDcmSBizF\nmpJlD8EgNcUCy5tz7W3+AAhEbA3vsHP4D/UyV3UgcESx+L+Nye5uDYtTVm3lQejs3erN2BjW+ds+\nXFnpU/pVimd0aYv6mJfOieRILBF4XFomjhrJOLI55oVwLN/AgX6kuC3CJY2NMyJKlTao9oZgpHhs\nLlxB/r0n9JnUoN0Gq93oc1+OLFjPI7gNuPXYOP1N46oKgEmAEmNkP1etFrEjFRgsdIFHksrmlOlD\nIed9RcQ087VLjmuymLgqMTFX34Q3j7XgN2ENwBSnkHotE9CcuGRW+NuiOeJalL8DBmFXXWwHTKLQ\nPp5g6m1yZXylLJaFLKz7tdMmO355\n-----END CERTIFICATE-----\n",
			expectedCerts: 1,
		},
		{
			desc:          "certificate chain",
			pemString:     "-----BEGIN CERTIFICATE-----\nMIIEeDCCAmCgAwIBAgIUbztxbi/gSPMZGN53oZQZW1h/lbAwDQYJKoZIhvcNAQEL\nBQAwazELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdSZWRtb25k\nMRMwEQYDVQQKDApNeSBDb21wYW55MQ8wDQYDVQQLDAZNeSBPcmcxFzAVBgNVBAMM\nDmNhLmV4YW1wbGUuY29tMB4XDTIzMDIwMTIyNTMxMFoXDTI0MDIwMTIyNTMxMFow\nbTEZMBcGA1UEAxMQbGVhZi5leGFtcGxlLmNvbTEPMA0GA1UECxMGTXkgT3JnMRMw\nEQYDVQQKEwpNeSBDb21wYW55MRAwDgYDVQQHEwdSZWRtb25kMQswCQYDVQQIEwJX\nQTELMAkGA1UEBhMCVVMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCs\nMixf+WT34edYXs2c80zOg4Z/cxOVHU05gywjuISeaP+KS+Joc3emgbub1t5dPclk\nieIwrj3Olk3tvkrPiarOJIcrNR2zfBmQAufR4AUjoc4n1GQSp/voGgw1Hvh0wTkO\nYjhzLomrF242Ond8WTVO3Vq6/tfApfZMFM59eK9LMBkuvwTV4NeLnEnPvpLAoAvV\n9ZvCu7FuQ849R93Aoag2bZc3Tc3UCbahoJs9rTE/rnAqOhJWMGv2J1Y2Wu2eIvkD\n2uCmcVlY+7owG3TwLHTuIOBFl/5MXMvfR+B7yp1OkG23rTwwuSEBlMhYRzJFvssv\n8FX0sea7zhIg5dtoRjIlAgMBAAGjEjAQMA4GA1UdDwEB/wQEAwIHgDANBgkqhkiG\n9w0BAQsFAAOCAgEANRUu+9aBAuRf3OsdUWJflMAvuyzREp3skWSOUs4dw0MhcB6x\nb7BSyNdrBgPImLBpqYzNU6IT2eIlLXrYnKLehvPyZQx7LHvIeompA09aKMFFesqi\ngVoW5GRtp3qL3oNuuZJ80r/uKlB6Cj51TWqUbLcctBGHX7TWxFeWmFRnN0Bki00U\nJW/ElaTsr4GB+ltgZM+5USUqSNQqTa8t3d+vH6oVikyV1oYunM41xAfiRZtID04z\no15sLSkWTjavfmZ3+NjllipXFY2tnLqymCcObgdKtJHmTMFSDRngDjY+3+RVj4EY\npNaCCCepvtmXz1C5f06tlgY4ofaautJuAL7K93p/Q9ZcsIhmYWkCUZ0dkWq+eMdT\n9/lB9rQHbrDTaRxEQNIUezFMQEBxR9eC5JQfpw98LobAgA3r4vizQjQPsN0UZ54h\ncAHiyoo1VeckkotXaToRsoLjixPO9Fmss4H3urJTLpcU0drbVoG3emNh4K289vgR\nrjV11TenqvvR3+jJ2AX2ewSsF25m0afheZbrq2ZtyITPAbOqwMwTbTOvJT3HUztt\nhUP3qwsKNPR/hF3FSqZewiYOSqJi5Dk28Vd6mUEQzZa/Ma9RpBR+BAmfgH3Z9gX5\n0TqmAVQn1P8yh+bhEjiNa20bTJ+y5vQ9OrA7fiQ+6vpZCio4NFiEbYK4UBI=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIFxzCCA6+gAwIBAgIUY1fnGcYJFIGNk8fevKdGtOuZ5vcwDQYJKoZIhvcNAQEL\nBQAwazELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdSZWRtb25k\nMRMwEQYDVQQKDApNeSBDb21wYW55MQ8wDQYDVQQLDAZNeSBPcmcxFzAVBgNVBAMM\nDmNhLmV4YW1wbGUuY29tMB4XDTIzMDIwMTIyNTIwN1oXDTMzMDEyOTIyNTIwN1ow\nazELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdSZWRtb25kMRMw\nEQYDVQQKDApNeSBDb21wYW55MQ8wDQYDVQQLDAZNeSBPcmcxFzAVBgNVBAMMDmNh\nLmV4YW1wbGUuY29tMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtQXd\n+qySVVMHx7iGz9xRdpDKb+zATK3asMFnMXWBn0MCkcOMJvhjajA8yrCpfMudjiW0\nReQ6xcjEfnJqkzwxrK7tPE9cQ6JzQGCxsmscRzKf1NoJL2ske5xBteKuJhSfZ9la\ncIZn/EU2F6eMAl9U4Y+ncIIrl4UoB5H/AJJj62WMl5QAvzXXwCBwlLHQe4/T3Axu\n9xmD3HTC7iQExOUFLdJx86fK3ym0futi1RgOUgD+OrnyDEIkD8mGxffPYPgszS71\nUuJX+NTsLZ/JW3ER/PMAPnBsMsMTTxEIGnrp1CXf2RnwQnDHVsxZFMgkLTS6dksT\nTGevnulTNtVvSKsZ7MpzE01j7zDie4V4dQJzBJMktbeVq9KRoPIEX0WcKpg8bGS8\nd5p5pr+Lu/NOv6ur+av8M649qCPwJAv5i2P6ggT4YMNtY0wMD2kjcHJ9/l2gYpZj\n3DG0Hy3Xo8uKUmTSC7iGhLsSjleNhJkKyh3RCsuMKB9juE4qeXPoaoPWBIuarbkq\neVVZJu2PlgN8UcxM/ym+9GNJIfNJ18WGWm+5P3IDvfBbJ39yvzZlG2czju4BUzYn\nlPOHA/Z18TxZhlPrPVnyKSVbeg7sW17yMUI2LCeaFIOYdIFvM09RaCyLIGrQwhpe\nkLh56xXk702oNHaLxyh/v5kz8EyxnpXIlDntis0CAwEAAaNjMGEwHQYDVR0OBBYE\nFCESfoahHMx7GyBdTBARen7mc37nMB8GA1UdIwQYMBaAFCESfoahHMx7GyBdTBAR\nen7mc37nMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgEGMA0GCSqGSIb3\nDQEBCwUAA4ICAQBst1MbRJd3FLSY36qaNuDhGncvcXIcYqP2A/SXVhjuAhVOTsrE\nejDYBSkjfxCgoA3LZnQLTcbcPwRL5dTBEvqBqzbyArOjc+G3kbjOeuZYv09M8kUU\nYG1bVnxXV17GMk7zcBUnr1iwnp4E0pzB9gTv+Jv1oV+EtAHe5QOTOmW9mm2SUXbH\nmIya+KlzkIgVJJ8kiGOyXsr8i4wpyDXZf720tqTzPQFTf6rUXo9PhYzOWrrj8c7V\n+bmJurV3XkgvdOiwNase17wXUG9Ad8FhVYpUicq3Csfx5M87IXUIlx51AxaOQK3G\n3skyJyAm8R8pHzhcsEuVV7bGZlbFPZeWAHpbIEwpKHlLoN6qMINk2kEbcVahL2Mu\nXBUcvJdO2LbmEvfS+1imr32YbJ5Ufetru+G4mwAp6a0P6u5FU8lbE1ZoFHhflsVq\nErvOcRKhKjAim4iwIVGS55Xyx1IpF7YSTYpL89vlMmaJsssEoCQAkf+lxmC3eEuy\n2kBu3QB3cUHZXJK+01krC5MqeHcEVuc/fbNcTsCBp+RYXNRjYsp6IzGjqltUfInE\n3Tywg3P0ZLPO06WLNjNeAFZw6T8yV5gLcTJ1xc5pEf4UiY1uCf4NmDpeWhT+vvto\n2AxC/+7x7EkEfZnYiD6tcyHyY+iuroptws8lc5wRis859kydnq3vtxbXPQ==\n-----END CERTIFICATE-----\n",
			expectedCerts: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			certs, err := DecodeCertificates([]byte(tc.pemString))

			assert.Equal(t, tc.expectedErr, err != nil)
			assert.Equal(t, tc.expectedCerts, len(certs))
		})
	}
}

func TestDecodeCertificates_ByteArrayToCertificates(t *testing.T) {
	certString := "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQMdNmNTKwQ9aOe6iuMRokDzANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIy\nMTIxNDIxNTAzMVoXDTIzMTIxNDIyMDAzMVowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAOP6AHCFz41kRqsAiv6guFtQVsrzMgzoCX7o9NtQ57rr8BESP1LTGRAO\n4bjyP0i+at5uwIm4tdz0gW+g0P+f9bmfiScYgOFuxAJxLkMkBWPN3dJ9ulP/OGgB\n6mSCsEGreB3uaGc5rMbWCRaux65bMPjEzx5ex0qRSsn+fFMTwINPQUJpXSvi/W2/\n1umEWE1x59x0vlkP2dN7CXtB5/Bh01QNNbMdKU9saYn0kaBrCYZLwr6AxFRzLqLM\nQggy/6bOp/+cTTVqTiChMcdyIX52GRr2lChRsB34dDPYxDeKSI5LoRy07bveLjex\n4wm9+vx/WOSS5z0QPvE/v8avuIkMXR0CAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUwVvE\nvqQPxnE6j6pfX6jpSyv2dOAwHQYDVR0OBBYEFMFbxL6kD8ZxOo+qX1+o6Usr9nTg\nMA0GCSqGSIb3DQEBCwUAA4IBAQDE61FLbagvlCcXf0zcv+mUQ+0HvDVs7ofQe3Yw\naz7gAwxgTspr+jIFQWnPOOBupsyx/jucoz78ndbc5DGWPs2Qz/pIEGnLto2W/PYy\nas/9n8xHxembS4n/Mxxp60PF6ladi/nJAtDJds67sBeqLOfJzh6jV2uQvW7PXe1P\nOMSUHbBn8AfArZ/9njusiLs75+XcAgpnBFqKVv2Vd/INp2YQpVzusuiodeM8A9Qt\n/5xykjdCJw3ceZxD7dSkHgchKZPINFBYHt/EkN/d8mXFOKjGXZyntp4PO6PJ2HYN\nhMMDwdNu4mBmlMTdZMPEpIZIeW7G0P9KpCuvvD7po7NxdBgI\n-----END CERTIFICATE-----\n"

	c1 := []byte(certString)

	r, err := DecodeCertificates(c1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedLen := 1
	if len(r) != expectedLen {
		t.Fatalf("unexpected count of certificate, expected %+v, got %+v", expectedLen, len(r))
	}

	cert1 := r[0]
	serialNumber1 := cert1.SerialNumber.String()

	expectedserialNumber1 := "66229819451171253920043613209346319375"
	if serialNumber1 != expectedserialNumber1 {
		t.Fatalf("unexpected certificate, expected %+v, got %+v", expectedserialNumber1, serialNumber1)
	}
}

func TestDecodeCertificates_FailedToDecode(t *testing.T) {
	certString := "foo"

	c1 := []byte(certString)

	_, err := DecodeCertificates(c1)
	if err == nil {
		t.Fatalf("DecodeCertificates should return an error")
	}

	expectedError := "failed to decode pem block"
	if err.Error() != expectedError {
		t.Fatalf("unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

func TestDecodeCertificates_FailedX509ParseError(t *testing.T) {
	certString := "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQFJMQeqR8TRuHqNu+x0MuEDANBgkqhkiG9w0BAQsFABAD\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIz\nMDExMTE5MjAxMloXDTI0MDExMTE5MzAxMlowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAMh7F6sZyeiQRva83SvQu0PbsyD4zkEeWAyj03n1dx91FEeEXItCr+Y1\nghQKgdBOY/wJQmSq/We1e+17NoNICrzy2Y1sOVMYR5sx8H/UxO3q8oS7bxctFy+e\nHs4BxlHIqeIiWnz9bFAJFqV6BkJDVjp9k5QfHlkqH08WBvm/D8YTpWzvEPn+71ZG\nN1RKqFUeeM949oGGnC63vVMRRYIx2LoJliNZXdj9qoOHZksDrX2jkgPykkOYcmfo\n9CH9v0JNX+0t0Enp0ruUFK1pSZW+TicI22GvENYHGZNZ0m+6oD5ePRZoYhWyAzgZ\nndHO5bYh3yC7DMc6ssOEJeNN0I2+iLUCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUYhhf\nPFgAqU8PF3ClvfKs67HmpWwwHQYDVR0OBBYEFGIYXzxYAKlPDxdwpb3yrOux5qVs\nMA0GCSqGSIb3DQEBCwUAA4IBAQCXu1w+6s2RO2/KPmC+29m9EjbDReI4bGlDGgiv\nwk1fmvPvDrqL4Ebpcrb1nstNlsxpKYQP+3Vi8gPiqNQ7JvPStd1NBu+ViCXdvOe5\nCtN7tBFTCBgdgXNZ9bvIM2dS+xW/ZAJdyHbV9Hn77+rs/uCDHtbaQMJ3N9LGW8GR\nGY+uYylrrCrjb9fzotMaONnF9c1GKiANskc9371wbaninpxcwMNA5j027XzfnMEW\nm807wjlNV3Kuf4fdDpzBLe940iplfTlQMylWMqgANpEw4EqHCrBJPQAHfQEpQlo+\n9H72lrqOiYNNwApfB9P+UqMDi1B7T2yzfvXcqQ75FpSRIBAD\n-----END CERTIFICATE-----\n"

	c1 := []byte(certString)

	_, err := DecodeCertificates(c1)
	if err == nil {
		t.Fatalf("DecodeCertificates should return an error")
	}

	expectedError := "error parsing x509 certificate: x509: malformed issuer"
	if err.Error() != expectedError {
		t.Fatalf("unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}
