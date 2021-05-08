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

package meta_test

import (
	"testing"
	"time"

	"kmodules.xyz/client-go/meta"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var (
	a1 = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:03:45Z'
    lastTransitionTime: '2021-05-08T19:03:45Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'True'
    lastUpdateTime: '2021-05-08T19:03:45Z'
    lastTransitionTime: '2021-05-08T19:03:45Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
	a1MissingCondition = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
`
	a1ConditionTimeUpdated = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:11:21Z'
    lastTransitionTime: '2021-05-08T19:11:21Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'True'
    lastUpdateTime: '2021-05-08T19:11:21Z'
    lastTransitionTime: '2021-05-08T19:11:21Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
	a1ConditionStatusUpdated = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: d1
  namespace: demo
spec:
  replicas: 3
status:
  observedGeneration: 2
  replicas: 3
  updatedReplicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
  - type: Available
    status: 'True'
    lastUpdateTime: '2021-05-08T19:12:51Z'
    lastTransitionTime: '2021-05-08T19:12:51Z'
    reason: MinimumReplicasAvailable
    message: Deployment has minimum availability.
  - type: Progressing
    status: 'False'
    lastUpdateTime: '2021-05-08T19:12:51Z'
    lastTransitionTime: '2021-05-08T19:12:51Z'
    reason: NewReplicaSetAvailable
    message: ReplicaSet "d1" has successfully progressed.
`
)

var (
	d1 = &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apps.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "demo",
			Name:      "d1",
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(3),
		},
		Status: apps.DeploymentStatus{
			ObservedGeneration:  2,
			Replicas:            3,
			UpdatedReplicas:     3,
			ReadyReplicas:       3,
			AvailableReplicas:   3,
			UnavailableReplicas: 0,
			Conditions: []apps.DeploymentCondition{
				{
					Type:               "Available",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now()),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability.",
				},
				{
					Type:               "Progressing",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now()),
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet \"d1\" has successfully progressed.",
				},
			},
		},
	}
	d1MissingCondition = &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apps.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "demo",
			Name:      "d1",
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(3),
		},
		Status: apps.DeploymentStatus{
			ObservedGeneration:  2,
			Replicas:            3,
			UpdatedReplicas:     3,
			ReadyReplicas:       3,
			AvailableReplicas:   3,
			UnavailableReplicas: 0,
			Conditions:          nil,
		},
	}
	d1ConditionTimeUpdated = &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apps.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "demo",
			Name:      "d1",
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(3),
		},
		Status: apps.DeploymentStatus{
			ObservedGeneration:  2,
			Replicas:            3,
			UpdatedReplicas:     3,
			ReadyReplicas:       3,
			AvailableReplicas:   3,
			UnavailableReplicas: 0,
			Conditions: []apps.DeploymentCondition{
				{
					Type:               "Available",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
					LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability.",
				},
				{
					Type:               "Progressing",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
					LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet \"d1\" has successfully progressed.",
				},
			},
		},
	}
	d1ConditionStatusUpdated = &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apps.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "demo",
			Name:      "d1",
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(3),
		},
		Status: apps.DeploymentStatus{
			ObservedGeneration:  2,
			Replicas:            3,
			UpdatedReplicas:     3,
			ReadyReplicas:       3,
			AvailableReplicas:   3,
			UnavailableReplicas: 0,
			Conditions: []apps.DeploymentCondition{
				{
					Type:               "Available",
					Status:             core.ConditionTrue,
					LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
					LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability.",
				},
				{
					Type:               "Progressing",
					Status:             core.ConditionFalse,
					LastUpdateTime:     metav1.NewTime(time.Now().Add(5 * time.Minute)),
					LastTransitionTime: metav1.NewTime(time.Now().Add(5 * time.Minute)),
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet \"d1\" has successfully progressed.",
				},
			},
		},
	}
)

func TestStatusConditionAwareEqual(t *testing.T) {
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
			name: "Map Same",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1),
			},
			want: true,
		},
		{
			name: "Map Missing Conditions",
			args: args{
				old: toUnstructured(a1MissingCondition),
				new: toUnstructured(a1MissingCondition),
			},
			want: true,
		},
		{
			name: "Map Condition Time Modified",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1ConditionTimeUpdated),
			},
			want: true,
		},
		{
			name: "Map Condition Status Modified",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1ConditionStatusUpdated),
			},
			want: false,
		},
		{
			name: "Struct Same",
			args: args{
				old: d1,
				new: d1,
			},
			want: true,
		},
		{
			name: "Struct Missing Conditions",
			args: args{
				old: d1MissingCondition,
				new: d1MissingCondition,
			},
			want: true,
		},
		{
			name: "Struct Condition Time Modified",
			args: args{
				old: d1,
				new: d1ConditionTimeUpdated,
			},
			want: true,
		},
		{
			name: "Struct Condition Status Modified",
			args: args{
				old: d1,
				new: d1ConditionStatusUpdated,
			},
			want: false,
		},
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
			if got := meta.StatusConditionAwareEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("StatusConditionAwareEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toUnstructured(s string) runtime.Object {
	data, err := yaml.YAMLToJSON([]byte(s))
	if err != nil {
		panic(err)
	}
	out, _, err := unstructured.UnstructuredJSONScheme.Decode(data, nil, nil)
	if err != nil {
		panic(err)
	}
	return out
}

func TestStatusEqual(t *testing.T) {
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
			name: "Map Same",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1),
			},
			want: true,
		},
		{
			name: "Map Missing Conditions",
			args: args{
				old: toUnstructured(a1MissingCondition),
				new: toUnstructured(a1MissingCondition),
			},
			want: true,
		},
		{
			name: "Map Condition Time Modified",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1ConditionTimeUpdated),
			},
			want: false,
		},
		{
			name: "Map Condition Status Modified",
			args: args{
				old: toUnstructured(a1),
				new: toUnstructured(a1ConditionStatusUpdated),
			},
			want: false,
		},
		{
			name: "Struct Same",
			args: args{
				old: d1,
				new: d1,
			},
			want: true,
		},
		{
			name: "Struct Missing Conditions",
			args: args{
				old: d1MissingCondition,
				new: d1MissingCondition,
			},
			want: true,
		},
		{
			name: "Struct Condition Time Modified",
			args: args{
				old: d1,
				new: d1ConditionTimeUpdated,
			},
			want: false,
		},
		{
			name: "Struct Condition Status Modified",
			args: args{
				old: d1,
				new: d1ConditionStatusUpdated,
			},
			want: false,
		},
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
			if got := meta.StatusEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("StatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
