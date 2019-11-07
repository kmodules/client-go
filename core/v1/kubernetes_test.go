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

package v1

import (
	"reflect"
	"testing"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoveOwnerReference(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       "Deployment",
				Name:       "dep-0",
				APIVersion: "v1",
				UID:        "0",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-1",
				APIVersion: "v1",
				UID:        "1",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-2",
				APIVersion: "v1",
				UID:        "2",
			},
		},
	}

	ref := core.ObjectReference{
		Kind:       "Deployment",
		Name:       "dep-3",
		APIVersion: "v1",
		UID:        "3",
	}

	appendedMeta := objectMeta
	appendedMeta.OwnerReferences = append(appendedMeta.OwnerReferences, metav1.OwnerReference{
		UID:        ref.UID,
		APIVersion: ref.APIVersion,
		Name:       ref.Name,
		Kind:       ref.Kind,
	})

	cases := []struct {
		testName string
		objMeta  metav1.ObjectMeta
		owner    core.ObjectReference
		newMeta  metav1.ObjectMeta
	}{
		{
			"Reference is Not Owner of Object",
			objectMeta,
			ref,
			objectMeta,
		},
		{
			"Reference is Owner of Object",
			appendedMeta,
			ref,
			objectMeta,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			RemoveOwnerReference(&c.objMeta, &c.owner)
			if !reflect.DeepEqual(c.objMeta, c.newMeta) {
				t.Errorf("Remove of owner Reference is not successful, expected: %v. But Got: %v", c.newMeta, c.objMeta)
			}
		})
	}
}
