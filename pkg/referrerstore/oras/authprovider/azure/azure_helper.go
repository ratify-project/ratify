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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	utilnet "k8s.io/apimachinery/pkg/util/net"
)

const (
	userAgentHeader = "User-Agent"
	userAgent       = "kubernetes-credentialprovider-acr"
	maxReadLength   = 10 * 1 << 20 // 10MB
)

var client = &http.Client{
	Transport: utilnet.SetTransportDefaults(&http.Transport{}),
	Timeout:   time.Second * 10,
}

type authDirective struct {
	service string
	realm   string
}

type acrAuthResponse struct {
	RefreshToken string `json:"refresh_token"`
}

// formats and sends a challenge request to the server address
func receiveChallengeFromLoginServer(serverAddress, scheme string) (*authDirective, error) {
	challengeURL := url.URL{
		Scheme: scheme,
		Host:   serverAddress,
		Path:   "v2/",
	}
	var err error
	var r *http.Request
	r, err = http.NewRequest("GET", challengeURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request, got %v", err)
	}
	r.Header.Add(userAgentHeader, userAgent)

	var challenge *http.Response
	if challenge, err = client.Do(r); err != nil {
		return nil, fmt.Errorf("error reaching registry endpoint %s, error: %s", challengeURL.String(), err)
	}
	defer challenge.Body.Close()

	if challenge.StatusCode != 401 {
		return nil, fmt.Errorf("registry did not issue a valid AAD challenge, status: %d", challenge.StatusCode)
	}

	var authHeader []string
	var ok bool
	if authHeader, ok = challenge.Header["Www-Authenticate"]; !ok {
		return nil, fmt.Errorf("challenge response does not contain header 'Www-Authenticate'")
	}

	if len(authHeader) != 1 {
		return nil, fmt.Errorf("registry did not issue a valid AAD challenge, authenticate header [%s]",
			strings.Join(authHeader, ", "))
	}

	authSections := strings.SplitN(authHeader[0], " ", 2)
	authType := strings.ToLower(authSections[0])
	var authParams *map[string]string
	if authParams, err = parseAssignments(authSections[1]); err != nil {
		return nil, fmt.Errorf("unable to understand the contents of Www-Authenticate header %s", authSections[1])
	}

	// verify headers
	if !strings.EqualFold("Bearer", authType) {
		return nil, fmt.Errorf("Www-Authenticate: expected realm: Bearer, actual: %s", authType)
	}
	if len((*authParams)["service"]) == 0 {
		return nil, fmt.Errorf("Www-Authenticate: missing header \"service\"")
	}
	if len((*authParams)["realm"]) == 0 {
		return nil, fmt.Errorf("Www-Authenticate: missing header \"realm\"")
	}

	return &authDirective{
		service: (*authParams)["service"],
		realm:   (*authParams)["realm"],
	}, nil
}

// exchanges the specifed access token for a registry refresh token
func performTokenExchange(
	serverAddress string,
	directive *authDirective,
	tenant string,
	accessToken string) (string, error) {
	var err error
	data := url.Values{
		"service":       []string{directive.service},
		"grant_type":    []string{"access_token_refresh_token"},
		"access_token":  []string{accessToken},
		"refresh_token": []string{accessToken},
		"tenant":        []string{tenant},
	}

	var realmURL *url.URL
	if realmURL, err = url.Parse(directive.realm); err != nil {
		return "", fmt.Errorf("Www-Authenticate: invalid realm %s", directive.realm)
	}
	authEndpoint := fmt.Sprintf("%s://%s/oauth2/exchange", realmURL.Scheme, realmURL.Host)

	datac := data.Encode()
	var r *http.Request
	r, err = http.NewRequest("POST", authEndpoint, bytes.NewBufferString(datac))
	if err != nil {
		return "", fmt.Errorf("failed to construct request, got %v", err)
	}
	r.Header.Add(userAgentHeader, userAgent)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(datac)))

	var exchange *http.Response
	if exchange, err = client.Do(r); err != nil {
		return "", fmt.Errorf("Www-Authenticate: failed to reach auth url %s", authEndpoint)
	}

	defer exchange.Body.Close()
	if exchange.StatusCode != 200 {
		return "", fmt.Errorf("Www-Authenticate: auth url %s responded with status code %d", authEndpoint, exchange.StatusCode)
	}

	var content []byte
	limitedReader := &io.LimitedReader{R: exchange.Body, N: maxReadLength}
	if content, err = ioutil.ReadAll(limitedReader); err != nil {
		return "", fmt.Errorf("Www-Authenticate: error reading response from %s", authEndpoint)
	}

	if limitedReader.N <= 0 {
		return "", errors.New("the read limit is reached")
	}

	var authResp acrAuthResponse
	if err = json.Unmarshal(content, &authResp); err != nil {
		return "", fmt.Errorf("Www-Authenticate: unable to read response %s", content)
	}

	return authResp.RefreshToken, nil
}

// Try and parse a string of assignments in the form of:
// key1 = value1, key2 = "value 2", key3 = ""
// Note: this method and handle quotes but does not handle escaping of quotes
func parseAssignments(statements string) (*map[string]string, error) {
	var cursor int
	result := make(map[string]string)
	var errorMsg = fmt.Errorf("malformed header value: %s", statements)
	for {
		// parse key
		equalIndex := nextOccurrence(statements, cursor, "=")
		if equalIndex == -1 {
			return nil, errorMsg
		}
		key := strings.TrimSpace(statements[cursor:equalIndex])

		// parse value
		cursor = nextNoneSpace(statements, equalIndex+1)
		if cursor == -1 {
			return nil, errorMsg
		}
		// case: value is quoted
		if statements[cursor] == '"' {
			cursor = cursor + 1
			// like I said, not handling escapes, but this will skip any comma that's
			// within the quotes which is somewhat more likely
			closeQuoteIndex := nextOccurrence(statements, cursor, "\"")
			if closeQuoteIndex == -1 {
				return nil, errorMsg
			}
			value := statements[cursor:closeQuoteIndex]
			result[key] = value

			commaIndex := nextNoneSpace(statements, closeQuoteIndex+1)
			if commaIndex == -1 {
				// no more comma, done
				return &result, nil
			} else if statements[commaIndex] != ',' {
				// expect comma immediately after close quote
				return nil, errorMsg
			} else {
				cursor = commaIndex + 1
			}
		} else {
			commaIndex := nextOccurrence(statements, cursor, ",")
			endStatements := commaIndex == -1
			var untrimmed string
			if endStatements {
				untrimmed = statements[cursor:commaIndex]
			} else {
				untrimmed = statements[cursor:]
			}
			value := strings.TrimSpace(untrimmed)

			if len(value) == 0 {
				// disallow empty value without quote
				return nil, errorMsg
			}

			result[key] = value

			if endStatements {
				return &result, nil
			}
			cursor = commaIndex + 1
		}
	}
}

// finds index of the next occurence of specified separator
func nextOccurrence(str string, start int, sep string) int {
	if start >= len(str) {
		return -1
	}
	offset := strings.Index(str[start:], sep)
	if offset == -1 {
		return -1
	}
	return offset + start
}

// finds the index of the next non-space character
func nextNoneSpace(str string, start int) int {
	if start >= len(str) {
		return -1
	}
	offset := strings.IndexFunc(str[start:], func(c rune) bool { return !unicode.IsSpace(c) })
	if offset == -1 {
		return -1
	}
	return offset + start
}
