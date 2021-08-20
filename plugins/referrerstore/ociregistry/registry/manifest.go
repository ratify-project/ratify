package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *Client) GetReferenceManifest(ref common.Reference) (ocispecs.ReferenceManifest, error) {
	manifestBytes, desc, err := c.getManifest(ref, ocispecs.MediaTypeArtifactManifest)

	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	if desc.MediaType != ocispecs.MediaTypeArtifactManifest {
		return ocispecs.ReferenceManifest{}, fmt.Errorf("manifest has a different mediatype. expected %s but got %s", ocispecs.MediaTypeArtifactManifest, desc.MediaType)
	}

	var manifest ocispecs.ReferenceManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	return manifest, nil

}

func (c *Client) getManifest(ref common.Reference, mediaTypes ...string) ([]byte, *oci.Descriptor, error) {
	scheme := "https"
	if c.plainHTTP {
		scheme = "http"
	}

	refParts := strings.Split(ref.Path, "/")

	url := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		scheme,
		refParts[0],
		refParts[1],
		ref.Digest.String(),
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid reference: %v", ref.Original)
	}
	//req.Header.Set("Connection", "close")
	for _, mediaType := range mediaTypes {
		req.Header.Add("Accept", mediaType)
	}

	resp, err := c.base.RoundTrip(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %v", url, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// no op
	case http.StatusUnauthorized, http.StatusNotFound:
		return nil, nil, fmt.Errorf("%v: %s", ref.Original, resp.Status)
	default:
		return nil, nil, fmt.Errorf("%v: %s", ref.Original, resp.Status)
	}

	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	digest, err := digest.SHA256.FromReader(bytes.NewReader(manifest))
	if err != nil {
		return nil, nil, err
	}

	if digest.String() != ref.Digest.String() {
		return nil, nil, fmt.Errorf("manifest digest: $q does not match requested digest %q for %q", digest, ref.Digest, ref.Path)
	}

	header := resp.Header
	mediaType := header.Get("Content-Type")
	if mediaType == "" {
		return nil, nil, fmt.Errorf("%v: missing Content-Type", url)
	}

	length := header.Get("Content-Length")
	if length == "" {
		return nil, nil, fmt.Errorf("%v: missing Content-Length", url)
	}
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: invalid Content-Length", url)
	}
	return manifest, &oci.Descriptor{
		MediaType: mediaType,
		Digest:    digest,
		Size:      size,
	}, nil
}
