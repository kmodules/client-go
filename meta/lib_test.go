/*
Copyright The Kmodules Authors.

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
	"fmt"
	"reflect"
	"testing"

	"kmodules.xyz/client-go/meta"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var lblAphlict = map[string]string{
	"app": "AppAphlictserver",
}

func TestMarshalToYAML(t *testing.T) {
	obj := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "AppAphlictserver",
			Namespace: core.NamespaceDefault,
			Labels:    lblAphlict,
		},
		Spec: core.ServiceSpec{
			Selector: lblAphlict,
			Type:     core.ServiceTypeNodePort,
			Ports: []core.ServicePort{
				{
					Port:       int32(22280),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("client-server"),
					Name:       "client-server",
				},
				{
					Port:       int32(22281),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("admin-server"),
					Name:       "admin-server",
				},
			},
		},
	}

	b, err := meta.MarshalToYAML(obj, core.SchemeGroupVersion)
	fmt.Println(err)
	fmt.Println(string(b))
}

const domainKey = "kubedb.com"

func TestFilterKeys(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]string
		out  map[string]string
	}{
		{
			"IndexRune < 0",
			map[string]string{
				"k": "v",
			},
			map[string]string{
				"k": "v",
			},
		},
		{
			"IndexRune == 0",
			map[string]string{
				"/k": "v",
			},
			map[string]string{
				"/k": "v",
			},
		},
		{
			"IndexRune < n - xyz.abc/w1",
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
		},
		{
			"IndexRune < n - .abc/w1",
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
		},
		{
			"IndexRune == n - matching_domain",
			map[string]string{
				domainKey + "/w1": "v1",
				"w2":              "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain",
			map[string]string{
				"xyz." + domainKey + "/w1": "v1",
				"w2":                       "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain-2",
			map[string]string{
				"." + domainKey + "/w1": "v1",
				"w2":                    "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune == n - unmatched_domain",
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain",
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain-2",
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := meta.FilterKeys(domainKey, nil, c.in)
			if !reflect.DeepEqual(c.out, result) {
				t.Errorf("Failed filterTag test for '%v': expected %+v, got %+v", c.in, c.out, result)
			}
		})
	}
}

func TestValidNameWithPrefix(t *testing.T) {

	testCases := []struct {
		title    string
		prefix   string
		maxLen   []int
		name     string
		expected string
	}{
		{
			title:    "name empty",
			prefix:   "abc",
			name:     "",
			expected: "abc",
		},
		{
			title:    "name empty and prefix contain `-` in last",
			prefix:   "abc-",
			name:     "",
			expected: "abc",
		},
		{
			title:    "name and prefix contain 63 character",
			prefix:   "xyz",
			name:     "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1234567",
			expected: "xyz-aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1234567",
		},
		{
			title:    "name and prefix contain more than 63 character",
			prefix:   "xyz",
			name:     "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz123456789",
			expected: "xyz-aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1234567",
		},
		{
			title:    "custom maximum length",
			prefix:   "abcd",
			name:     "efgh",
			maxLen:   []int{6},
			expected: "abcd-e",
		},
		{
			title:    "custom maximum length with prefix len maxLen-1",
			prefix:   "abcd",
			name:     "efgh",
			maxLen:   []int{5},
			expected: "abcd",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithPrefix(tt.prefix, tt.name, tt.maxLen...); got != tt.expected {
				t.Errorf("ValidNameWithPrefix() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidNameWithSuffix(t *testing.T) {
	testCases := []struct {
		title    string
		name     string
		suffix   string
		maxLen   []int
		expected string
	}{
		{
			title:    "name empty",
			name:     "",
			suffix:   "abc",
			expected: "abc",
		},
		{
			title:    "name empty and suffix contain `-` in first",
			name:     "",
			suffix:   "-abc",
			expected: "abc",
		},
		{
			title:    "name and prefix contain 63 character",
			name:     "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1234567",
			suffix:   "abc",
			expected: "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1234567-abc",
		},
		{
			title:    "name and prefix contain more than 63 character",
			name:     "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz123456789",
			suffix:   "abc",
			expected: "bbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz123456789-abc",
		},
		{
			title:    "custom maximum length",
			name:     "abcd",
			suffix:   "efgh",
			maxLen:   []int{6},
			expected: "d-efgh",
		},
		{
			title:    "custom maximum length with suffix len maxLen-1",
			name:     "abcd",
			suffix:   "efgh",
			maxLen:   []int{5},
			expected: "efgh",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithSuffix(tt.name, tt.suffix, tt.maxLen...); got != tt.expected {
				t.Errorf("ValidNameWithSuffix() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidNameWithPrefixNSuffix(t *testing.T) {
	testCases := []struct {
		title    string
		prefix   string
		name     string
		suffix   string
		maxLen   []int
		expected string
	}{
		{
			title:    "name empty",
			prefix:   "xyz",
			name:     "",
			suffix:   "abc",
			expected: "xyz--abc",
		},
		{
			title:    "name and prefix contain 63 character",
			prefix:   "xyz",
			name:     "aabbccddeeffgghhiijjkkllmmnn123ooppqqrrssttuuvvwwxxyyzz",
			suffix:   "abc",
			expected: "xyz-aabbccddeeffgghhiijjkkllmmnn123ooppqqrrssttuuvvwwxxyyzz-abc",
		},
		{
			title:    "name and prefix contain more than 63 character",
			prefix:   "xyz",
			name:     "aabbccddeeffgghhiijjkkllmmnn123456789ooppqqrrssttuuvvwwxxyyzz",
			suffix:   "abc",
			expected: "xyz-aabbccddeeffgghhiijjkkllmmnn789ooppqqrrssttuuvvwwxxyyzz-abc",
		},
		{
			title:    "custom max length(even)",
			prefix:   "xyz",
			name:     "mn",
			suffix:   "abc",
			maxLen:   []int{8},
			expected: "xyz--abc",
		},
		{
			title:    "custom max length(odd)",
			prefix:   "xyz",
			name:     "mn",
			suffix:   "abc",
			maxLen:   []int{9},
			expected: "xyz-m-abc",
		},
		{
			title:    "custom max length with suffix+prefix == max len",
			prefix:   "xyz",
			name:     "mn",
			suffix:   "abc",
			maxLen:   []int{6},
			expected: "xyzabc",
		},
		{
			title:    "custom max length with suffix+prefix > max len",
			prefix:   "xyz",
			name:     "mn",
			suffix:   "abc",
			maxLen:   []int{4},
			expected: "xybc",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithPrefixNSuffix(tt.prefix, tt.name, tt.suffix, tt.maxLen...); got != tt.expected {
				t.Errorf("ValidNameWithPrefixNSuffix() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNameWithSuffix(t *testing.T) {
	type args struct {
		name         string
		suffix       string
		customLength int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "long suffix",
			args: args{
				name:         "vault",
				suffix:       "auth-controller",
				customLength: 11,
			},
			want: "controller",
		},
		{
			name: "suffix matched length",
			args: args{
				name:         "vault",
				suffix:       "controller",
				customLength: 10,
			},
			want: "controller",
		},
		{
			name: "starts with -",
			args: args{
				name:         "vault",
				suffix:       "controller",
				customLength: 11,
			},
			want: "controller",
		},
		{
			name: "trim name",
			args: args{
				name:         "vault",
				suffix:       "controller",
				customLength: 12,
			},
			want: "v-controller",
		},
		{
			name: "name-suffix matches length",
			args: args{
				name:         "vault",
				suffix:       "controller",
				customLength: 16,
			},
			want: "vault-controller",
		},
		{
			name: "name-suffix > length",
			args: args{
				name:         "vault",
				suffix:       "controller",
				customLength: 20,
			},
			want: "vault-controller",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.NameWithSuffix(tt.args.name, tt.args.suffix, tt.args.customLength); got != tt.want {
				t.Errorf("NameWithSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}
