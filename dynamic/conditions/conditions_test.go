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
package conditions

import (
	"testing"

	kmapi "kmodules.xyz/client-go/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var transitionTime = metav1.Now()

var conds = []kmapi.Condition{
	{
		Type:               "type-1",
		Status:             "True",
		Reason:             "No reason",
		Message:            "No msg",
		ObservedGeneration: 1,
	},
	{
		Type:    "type-2",
		Status:  "False",
		Reason:  "No reason",
		Message: "No msg",
	},
	{
		Type:    "type-3",
		Status:  "Unknown",
		Reason:  "No reason",
		Message: "No msg",
	},
	{
		Type:               "type-4",
		Status:             "True",
		Reason:             "No reason",
		Message:            "No msg",
		LastTransitionTime: transitionTime,
	},
}

func newFoo() (*unstructured.Unstructured, error) {
	unstrConds := make([]interface{}, len(conds))
	for i := range conds {
		cond, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&conds[i])
		if err != nil {
			return nil, err
		}
		unstrConds[i] = cond
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "foo/v1",
			"kind":       "Foo",
			"metadata": map[string]interface{}{
				"name":      "foo-test",
				"namespace": "bar",
			},
			"spec": map[string]interface{}{
				"foo": "bar",
			},
			"status": map[string]interface{}{
				"conditions": unstrConds,
			},
		},
	}, nil
}

func runner(t *testing.T, check func(do DynamicOptions)) {
	obj, err := newFoo()
	if err != nil {
		t.Error(err)
		return
	}
	do := DynamicOptions{
		Client: fake.NewSimpleDynamicClient(runtime.NewScheme(), obj),
		GVR: schema.GroupVersionResource{
			Group:    "foo",
			Version:  "v1",
			Resource: "foos",
		},
		Name:      "foo-test",
		Namespace: "bar",
	}
	check(do)
}

func TestHasCondition(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        bool
	}{
		{
			title:           "condition is present in the condition list",
			desiredCondType: "type-1",
			expected:        true,
		},
		{
			title:           "condition is not present in the condition list",
			desiredCondType: "type-5",
			expected:        false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			runner(t, func(do DynamicOptions) {
				got, err := do.HasCondition(tt.desiredCondType)
				if err != nil {
					t.Error(t)
					return
				}
				if got != tt.expected {
					t.Errorf("Expected: %v Found: %v", tt.expected, got)
				}
			})
		})
	}
}

func TestGetCondition(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        *kmapi.Condition
	}{
		{
			title:           "condition is present in the condition list",
			desiredCondType: "type-1",
			expected: &kmapi.Condition{
				Type:               "type-1",
				Status:             "True",
				Reason:             "No reason",
				Message:            "No msg",
				ObservedGeneration: 1,
			},
		},
		{
			title:           "condition is not present in the condition list",
			desiredCondType: "type-5",
			expected:        nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			runner(t, func(do DynamicOptions) {
				_, got, err := do.GetCondition(tt.desiredCondType)
				if err != nil {
					t.Error(err)
					return
				}
				if !equalCondition(tt.expected, got) {
					t.Errorf("Expected: %v Found: %v", tt.expected, got)
				}
			})
		})
	}
}

func TestSetCondition(t *testing.T) {
	cases := []struct {
		title    string
		desired  kmapi.Condition
		expected *kmapi.Condition
	}{
		{
			title: "condition is not in the condition list",
			desired: kmapi.Condition{
				Type:    "type-5",
				Status:  "True",
				Reason:  "Never seen before",
				Message: "New condition added",
			},
			expected: &kmapi.Condition{
				Type:    "type-5",
				Status:  "True",
				Reason:  "Never seen before",
				Message: "New condition added",
			},
		},
		{
			title: "condition is in the condition list but not in desired state",
			desired: kmapi.Condition{
				Type:               "type-1",
				Status:             "True",
				Reason:             "Updated",
				Message:            "Condition has changed",
				ObservedGeneration: 2,
			},
			expected: &kmapi.Condition{
				Type:               "type-1",
				Status:             "True",
				Reason:             "Updated",
				Message:            "Condition has changed",
				ObservedGeneration: 2,
			},
		},
		{
			title: "condition is already in the desired state",
			desired: kmapi.Condition{
				Type:    "type-4",
				Status:  "True",
				Reason:  "No reason",
				Message: "No msg",
			},
			expected: &kmapi.Condition{
				Type:               "type-4",
				Status:             "True",
				Reason:             "No reason",
				Message:            "No msg",
				LastTransitionTime: transitionTime,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			runner(t, func(do DynamicOptions) {
				err := do.SetCondition(tt.desired)
				if err != nil {
					t.Error(err)
					return
				}
				_, got, err := do.GetCondition(string(tt.desired.Type))
				if err != nil {
					t.Error(err)
					return
				}
				if !equalCondition(tt.expected, got) {
					t.Errorf("Expected: %v Found: %v", tt.expected, got)
				}
			})
		})
	}
}

func TestRemoveCondition(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        *kmapi.Condition
	}{
		{
			title:           "condition is present in the condition list",
			desiredCondType: "type-2",
			expected:        nil,
		},
		{
			title:           "condition is not present in the condition list",
			desiredCondType: "type-5",
			expected:        nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			runner(t, func(do DynamicOptions) {
				err := do.RemoveCondition(tt.desiredCondType)
				if err != nil {
					t.Error(err)
					return
				}
				_, got, err := do.GetCondition(tt.desiredCondType)
				if err != nil {
					t.Error(err)
					return
				}

				if !equalCondition(tt.expected, got) {
					t.Errorf("Expected: %v Found: %v", tt.expected, got)
				}
			})
		})
	}
}

func TestIsConditionTrue(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        bool
	}{
		{
			title:           "condition is true",
			desiredCondType: "type-1",
			expected:        true,
		},
		{
			title:           "condition is false",
			desiredCondType: "type-2",
			expected:        false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			runner(t, func(do DynamicOptions) {
				got, err := do.IsConditionTrue(tt.desiredCondType)
				if err != nil {
					t.Error(err)
					return
				}
				if got != tt.expected {
					t.Errorf("Expected: %v Found: %v", tt.expected, got)
				}
			})
		})
	}
}

func equalCondition(expected, got *kmapi.Condition) bool {
	if expected == nil && got == nil {
		return true
	}
	if expected != nil &&
		got != nil &&
		expected.Type == got.Type &&
		expected.Status == got.Status &&
		expected.Reason == got.Reason &&
		expected.Message == got.Message &&
		expected.ObservedGeneration == got.ObservedGeneration {
		return true
	}
	return false
}
