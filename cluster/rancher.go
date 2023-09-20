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

package cluster

import (
	"context"
	"errors"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const LabelKeyRancherProjectId = "field.cattle.io/projectId"

func IsRancherManaged(mapper meta.RESTMapper) bool {
	if _, err := mapper.RESTMappings(schema.GroupKind{
		Group: "management.cattle.io",
		Kind:  "Cluster",
	}); err == nil {
		return true
	}
	return false
}

func IsRancherSystemResource(kc client.Client, key types.NamespacedName) (bool, error) {
	if !IsRancherManaged(kc.RESTMapper()) {
		return false, errors.New("not a Rancher managed cluster")
	}

	if key.Namespace == metav1.NamespaceSystem {
		return true, nil
	}

	var ns core.Namespace
	err := kc.Get(context.TODO(), client.ObjectKey{Name: key.Namespace}, &ns)
	if err != nil {
		return false, err
	}
	projectId, exists := ns.Labels[LabelKeyRancherProjectId]
	if !exists {
		return false, nil
	}

	var sysNS core.Namespace
	err = kc.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &sysNS)
	if err != nil {
		return false, err
	}

	sysProjectId, exists := ns.Labels[LabelKeyRancherProjectId]
	if !exists {
		return false, nil
	}
	return projectId == sysProjectId, nil
}
