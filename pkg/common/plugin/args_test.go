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

func TestConcat_ReturnsExpected(t *testing.T) {
	testcases := []struct {
		input  [][2]string
		output string
	}{
		{
			input: [][2]string{
				{"testkey1", "testvalue1"},
				{"testkey2", "testvalue2"},
			},
			output: "testkey1=testvalue1;testkey2=testvalue2",
		},
		{
			input: [][2]string{
				{"testkey1", "testvalue1"},
			},
			output: "testkey1=testvalue1",
		},
		{
			input: [][2]string{
				{"testkey1", ""},
			},
			output: "testkey1=",
		},
	}

	for _, testcase := range testcases {
		actual := Concat(testcase.input)
		if actual != testcase.output {
			t.Fatalf("concat failed. expected: %v, actual: %v", testcase.output, actual)
		}
	}
}

func TestParseInputArgs_ReturnsExpected(t *testing.T) {
	testcases := []struct {
		output [][2]string
		input  string
	}{
		{
			output: [][2]string{
				{"testkey1", "testvalue1"},
				{"testkey2", "testvalue2"},
			},
			input: "testkey1=testvalue1;testkey2=testvalue2",
		},
		{
			output: [][2]string{
				{"testkey1", "testvalue1"},
			},
			input: "testkey1=testvalue1",
		},
		{
			output: [][2]string{
				{"testkey1", ""},
			},
			input: "testkey1=",
		},
		{
			output: [][2]string{},
			input:  "",
		},
	}

	for _, testcase := range testcases {
		actual, err := ParseInputArgs(testcase.input)
		if err != nil {
			t.Fatalf("parse input args failed %v", err)
		}
		if len(actual) != len(testcase.output) {
			t.Fatalf("parse input args failed. expected count: %v, actual count: %v", len(testcase.output), len(actual))
		}

		hasArg := func(key string, value string) bool {
			for _, args := range actual {
				if args[0] == key && args[1] == value {
					return true
				}
			}

			return false
		}

		for _, expectedArgs := range testcase.output {
			if !hasArg(expectedArgs[0], expectedArgs[1]) {
				t.Fatalf("cannot find expected arg %v", expectedArgs)
			}
		}
	}
}

func TestMergeDuplicateEnviron_ReturnsExpected(t *testing.T) {
	testcases := []struct {
		input       []string
		expectedLen int
		expectedEnv string
	}{
		{
			input:       []string{"testkey1=testvalue1", "testkey2=testvalue2"},
			expectedLen: 2,
			expectedEnv: "testkey1=testvalue1",
		},
		{
			input:       []string{"testkey1=testvalue1", "testkey2"},
			expectedLen: 2,
			expectedEnv: "testkey2",
		},
		{
			input:       []string{"testkey1=testvalue1", "testkey1=newtestvalue1"},
			expectedLen: 1,
			expectedEnv: "testkey1=newtestvalue1",
		},
		{
			input:       []string{"testkey1", "testkey1"},
			expectedLen: 2,
			expectedEnv: "testkey1",
		},
	}

	for _, testcase := range testcases {
		actual := MergeDuplicateEnviron(testcase.input)

		if len(actual) != testcase.expectedLen {
			t.Fatalf("mismatch count of returned env expected %v, actual %v", testcase.expectedLen, len(actual))
		}

		hasEnv := func() bool {
			for _, env := range actual {
				if env == testcase.expectedEnv {
					return true
				}
			}
			return false
		}

		if !hasEnv() {
			t.Fatalf("cannot find expected env %v in actual %v", testcase.expectedEnv, actual)
		}
	}
}
