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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsOpenClusterHub(kc client.Client) bool {
	var list unstructured.UnstructuredList
	list.SetAPIVersion("apps/v1")
	list.SetKind("Deployment")
	err := kc.List(context.TODO(), &list, client.InNamespace("open-cluster-management-hub"), client.MatchingLabels{
		"app": "clustermanager-controller",
	})
	if err != nil {
		klog.Errorln(err)
	}
	return len(list.Items) > 1
}

func IsOpenClusterSpoke(kc client.Client) bool {
	var list unstructured.UnstructuredList
	list.SetAPIVersion("apps/v1")
	list.SetKind("Deployment")
	err := kc.List(context.TODO(), &list, client.InNamespace("open-cluster-management-agent"), client.MatchingLabels{
		"app": "klusterlet-registration-agent",
	})
	if err != nil {
		klog.Errorln(err)
	}
	return len(list.Items) > 0
}

func IsOpenClusterMulticlusterControlplane(kc client.Client) bool {
	var missingDeployment bool
	if _, err := kc.RESTMapper().RESTMappings(schema.GroupKind{
		Group: "apps",
		Kind:  "Deployment",
	}); meta.IsNoMatchError(err) {
		missingDeployment = true
	}
	return IsOpenClusterHub(kc) && missingDeployment
}
