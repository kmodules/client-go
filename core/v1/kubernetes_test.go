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

package v1

import (
	"testing"

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoveOwnerReference(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       "Deployment",
				Name:       "dep-0",
				APIVersion: "apps/v1",
				UID:        "0",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-1",
				APIVersion: "apps/v1",
				UID:        "1",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-2",
				APIVersion: "apps/v1",
				UID:        "2",
			},
		},
	}

	ref := metav1.ObjectMeta{
		Name: "dep-3",
		UID:  "3",
	}

	appendedMeta := objectMeta
	appendedMeta.OwnerReferences = append(appendedMeta.OwnerReferences, metav1.OwnerReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       ref.Name,
		UID:        ref.UID,
	})

	cases := []struct {
		testName string
		objMeta  metav1.ObjectMeta
		owner    metav1.ObjectMeta
		newMeta  metav1.ObjectMeta
	}{
		{
			"Reference is Not Owner of Object",
			objectMeta,
			ref,
			objectMeta,
		},
		{
			"Reference is Owner of Object",
			appendedMeta,
			ref,
			objectMeta,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			RemoveOwnerReference(&c.objMeta, &c.owner)
			if !apiequality.Semantic.DeepEqual(c.objMeta, c.newMeta) {
				t.Errorf("Remove of owner Reference is not successful, expected: %v. But Got: %v", c.newMeta, c.objMeta)
			}
		})
	}
}

func TestUpsertVolume(t *testing.T) {
	type args struct {
		volumes []core.Volume
		nv      []core.Volume
	}
	tests := []struct {
		name string
		args args
		want []core.Volume
	}{
		{
			name: "secret volume",
			args: args{
				volumes: []core.Volume{
					{
						Name: "auth",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "db-auth",
								Items: []core.KeyToPath{
									{
										Key:  "password",
										Path: "/etc/password",
										Mode: nil,
									},
								},
								DefaultMode: pointer.Int32P(0o420),
								Optional:    nil,
							},
						},
					},
				},
				nv: []core.Volume{
					{
						Name: "auth",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "db-auth",
								Items: []core.KeyToPath{
									{
										Key:  "username",
										Path: "/etc/username",
										Mode: nil,
									},
									{
										Key:  "password",
										Path: "/etc/password",
										Mode: nil,
									},
								},
								DefaultMode: nil,
								Optional:    nil,
							},
						},
					},
				},
			},
			want: []core.Volume{
				{
					Name: "auth",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: "db-auth",
							Items: []core.KeyToPath{
								{
									Key:  "username",
									Path: "/etc/username",
									Mode: nil,
								},
								{
									Key:  "password",
									Path: "/etc/password",
									Mode: nil,
								},
							},
							DefaultMode: pointer.Int32P(0o420),
							Optional:    nil,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpsertVolume(tt.args.volumes, tt.args.nv...); !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("UpsertVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
