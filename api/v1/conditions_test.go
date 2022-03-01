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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var transitionTime = metav1.Now()

var conditions = []Condition{
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
			if got := HasCondition(conditions, tt.desiredCondType); got != tt.expected {
				t.Errorf("Expected: %v Found: %v", tt.expected, got)
			}
		})
	}
}

func TestGetCondition(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        *Condition
	}{
		{
			title:           "condition is present in the condition list",
			desiredCondType: "type-1",
			expected: &Condition{
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
			if _, got := GetCondition(conditions, tt.desiredCondType); !equalCondition(tt.expected, got) {
				t.Errorf("Expected: %v Found: %v", tt.expected, got)
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	cases := []struct {
		title    string
		desired  Condition
		expected *Condition
	}{
		{
			title: "condition is not in the condition list",
			desired: Condition{
				Type:    "type-5",
				Status:  "True",
				Reason:  "Never seen before",
				Message: "New condition added",
			},
			expected: &Condition{
				Type:    "type-5",
				Status:  "True",
				Reason:  "Never seen before",
				Message: "New condition added",
			},
		},
		{
			title: "condition is in the condition list but not in desired state",
			desired: Condition{
				Type:               "type-1",
				Status:             "True",
				Reason:             "Updated",
				Message:            "Condition has changed",
				ObservedGeneration: 2,
			},
			expected: &Condition{
				Type:               "type-1",
				Status:             "True",
				Reason:             "Updated",
				Message:            "Condition has changed",
				ObservedGeneration: 2,
			},
		},
		{
			title: "condition is already in the desired state",
			desired: Condition{
				Type:    "type-4",
				Status:  "True",
				Reason:  "No reason",
				Message: "No msg",
			},
			expected: &Condition{
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
			newConditions := SetCondition(conditions, tt.desired)
			if _, got := GetCondition(newConditions, tt.desired.Type); !equalCondition(tt.expected, got) {
				t.Errorf("Expected: %v Found: %v", tt.expected, got)
			}
		})
	}
}

func TestRemoveCondition(t *testing.T) {
	cases := []struct {
		title           string
		desiredCondType string
		expected        *Condition
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
			newConditions := RemoveCondition(conditions, tt.desiredCondType)
			if _, got := GetCondition(newConditions, tt.desiredCondType); !equalCondition(tt.expected, got) {
				t.Errorf("Expected: %v Found: %v", tt.expected, got)
			}
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
			if got := IsConditionTrue(conditions, tt.desiredCondType); got != tt.expected {
				t.Errorf("Expected: %v Found: %v", tt.expected, got)
			}
		})
	}
}

func equalCondition(expected, got *Condition) bool {
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
