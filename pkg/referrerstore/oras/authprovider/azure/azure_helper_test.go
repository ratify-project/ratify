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

Source: https://github.com/kubernetes-sigs/cloud-provider-azure/blob/3f223fe3931c6a6acfd01acab7c40bc3c5f75636/pkg/credentialprovider/azure_acr_helper.go
*/
package azure

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceiveChallengeFromLoginServer(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     []string
		httpStatusCode int
		want           *authDirective
		wantErr        error
	}{
		{
			name:           "Error should be returned when http status code is not 401",
			httpStatusCode: http.StatusOK,
			wantErr:        fmt.Errorf("registry did not issue a valid AAD challenge, status: 200"),
		},
		{
			name:           "Error should be returned when http response doesn't contain Www-Authenticate header",
			httpStatusCode: http.StatusUnauthorized,
			wantErr:        fmt.Errorf("challenge response does not contain header 'Www-Authenticate'"),
		},
		{
			name:           "Error should be returned when http response Www-Authenticate header doesn't have a valid value",
			httpStatusCode: http.StatusUnauthorized,
			authHeader:     []string{"a", "b"},
			wantErr:        fmt.Errorf("registry did not issue a valid AAD challenge, authenticate header [a, b]"),
		},
		{
			name:           "Error should be returned when http response Www-Authenticate header doesn't have a valid auth type",
			httpStatusCode: http.StatusUnauthorized,
			authHeader:     []string{`Auth realm="https://test.azurecr.io/oauth2/token",service="test.azurecr.io"`},
			wantErr:        fmt.Errorf("Www-Authenticate: expected realm: Bearer, actual: auth"),
		},
		{
			name:           "Error should be returned when http response Www-Authenticate header doesn't have a valid realm",
			httpStatusCode: http.StatusUnauthorized,
			authHeader:     []string{`Bearer realm="",service="test.azurecr.io"`},
			wantErr:        fmt.Errorf("Www-Authenticate: missing header \"realm\""),
		},
		{
			name:           "Error should be returned when http response Www-Authenticate header doesn't have a valid realm",
			httpStatusCode: http.StatusUnauthorized,
			authHeader:     []string{`Bearer realm="test.azurecr.io",service=""`},
			wantErr:        fmt.Errorf("Www-Authenticate: missing header \"service\""),
		},
		{
			name:           "Correct auth directive should be returned when everying is good",
			httpStatusCode: http.StatusUnauthorized,
			authHeader:     []string{`Bearer realm="https://test.azurecr.io/oauth2/token",service="test.azurecr.io"`},
			want: &authDirective{
				realm:   "https://test.azurecr.io/oauth2/token",
				service: "test.azurecr.io",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v2/", r.RequestURI)
				for _, h := range tt.authHeader {
					w.Header().Add("Www-Authenticate", h)
				}
				w.WriteHeader(tt.httpStatusCode)
			}))
			defer server.Close()

			host, err := url.Parse(server.URL)
			assert.NoError(t, err)
			got, err := receiveChallengeFromLoginServer(host.Host, "http")
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})

	}
}

func TestPerformTokenExchange(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		httpStatusCode int
		want           string
		wantErr        error
	}{
		{
			name:           "Error should be returned when http status code is not 200",
			httpStatusCode: http.StatusNotFound,
			wantErr:        fmt.Errorf("responded with status code 404"),
		},
		{
			name:           "Token should be returned if everything is good",
			token:          "token",
			httpStatusCode: http.StatusOK,
			want:           "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/oauth2/exchange", r.RequestURI)
				w.WriteHeader(tt.httpStatusCode)

				if tt.token != "" {
					_, err := w.Write([]byte(fmt.Sprintf(`{"refresh_token": "%s"}`, tt.token)))
					assert.NoError(t, err)
				}
			}))
			defer server.Close()

			got, err := performTokenExchange(server.URL, &authDirective{realm: server.URL}, "tenant", "token")
			assert.Equal(t, tt.want, got, tt.name)
			if tt.wantErr != nil {
				assert.Contains(t, err.Error(), tt.wantErr.Error(), tt.name)
			}
		})

	}
}

func TestParseAssignments(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *map[string]string
		wantErr error
	}{
		{
			name:    "error should be returned for empty string",
			in:      "",
			want:    nil,
			wantErr: fmt.Errorf("malformed header value: "),
		},
		{
			name: "correct values should be returned",
			in:   `a="b",c="d"`,
			want: &map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "correct values should be returned with spaces",
			in:   `a = "b", c ="d" `,
			want: &map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "correct values should be returned with empty value",
			in:   `a="",c="d" `,
			want: &map[string]string{"a": "", "c": "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAssignments(tt.in)
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}

func TestNextOccurrence(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{
			name: "-1 should be returned for empty string",
			in:   "",
			want: -1,
		},
		{
			name: "-1 should be returned for no space string",
			in:   "abc",
			want: -1,
		},
		{
			name: "correct index should be returned",
			in:   "abc def",
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextOccurrence(tt.in, 0, " ")
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestNextNoneSpace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{
			name: "-1 should be returned for empty string",
			in:   "",
			want: -1,
		},
		{
			name: "0 should be returned for no space string",
			in:   "abc",
			want: 0,
		},
		{
			name: "correct index should be returned for string with space prefix",
			in:   "   abc def",
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextNoneSpace(tt.in, 0)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
