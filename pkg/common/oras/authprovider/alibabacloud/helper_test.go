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

package alibabacloud

import (
	"reflect"
	"testing"
)

func Test_getRegionFromArtifict(t *testing.T) {
	type args struct {
		artifact string
	}
	arg1 := "dahu-registry-vpc.cn-hangzhou.cr.aliyuncs.com"
	arg2 := "registry-vpc.cn-beijing.aliyuncs.com"
	arg3 := "test.bad"
	arg4 := "registry-vpc.cr.aliyuncs.com"

	tests := []struct {
		name    string
		args    args
		want    *AcrMetaInfo
		wantErr bool
	}{
		{"mock-test-get-region-from-artifict-1", args{arg1}, &AcrMetaInfo{
			InstanceName: "dahu",
			Region:       "cn-hangzhou",
		}, false},
		{"mock-test-get-region-from-artifict-2", args{arg2}, &AcrMetaInfo{
			Region: "cn-beijing",
		}, false},
		{"mock-test-get-region-from-artifict-3", args{arg3}, nil, true},
		{"mock-test-get-region-from-artifict-4", args{arg4}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegionFromArtifact(tt.args.artifact)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRegionFromArtifict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRegionFromArtifict() got = %v, want %v", got, tt.want)
			}
		})
	}
}
