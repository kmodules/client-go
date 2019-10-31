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

func TestGetValidNameWithFixedPrefix(t *testing.T) {
	type args struct {
		prefix string
		str    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				prefix: "",
				str:    "",
			},
			want: "",
		},
		{
			name: "equal to 64 char",
			args: args{
				prefix: "xyz",
				str:    "1122aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1122",
			},
			want: "xyz-1122aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz1122",
		},
		{
			name: "less than 64 char",
			args: args{
				prefix: "xyz-",
				str:    "aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz",
			},
			want: "xyz-aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz",
		},
		{
			name: "gather than 64 char",
			args: args{
				prefix: "xyz-",
				str:    "1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122444",
			},
			want: "xyz-1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.GetValidNameWithFixedPrefix(tt.args.prefix, tt.args.str); got != tt.want {
				t.Errorf("GetValidNameWithFixedPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValidNameWithFixedSuffix(t *testing.T) {
	type args struct {
		suffix string
		str    string
	}
	tests := []struct {
		testCase string
		prefix string
		name string
		suffix string
		expected string
	}{
		{
			name: "empty",
			args: args{
				suffix: "",
				str:    "",
			},
			want: "",
		},
		{
			name: "equal to 64 char",
			args: args{
				suffix: "-abc",
				str:    "1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122",
			},
			want: "1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122-abc",
		},
		{
			name: "less than 64 char",
			args: args{
				suffix: "-abc",
				str:    "aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz",
			},
			want: "aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz-abc",
		},
		{
			name: "gather than 64 char",
			args: args{
				suffix: "-abc",
				str:    "1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122444",
			},
			want: "1122aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz1122-abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.GetValidNameWithFixedSuffix(tt.args.suffix, tt.args.str); got != tt.want {
				t.Errorf("GetValidNameWithFixedSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValidNameWithFixedPefixNSuffix(t *testing.T) {
	type args struct {
		prefix string
		suffix string
		str    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				prefix: "",
				str:    "",
				suffix: "",
			},
			want: "",
		},
		{
			name: "equal to 64 char",
			args: args{
				prefix: "xyz-",
				str:    "12aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz12",
				suffix: "-abc",
			},
			want: "xyz-12aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz12-abc",
		},
		{
			name: "less than 64 char",
			args: args{
				prefix: "xyz-",
				str:    "aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz",
				suffix: "-abc",
			},
			want: "xyz-aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz-abc",
		},
		{
			name: "gather than 64 char",
			args: args{
				prefix: "xyz-",
				str:    "12aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz12444",
				suffix: "-abc",
			},
			want: "xyz-12aabbccddeeffgghhiijjkklmmnnooppqqrrssttuuvvwwxxyyzz12-abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.GetValidNameWithFixedPefixNSuffix(tt.args.prefix, tt.args.suffix, tt.args.str); got != tt.want {
				t.Errorf("GetValidNameWithFixedPefixNSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}
