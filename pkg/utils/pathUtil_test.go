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
package utils

import (
	"os"
	"testing"
)

func TestReadCertificatesFromPath_NestedDirectory(t *testing.T) {
	// Setup to create nested directory structure
	testDir := "TestDirectory"
	nestedDir := ".nestedFolder"
	testFile1 := testDir + string(os.PathSeparator) + "file1.txt"
	testFile2 := testDir + string(os.PathSeparator) + nestedDir + string(os.PathSeparator) + ".file2.crt"

	setupDiretoryForTesting(t, testDir)
	setupDiretoryForTesting(t, testDir+string(os.PathSeparator)+nestedDir)

	createFile(t, testFile1)
	createFile(t, testFile2)

	// Invoke method to test
	files, err := GetCertificatesFromPath(testDir)

	// Tear down
	os.RemoveAll(testDir)

	// Validate
	if len(files) != 1 || err != nil {
		t.Fatalf("File response length expected to be 1, actual %v, error %v", len(files), err)
	}

	if files[0] != testFile2 {
		t.Fatalf(" Expected file name %v, actual '%v'", testFile2, files[0])
	}
}

func TestReadFilesFromPath_SingleFile(t *testing.T) {
	// Setup
	testDir := "TestDirectory"
	testFile1 := testDir + string(os.PathSeparator) + "file1.Crt"

	setupDiretoryForTesting(t, testDir)
	createFile(t, testFile1)

	// Invoke method to test
	files, err := GetCertificatesFromPath(testDir)

	// Teardown
	os.RemoveAll(testDir)

	// validation
	// Validate
	if len(files) != 1 || err != nil {
		t.Fatalf("File response length expected to be 1, actual %v, error %v", len(files), err)
	}

	if files[0] != testFile1 {
		t.Fatalf(" Expected file name %v, actual '%v'", testFile1, files[0])
	}

}

func createFile(t *testing.T, path string) {
	_, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("write file '%s' failed with error '%v'", path, err)
	}
}

func setupDiretoryForTesting(t *testing.T, path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		t.Fatalf("Creating directory '%s' failed with '%v'", path, err)
	}
}
