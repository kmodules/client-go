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

package apiextensions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v1 "kmodules.xyz/client-go/apiextensions/v1"
	"kmodules.xyz/client-go/apiextensions/v1beta1"
	discovery_util "kmodules.xyz/client-go/discovery"
	meta_util "kmodules.xyz/client-go/meta"

	"github.com/pkg/errors"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

func RegisterCRDs(client crd_cs.Interface, crds []*CustomResourceDefinition) error {
	major, minor, _, _, _, err := discovery_util.GetVersionInfo(client.Discovery())
	if err != nil {
		return err
	}

	for _, crd := range crds {
		// Use crd v1 for k8s >= 1.16, if available
		// ref: https://github.com/kubernetes/kubernetes/issues/91395
		if major == 1 && minor >= 16 && crd.V1 != nil {
			_, _, err := v1.CreateOrPatchCustomResourceDefinition(
				context.TODO(),
				client,
				crd.V1.Name,
				func(in *crdv1.CustomResourceDefinition) *crdv1.CustomResourceDefinition {
					in.Labels = meta_util.MergeKeys(in.Labels, crd.V1.Labels)
					in.Annotations = meta_util.MergeKeys(in.Annotations, crd.V1.Annotations)

					in.Spec = crd.V1.Spec
					return crd.V1
				},
				metav1.PatchOptions{},
			)
			if err != nil {
				return err
			}
		} else {
			if crd.V1beta1 == nil {
				return fmt.Errorf("missing crd v1beta1 definition")
			}

			if major == 1 && minor <= 11 {
				// CRD schema must only have "properties", "required" or "description" at the root if the status subresource is enabled
				// xref: https://github.com/stashed/stash/issues/1007#issuecomment-570888875
				crd.V1beta1.Spec.Validation.OpenAPIV3Schema.Type = ""
			}

			_, _, err := v1beta1.CreateOrPatchCustomResourceDefinition(
				context.TODO(),
				client,
				crd.V1beta1.Name,
				func(in *crdv1beta1.CustomResourceDefinition) *crdv1beta1.CustomResourceDefinition {
					in.Labels = meta_util.MergeKeys(in.Labels, crd.V1beta1.Labels)
					in.Annotations = meta_util.MergeKeys(in.Annotations, crd.V1beta1.Annotations)

					in.Spec = crd.V1beta1.Spec
					return crd.V1beta1
				},
				metav1.PatchOptions{},
			)
			if err != nil {
				return err
			}
		}
	}
	return WaitForCRDReady(client.ApiextensionsV1beta1().RESTClient(), crds)
}

func WaitForCRDReady(restClient rest.Interface, crds []*CustomResourceDefinition) error {
	err := wait.Poll(3*time.Second, 5*time.Minute, func() (bool, error) {
		for _, crd := range crds {
			var gvr schema.GroupVersionResource
			if crd.V1 != nil {
				gvr = schema.GroupVersionResource{
					Group:    crd.V1.Spec.Group,
					Version:  crd.V1.Spec.Versions[0].Name,
					Resource: crd.V1.Spec.Names.Plural,
				}
			} else if crd.V1beta1 != nil {
				gvr = schema.GroupVersionResource{
					Group:    crd.V1beta1.Spec.Group,
					Version:  crd.V1beta1.Spec.Versions[0].Name,
					Resource: crd.V1beta1.Spec.Names.Plural,
				}
			}

			res := restClient.Get().AbsPath("apis", gvr.Group, gvr.Version, gvr.Resource).Do(context.TODO())
			err := res.Error()
			if err != nil {
				// RESTClient returns *apierrors.StatusError for any status codes < 200 or > 206
				// and http.Client.Do errors are returned directly.
				if se, ok := err.(*kerr.StatusError); ok {
					if se.Status().Code == http.StatusNotFound {
						return false, nil
					}
				}
				return false, err
			}

			var statusCode int
			res.StatusCode(&statusCode)
			if statusCode != http.StatusOK {
				return false, errors.Errorf("invalid status code: %d", statusCode)
			}
		}

		return true, nil
	})

	return errors.Wrap(err, "timed out waiting for CRD")
}
