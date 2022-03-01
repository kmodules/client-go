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

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/mergepatch"
)

func newObj() apps.Deployment {
	return apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32P(3),
			Template: core.PodTemplateSpec{
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "foo",
							Image: "foo/bar:latest",
						},
						{
							Name:  "bar",
							Image: "foo/bar:latest",
						},
					},
					Hostname: "foo-bar",
				},
			},
		},
	}
}

func getPreconditionFuncs() []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
		mergepatch.RequireKeyUnchanged("status"),
		// below methods are added in kutil/meta/patch.go
		RequireChainKeyUnchanged("spec.replicas"),
		RequireChainKeyUnchanged("spec.template.spec.containers.image"), // here container is array, yet works fine
	}
	return preconditions
}

func TestCreateStrategicPatch_Conditions(t *testing.T) {
	obj, validMod, badMod, badArrayMod := newObj(), newObj(), newObj(), newObj()
	validMod.Spec.Template.Spec.Hostname = "NewHostName"
	badMod.Spec.Replicas = pointer.Int32P(2)
	badArrayMod.Spec.Template.Spec.Containers[0].Image = "newImage"

	preconditions := getPreconditionFuncs()

	cases := []struct {
		name   string
		x      apps.Deployment
		y      apps.Deployment
		cond   []mergepatch.PreconditionFunc
		result bool
	}{
		{"bad modification without condition", obj, badMod, nil, true}, //	// without preconditions
		{"valid modification with condition", obj, validMod, preconditions, true},
		{"bad modification with condition", obj, badMod, preconditions, false},
		{"bad array modification with condition", obj, badArrayMod, preconditions, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := CreateStrategicPatch(&c.x, &c.y, c.cond...)
			if c.result == true {
				if err != nil {
					t.Errorf("Modifications should be passed. error: %v", err)
				}
			} else if c.result == false {
				if err == nil || !mergepatch.IsPreconditionFailed(err) {
					t.Errorf("Modifications should be failed. error: %v", err)
				}
			}
		})
	}
}

func TestCreateJSONMergePatch_Conditions(t *testing.T) {
	obj, validMod, badMod, badArrayMod := newObj(), newObj(), newObj(), newObj()
	validMod.Spec.Template.Spec.Hostname = "NewHostName"
	badMod.Spec.Replicas = pointer.Int32P(2)
	badArrayMod.Spec.Template.Spec.Containers[0].Image = "newImage"

	preconditions := getPreconditionFuncs()

	cases := []struct {
		name   string
		x      apps.Deployment
		y      apps.Deployment
		cond   []mergepatch.PreconditionFunc
		result bool
	}{
		{"bad modification without condition", obj, badMod, nil, true}, //	// without preconditions
		{"valid modification with condition", obj, validMod, preconditions, true},
		{"bad modification with condition", obj, badMod, preconditions, false},
		{"bad array modification with condition", obj, badArrayMod, preconditions, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := CreateJSONMergePatch(&c.x, &c.y, c.cond...)
			if c.result == true {
				if err != nil {
					t.Errorf("Modifications should be passed. error: %v", err)
				}
			} else if c.result == false {
				if err == nil || !mergepatch.IsPreconditionFailed(err) {
					t.Errorf("Modifications should be failed. error: %v", err)
				}
			}
		})
	}
}
