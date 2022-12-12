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
