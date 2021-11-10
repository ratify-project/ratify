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

package core

import "errors"

var (
	ReferrerStoreNotFound = errors.New("Cannot find any referrer stores for the given subject")
	ReferrersNotFound     = errors.New("no referrers found for this artifact")
)

type ReferenceParseError struct {
	Enclosed string
}

func (err ReferenceParseError) Error() string {
	return err.Enclosed
}
