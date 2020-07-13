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
	"reflect"
	"testing"
)

func TestUpsertArgumentList(t *testing.T) {
	cases := []struct {
		name string
		x    []string
		y    []string
		z    []string
		r    []string
	}{
		{
			"t1",
			[]string{},
			[]string{},
			nil,
			[]string{},
		},
		{
			"t2",
			nil,
			nil,
			nil,
			[]string{},
		},
		{
			"t3",
			[]string{"--k1=v1"},
			[]string{"--k1=w1"},
			nil,
			[]string{"--k1=w1"},
		},
		{
			"t4",
			[]string{"--k1=v1", "--k2=v2"},
			[]string{"--k1=w1"},
			nil,
			[]string{"--k1=w1", "--k2=v2"},
		},
		{
			"t5",
			[]string{"--k1=v1", "--k2=v2"},
			[]string{"--k3=w3"},
			nil,
			[]string{"--k1=v1", "--k2=v2", "--k3=w3"},
		},
		{
			"t6",
			[]string{"app", "--k1=v1", "-k2", "v2", "-k3"},
			[]string{"--k1=w1", "--k4=w4", "-k5", "v5"},
			nil,
			[]string{"app", "--k1=w1", "-k2", "v2", "-k3", "--k4=w4", "-k5", "v5"},
		},
		{
			"t7",
			[]string{"--k1=v1", "--k2=v2"},
			[]string{"--k1=w1", "--k2=w2"},
			[]string{"--k2"},
			[]string{"--k1=w1", "--k2=v2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := UpsertArgumentList(c.x, c.y, c.z...)
			if !reflect.DeepEqual(c.r, result) {
				t.Errorf("Failed UpsertArgumentList test for ('%v', '%v'): expected %+v, got %+v", c.x, c.y, c.r, result)
			}
		})
	}
}
