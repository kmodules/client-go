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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectHashForDeployment(t *testing.T) {
	obj := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta2",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "deploy-01",
			Namespace:  "default",
			Generation: 3,
			Annotations: map[string]string{
				"hello": "world",
			},
			Labels: map[string]string{
				"you": "me",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Paused: false,
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 2,
		},
	}

	hash := ObjectHash(obj)

	// generation changed, hash should change
	objNew := obj.DeepCopy()
	objNew.Generation = 2
	hashNew := ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("generation changed, hash should change")
	}

	// annotation changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Annotations["hello"] = "hell"
	hashNew = ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("annotation changed, hash should change")
	}

	// labels changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Labels["you"] = "not-me"
	hashNew = ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("labels changed, hash should change")
	}

	// spec changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Spec.Paused = true
	hashNew = ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("spec changed, hash should change")
	}

	// status changed, hash should not change
	objNew = obj.DeepCopy()
	objNew.Status.ObservedGeneration = 3
	hashNew = ObjectHash(objNew)
	if hash != hashNew {
		t.Errorf("status changed, hash should not changee")
	}
}

func TestObjectHashForConfigmap(t *testing.T) {
	obj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cfg-01",
			Namespace: "default",
		},
		Data: map[string]string{
			"performance": "average",
		},
	}

	hash := ObjectHash(obj)

	// data changed, hash should change
	objNew := obj.DeepCopy()
	objNew.Data["performance"] = "excellent"
	hashNew := ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("data changed, hash should change")
	}
}
