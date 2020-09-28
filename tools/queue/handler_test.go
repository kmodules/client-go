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

package queue

import (
	"testing"

	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_statusEqual(t *testing.T) {
	type args struct {
		old interface{}
		new interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Object: old.status=empty, new.status=empty",
			args: args{
				old: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status:     apps.DeploymentStatus{},
				},
				new: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status:     apps.DeploymentStatus{},
				},
			},
			want: true,
		},
		{
			name: "Object: old.status=full, new.status=full",
			args: args{
				old: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status: apps.DeploymentStatus{
						ObservedGeneration: 1,
					},
				},
				new: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status: apps.DeploymentStatus{
						ObservedGeneration: 1,
					},
				},
			},
			want: true,
		},
		{
			name: "Object: old.status=empty, new.status=full",
			args: args{
				old: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status:     apps.DeploymentStatus{},
				},
				new: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status: apps.DeploymentStatus{
						ObservedGeneration: 1,
					},
				},
			},
			want: false,
		},
		{
			name: "Object: old.status=full, new.status=empty",
			args: args{
				old: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status: apps.DeploymentStatus{
						ObservedGeneration: 1,
					},
				},
				new: &apps.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       apps.DeploymentSpec{},
					Status:     apps.DeploymentStatus{},
				},
			},
			want: false,
		},
		{
			name: "Unstructured: old.status=missing, new.status=missing",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{},
				},
			},
			want: true,
		},
		{
			name: "Unstructured: old.status=nil, new.status=nil",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": nil,
					},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": nil,
					},
				},
			},
			want: true,
		},
		{
			name: "Unstructured: old.status=empty, new.status=empty",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{},
					},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{},
					},
				},
			},
			want: true,
		},
		{
			name: "Unstructured: old.status=full, new.status=full",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"observedGeneration": 1,
						},
					},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"observedGeneration": 1,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Unstructured: old.status=empty, new.status=full",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{},
					},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"observedGeneration": 1,
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Unstructured: old.status=full, new.status=empty",
			args: args{
				old: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"observedGeneration": 1,
						},
					},
				},
				new: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("statusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
