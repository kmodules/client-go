package meta

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var transitionTime = metav1.Now()

var conditions = []Condition{
	{
		Type:    "type-1",
		Status:  "True",
		Reason:  "No reason",
		Message: "No msg",
	}, {
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
				Type:    "type-1",
				Status:  "True",
				Reason:  "No reason",
				Message: "No msg",
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
			if got := GetCondition(conditions, tt.desiredCondType); !equalCondition(tt.expected, got) {
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
				Type:    "type-1",
				Status:  "False",
				Reason:  "Updated",
				Message: "Condition has changed",
			},
			expected: &Condition{
				Type:    "type-1",
				Status:  "False",
				Reason:  "Updated",
				Message: "Condition has changed",
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
			newConditions := SetCondition(conditions, tt.desired.Type, tt.desired.Status, tt.desired.Reason, tt.desired.Message)
			if got := GetCondition(newConditions, tt.desired.Type); !equalCondition(tt.expected, got) {
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
			desiredCondType: "type-1",
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
			if got := GetCondition(newConditions, tt.desiredCondType); !equalCondition(tt.expected, got) {
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
		expected.Message == got.Message {
		return true

	}
	return false
}
