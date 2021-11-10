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
