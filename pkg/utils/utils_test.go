package utils

import (
	"strings"
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/opencontainers/go-digest"
)

func TestParseDigest_ReturnsExpected(t *testing.T) {
	dg, err := ParseDigest("sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb")

	if dg == "" || err != nil {
		t.Fatalf("digest parsing failed. expected digest but returned error %v", err)
	}

	dg, err = ParseDigest("sha256:c570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb")

	if dg != "" || err == nil {
		t.Fatal("digest parsing expected to fail but succeeded")
	}
}

func TestParseSubjectReference_ReturnsExpected(t *testing.T) {
	getDigest := func(dig string) digest.Digest {
		dg, _ := digest.Parse(dig)
		return dg
	}
	testcases := []struct {
		input          string
		output         common.Reference
		expectedErrMsg string
	}{
		{
			input: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
			output: common.Reference{
				Path:   "localhost:5000/net-monitor",
				Tag:    "v1",
				Digest: getDigest("sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb"),
			},
		},
		{
			input: "net-monitor@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
			output: common.Reference{
				Path:   "net-monitor",
				Digest: getDigest("sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb"),
			},
		},
		{
			input:          "/localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
			output:         common.Reference{},
			expectedErrMsg: "failed to parse subject reference",
		},
		{
			input:          "localhost:5000/net-monitor:v1",
			output:         common.Reference{},
			expectedErrMsg: "failed to parse subject reference digest",
		},
		{
			input:          "localhost:5000/net&monitor:v1",
			output:         common.Reference{},
			expectedErrMsg: "failed to parse subject reference",
		},
	}

	for _, testcase := range testcases {
		actual, err := ParseSubjectReference(testcase.input)
		if err != nil {
			if testcase.expectedErrMsg == "" {
				t.Fatalf("parsing subject reference failed with err %v", err)
			} else if !strings.Contains(err.Error(), testcase.expectedErrMsg) {
				t.Fatalf("parsing subject reference expected to fail with err %v actual %v", testcase.expectedErrMsg, err)
			}
		} else {
			if actual.Original != testcase.input {
				t.Fatalf("parsing subject reference failed expected original %v actual %v", testcase.input, actual.Original)
			}

			if actual.Path != testcase.output.Path {
				t.Fatalf("parsing subject reference failed expected path %v actual %v", testcase.output.Path, actual.Path)
			}

			if actual.Tag != testcase.output.Tag {
				t.Fatalf("parsing subject reference failed expected tag %v actual %v", testcase.output.Tag, actual.Tag)
			}

			if actual.Digest != testcase.output.Digest {
				t.Fatalf("parsing subject reference failed expected digest %v actual %v", testcase.output.Digest, actual.Digest)
			}
		}
	}
}
