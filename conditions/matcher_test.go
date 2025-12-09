/*
Copyright 2020 The Kubernetes Authors.

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

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMatchConditions(t *testing.T) {
	testCases := []struct {
		name        string
		actual      any
		expected    kmapi.Conditions
		expectMatch bool
	}{
		{
			name:        "with an empty conditions",
			actual:      kmapi.Conditions{},
			expected:    kmapi.Conditions{},
			expectMatch: true,
		},
		{
			name: "with matching conditions",
			actual: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
			},
			expected: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
			},
			expectMatch: true,
		},
		{
			name: "with non-matching conditions",
			actual: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
			},
			expected: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
				{
					Type:               kmapi.ConditionType("different"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "different",
					Message:            "different",
				},
			},
			expectMatch: false,
		},
		{
			name: "with a different number of conditions",
			actual: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
			},
			expected: kmapi.Conditions{
				{
					Type:               kmapi.ConditionType("type"),
					Status:             metav1.ConditionTrue,
					Severity:           kmapi.ConditionSeverityNone,
					LastTransitionTime: metav1.Now(),
					Reason:             "reason",
					Message:            "message",
				},
			},
			expectMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			if tc.expectMatch {
				g.Expect(tc.actual).To(MatchConditions(tc.expected))
			} else {
				g.Expect(tc.actual).ToNot(MatchConditions(tc.expected))
			}
		})
	}
}

func TestMatchCondition(t *testing.T) {
	testCases := []struct {
		name        string
		actual      any
		expected    kmapi.Condition
		expectMatch bool
	}{
		{
			name:        "with an empty condition",
			actual:      kmapi.Condition{},
			expected:    kmapi.Condition{},
			expectMatch: true,
		},
		{
			name: "with a matching condition",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expectMatch: true,
		},
		{
			name: "with a different time",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Time{},
				Reason:             "reason",
				Message:            "message",
			},
			expectMatch: true,
		},
		{
			name: "with a different type",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("different"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expectMatch: false,
		},
		{
			name: "with a different status",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionFalse,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expectMatch: false,
		},
		{
			name: "with a different severity",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityInfo,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expectMatch: false,
		},
		{
			name: "with a different reason",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "different",
				Message:            "message",
			},
			expectMatch: false,
		},
		{
			name: "with a different message",
			actual: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			},
			expected: kmapi.Condition{
				Type:               kmapi.ConditionType("type"),
				Status:             metav1.ConditionTrue,
				Severity:           kmapi.ConditionSeverityNone,
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "different",
			},
			expectMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			if tc.expectMatch {
				g.Expect(tc.actual).To(MatchCondition(tc.expected))
			} else {
				g.Expect(tc.actual).ToNot(MatchCondition(tc.expected))
			}
		})
	}
}
