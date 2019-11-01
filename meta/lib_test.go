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
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithPrefix(tt.prefix, tt.name); got != tt.expected {
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
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithSuffix(tt.name, tt.suffix); got != tt.expected {
				t.Errorf("ValidNameWithSuffix() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidNameWithPefixNSuffix(t *testing.T) {
	testCases := []struct {
		title    string
		prefix   string
		name     string
		suffix   string
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
	}
	for _, tt := range testCases {
		t.Run(tt.title, func(t *testing.T) {
			if got := meta.ValidNameWithPefixNSuffix(tt.prefix, tt.name, tt.suffix); got != tt.expected {
				t.Errorf("ValidNameWithPefixNSuffix() = %v, want %v", got, tt.expected)
			}
		})
	}
}
