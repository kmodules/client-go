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

package manager

import (
	"context"

	kmapi "kmodules.xyz/client-go/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DetectClusterManager(kc client.Client) kmapi.ClusterManager {
	var result kmapi.ClusterManager
	if IsACEManaged(kc) {
		result |= kmapi.ClusterManagerACE
	}
	if IsOpenClusterManaged(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerOCM
	}
	if IsRancherManaged(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerRancher
	}
	return result
}

func IsDefault(kc client.Client, cm kmapi.ClusterManager, gvk schema.GroupVersionKind, key types.NamespacedName) (bool, error) {
	if cm.ManagedByRancher() {
		return IsRancherSystemResource(kc, key)
	}
	return IsSingletonResource(kc, gvk, key)
}

func IsSingletonResource(kc client.Client, gvk schema.GroupVersionKind, key types.NamespacedName) (bool, error) {
	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(gvk)
	err := kc.List(context.TODO(), &list, client.InNamespace(key.Namespace))
	if err != nil {
		return false, err
	}
	return len(list.Items) == 1, nil
}
