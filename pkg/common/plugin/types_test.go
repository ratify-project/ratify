package plugin

import "testing"

func TestError_ReturnsExpected(t *testing.T) {
	testError := NewError(123, "test error", "test err details")

	if testError.Error() != "test error; test err details" {
		t.Fatal("formatted error mismatches")
	}

	testErrorWithoutDetails := NewError(123, "test error", "")

	if testErrorWithoutDetails.Error() != "test error" {
		t.Fatal("formatted error without details mismatches")
	}
}
