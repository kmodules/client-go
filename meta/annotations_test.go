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
package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMap(t *testing.T) {
	in := map[string]string{
		"k1": `{"o1": "v1"}`,
	}

	actual, _ := GetMap(in, "k1")
	assert.Equal(t, map[string]string{"o1": "v1"}, actual)
}

func TestGetFloat(t *testing.T) {
	in := map[string]string{
		"k1": "17.33",
	}
	actual, _ := GetFloat(in, "k1")
	assert.Equal(t, 17.33, actual)
}

func TestGetDuration(t *testing.T) {
	in := map[string]string{
		"k1": "30s",
	}
	actual, _ := GetDuration(in, "k1")
	assert.Equal(t, time.Second*30, actual)
}
