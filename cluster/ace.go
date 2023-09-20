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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsACEManaged(kc client.Client) bool {
	var list unstructured.UnstructuredList
	list.SetAPIVersion("apps/v1")
	list.SetKind("Deployment")
	err := kc.List(context.TODO(), &list, client.InNamespace("kubeops"), client.MatchingLabels{
		"app.kubernetes.io/name":     "kube-ui-server",
		"app.kubernetes.io/instance": "kube-ui-server",
	})
	if err != nil {
		klog.Errorln(err)
	}
	return len(list.Items) > 0
}
