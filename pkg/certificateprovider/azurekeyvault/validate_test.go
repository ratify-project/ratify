package azurekeyvault

import (
	"fmt"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	cases := []struct {
		desc        string
		fileName    string
		expectedErr error
	}{
		{
			desc:        "file name is absolute path",
			fileName:    "/secret1",
			expectedErr: fmt.Errorf("file name must be a relative path"),
		},
		{
			desc:        "file name contains '..'",
			fileName:    "secret1/..",
			expectedErr: fmt.Errorf("file name must not contain '..'"),
		},
		{
			desc:        "file name starts with '..'",
			fileName:    "../secret1",
			expectedErr: fmt.Errorf("file name must not contain '..'"),
		},
		{
			desc:        "file name is empty",
			fileName:    "",
			expectedErr: fmt.Errorf("file name must not be empty"),
		},
		{
			desc:        "valid file name",
			fileName:    "secret1",
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateFileName(tc.fileName)
			if tc.expectedErr != nil && err.Error() != tc.expectedErr.Error() || tc.expectedErr == nil && err != nil {
				t.Fatalf("expected err: %+v, got: %+v", tc.expectedErr, err)
			}
		})
	}
}
