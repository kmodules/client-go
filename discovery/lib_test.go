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

package discovery

import (
	"testing"

	"gomodules.xyz/version"
)

func TestDefaultSupportedVersion(t *testing.T) {
	cases := []struct {
		version     string
		multiMaster bool
		err         bool
	}{
		{"1.8.5", false, true},
		{"1.9.0", false, true},
		{"1.9.0", true, true},
		{"1.9.8", true, true},
		{"1.10.0", false, true},
		{"1.10.0", true, true},
		{"1.10.2", true, true},
		{"1.11.0", false, false},
		{"1.11.0", true, false},
		{"1.16.0", false, true},
		{"1.16.0", true, true},
		{"1.16.3", false, false},
		{"1.16.3", true, false},
	}

	for _, tc := range cases {
		v, err := version.NewVersion(tc.version)
		if err != nil {
			t.Fatalf("failed parse version for input %s: %s", tc.version, err)
		}

		err = checkVersion(
			v,
			tc.multiMaster,
			DefaultConstraint,
			DefaultBlackListedVersions,
			DefaultBlackListedMultiMasterVersions)
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.version)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.version, err)
		}
	}
}
