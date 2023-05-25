/*
Copyright AppsCode Inc. and Contributors

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

package apiutil

import "testing"

func TestGetImageRef(t *testing.T) {
	type args struct {
		containerImage string
		statusImage    string
		statusImageID  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "appscode/scanner:extend_linux_amd64",
			args: args{
				containerImage: "appscode/scanner:extend_linux_amd64",
				statusImage:    "",
				statusImageID:  "",
			},
			want:    "index.docker.io/appscode/scanner:extend_linux_amd64",
			wantErr: false,
		},
		{
			name: "appscode/scanner:extend_linux_amd64",
			args: args{
				containerImage: "appscode/scanner:extend_linux_amd64",
				statusImage:    "docker.io/appscode/scanner:extend_linux_amd64",
				statusImageID:  "docker.io/library/import-2022-12-12@sha256:a21a96a7e93eed1d90b44d57c8b4a53608033a9858cc274561e930f0603acf1b",
			},
			want:    "index.docker.io/appscode/scanner:extend_linux_amd64",
			wantErr: false,
		},
		{
			name: "appscode/scanner:extend_linux_amd64",
			args: args{
				containerImage: "appscode/scanner:extend_linux_amd64",
				statusImage:    "docker.io/appscode/scanner:extend_linux_amd64",
				statusImageID:  "docker.io/appscode/scanner@sha256:a21a96a7e93eed1d90b44d57c8b4a53608033a9858cc274561e930f0603acf1b",
			},
			want:    "index.docker.io/appscode/scanner@sha256:a21a96a7e93eed1d90b44d57c8b4a53608033a9858cc274561e930f0603acf1b",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImageRef(tt.args.containerImage, tt.args.statusImage, tt.args.statusImageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetImageRef() got = %v, want %v", got, tt.want)
			}
		})
	}
}
