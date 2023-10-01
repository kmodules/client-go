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

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/tools/clusterid"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ClusterUID(c client.Reader) (string, error) {
	var ns core.Namespace
	err := c.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &ns)
	if err != nil {
		return "", err
	}
	return string(ns.UID), nil
}

func ClusterMetadata(c client.Reader) (*kmapi.ClusterMetadata, error) {
	var ns core.Namespace
	err := c.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &ns)
	if err != nil {
		return nil, err
	}
	return clusterid.ClusterMetadataForNamespace(&ns)
}

func DetectCAPICluster(kc client.Client) (*kmapi.CAPIClusterInfo, error) {
	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "Cluster",
	})
	err := kc.List(context.TODO(), &list)
	if meta.IsNoMatchError(err) || len(list.Items) == 0 {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else if len(list.Items) > 1 {
		return nil, errors.New("multiple CAPI cluster object found")
	}

	obj := list.Items[0].UnstructuredContent()
	capiProvider, clusterName, ns, err := getCAPIValues(obj)
	if err != nil {
		return nil, err
	}

	return &kmapi.CAPIClusterInfo{
		Provider:    getProviderName(capiProvider),
		Namespace:   ns,
		ClusterName: clusterName,
	}, nil
}

func getCAPIValues(values map[string]any) (string, string, string, error) {
	capiProvider, ok, err := unstructured.NestedString(values, "spec", "infrastructureRef", "kind")
	if err != nil {
		return "", "", "", err
	} else if !ok || capiProvider == "" {
		return "", "", "", nil
	}

	clusterName, ok, err := unstructured.NestedString(values, "metadata", "name")
	if err != nil {
		return "", "", "", err
	} else if !ok || clusterName == "" {
		return "", "", "", nil
	}

	ns, ok, err := unstructured.NestedString(values, "metadata", "namespace")
	if err != nil {
		return "", "", "", err
	} else if !ok || ns == "" {
		return "", "", "", nil
	}

	return capiProvider, clusterName, ns, nil
}

func getProviderName(kind string) string {
	switch kind {
	case "AWSManagedCluster", "AWSManagedControlPlane":
		return "capa"
	case "AzureManagedCluster":
		return "capz"
	case "GCPManagedCluster":
		return "capg"
	}
	return ""
}

func DetectClusterManager(kc client.Client) kmapi.ClusterManager {
	var result kmapi.ClusterManager
	if IsACEManaged(kc) {
		result |= kmapi.ClusterManagerACE
	}
	if IsOpenClusterHub(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerOCMHub
	}
	if IsOpenClusterSpoke(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerOCMSpoke
	}
	if IsOpenClusterMulticlusterControlplane(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerOCMMulticlusterControlplane
	}
	if IsRancherManaged(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerRancher
	}
	if IsOpenShiftManaged(kc.RESTMapper()) {
		result |= kmapi.ClusterManagerOpenShift
	}
	if MustIsVirtualCluster(kc) {
		result |= kmapi.ClusterManagerVirtualCluster
	}
	return result
}

func IsDefault(kc client.Client, cm kmapi.ClusterManager, gvk schema.GroupVersionKind, key types.NamespacedName) (bool, error) {
	if cm.ManagedByRancher() {
		return IsInSystemProject(kc, key.Namespace)
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
