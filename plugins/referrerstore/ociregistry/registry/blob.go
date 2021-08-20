package registry

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *Client) GetReferenceBlob(ref common.Reference, digest digest.Digest) ([]byte, oci.Descriptor, error) {
	blobBytes, desc, err := c.getReferenceBlob(ref, digest)

	if err != nil {
		return nil, oci.Descriptor{}, err
	}
	return blobBytes, *desc, nil

}

func (c *Client) getReferenceBlob(ref common.Reference, blobDigest digest.Digest) ([]byte, *oci.Descriptor, error) {
	scheme := "https"
	if c.plainHTTP {
		scheme = "http"
	}

	refParts := strings.Split(ref.Path, "/")

	url := fmt.Sprintf("%s://%s/v2/%s/blobs/%s",
		scheme,
		refParts[0],
		refParts[1],
		blobDigest.String(),
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid reference: %v", ref.Original)
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

	blob, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	digest, err := digest.SHA256.FromReader(bytes.NewReader(blob))
	if err != nil {
		return nil, nil, err
	}

	if digest.String() != blobDigest.String() {
		return nil, nil, fmt.Errorf("blob digest: %q does not match requested digest %q for %q", digest, blobDigest, ref.Path)
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
	return blob, &oci.Descriptor{
		MediaType: mediaType,
		Digest:    digest,
		Size:      size,
	}, nil
}
