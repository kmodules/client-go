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

	"kmodules.xyz/client-go/discovery"
	v1 "kmodules.xyz/client-go/policy/v1"
	"kmodules.xyz/client-go/policy/v1beta1"

	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchPodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policyv1.PodDisruptionBudget) *policyv1.PodDisruptionBudget, opts metav1.PatchOptions) (*policyv1.PodDisruptionBudget, kutil.VerbType, error) {
	if ok, err := discovery.CheckAPIVersion(c.Discovery(), ">= 1.21"); err == nil && ok {
		return v1.CreateOrPatchPodDisruptionBudget(ctx, c, meta, transform, opts)
	}

	p, vt, err := v1beta1.CreateOrPatchPodDisruptionBudget(
		ctx,
		c,
		meta,
		func(in *policyv1beta1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
			out := convert_v1_to_v1beta1(transform(convert_v1beta1_to_v1(in)))
			out.Status = in.Status
			return out
		},
		opts,
	)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return convert_v1beta1_to_v1(p), vt, nil
}

func DeletePodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta types.NamespacedName) error {
	if ok, err := discovery.CheckAPIVersion(c.Discovery(), ">= 1.21"); err == nil && ok {
		return c.PolicyV1().PodDisruptionBudgets(meta.Namespace).Delete(ctx, meta.Name, metav1.DeleteOptions{})
	}
	return c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Delete(ctx, meta.Name, metav1.DeleteOptions{})
}

func convert_v1beta1_to_v1(in *policyv1beta1.PodDisruptionBudget) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       in.Kind,
			APIVersion: policyv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: in.ObjectMeta,
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable:   in.Spec.MinAvailable,
			Selector:       in.Spec.Selector,
			MaxUnavailable: in.Spec.MaxUnavailable,
		},
		// Status:     policyv1.PodDisruptionBudgetStatus{},
	}
}

func convert_v1_to_v1beta1(in *policyv1.PodDisruptionBudget) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       in.Kind,
			APIVersion: policyv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: in.ObjectMeta,
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable:   in.Spec.MinAvailable,
			Selector:       in.Spec.Selector,
			MaxUnavailable: in.Spec.MaxUnavailable,
		},
		// Status:     policyv1beta1.PodDisruptionBudgetStatus{},
	}
}
