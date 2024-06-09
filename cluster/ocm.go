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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsOpenClusterHub(mapper meta.RESTMapper) bool {
	if _, err := mapper.RESTMappings(schema.GroupKind{
		Group: "cluster.open-cluster-management.io",
		Kind:  "ManagedCluster",
	}); err == nil {
		return true
	}
	return false
}

func IsOpenClusterSpoke(mapper meta.RESTMapper) bool {
	if _, err := mapper.RESTMappings(schema.GroupKind{
		Group: "work.open-cluster-management.io",
		Kind:  "AppliedManifestWork",
	}); err == nil {
		return true
	}
	return false
}

func IsOpenClusterMulticlusterControlplane(kc client.Client, mapper meta.RESTMapper) bool {
	var missingDeployment bool
	var deployment appsv1.Deployment
	if err := kc.Get(context.Background(), types.NamespacedName{Name: "multicluster-controlplane", Namespace: "multicluster-controlplane"}, &deployment); err != nil {
		missingDeployment = true
	}
	return IsOpenClusterSpoke(mapper) && !missingDeployment
}
