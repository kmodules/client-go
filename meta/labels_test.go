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

package meta

import (
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelsForLabelSelector(t *testing.T) {
	tests := []struct {
		name       string
		selector   *metav1.LabelSelector
		expected   map[string]string
		exactMatch bool
	}{
		{
			name:       "nil selector",
			selector:   nil,
			expected:   map[string]string{},
			exactMatch: true,
		},
		{
			name: "match labels only",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "web", "tier": "frontend"},
			},
			expected:   map[string]string{"app": "web", "tier": "frontend"},
			exactMatch: true,
		},
		{
			name: "multiple match expressions",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "web", "stale": "true"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod", "staging"}},
					{Key: "debug", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"true"}},
					{Key: "team", Operator: metav1.LabelSelectorOpExists},
					{Key: "stale", Operator: metav1.LabelSelectorOpDoesNotExist},
					{Key: "track", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"canary", "stable"}},
				},
			},
			expected: map[string]string{
				"app":   "web",
				"env":   "prod",
				"debug": "false",
				"team":  "",
				"track": "not-canary",
			},
			exactMatch: false,
		},
		{
			name: "ignore empty values and invert false",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "web"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: nil},
					{Key: "flag", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"false"}},
					{Key: "zone", Operator: metav1.LabelSelectorOpNotIn, Values: []string{}},
				},
			},
			expected: map[string]string{
				"app":  "web",
				"flag": "true",
			},
			exactMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, exact := LabelsForLabelSelector(tc.selector)
			if exact != tc.exactMatch {
				t.Fatalf("expected exactMatch=%v, got %v", tc.exactMatch, exact)
			}
			if !apiequality.Semantic.DeepEqual(tc.expected, got) {
				t.Fatalf("expected labels=%v, got %v", tc.expected, got)
			}
		})
	}
}
