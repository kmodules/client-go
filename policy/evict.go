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

package policy

import (
	"context"

	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func EvictPod(ctx context.Context, c kubernetes.Interface, meta types.NamespacedName, opts *metav1.DeleteOptions) error {
	detectVersion(c.Discovery())
	if useV1 {
		return c.CoreV1().Pods(meta.Namespace).EvictV1(ctx, &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meta.Name,
				Namespace: meta.Namespace,
			},
			DeleteOptions: opts,
		})
	}
	return c.CoreV1().Pods(meta.Namespace).EvictV1beta1(ctx, &policyv1beta1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
		},
		DeleteOptions: opts,
	})
}
