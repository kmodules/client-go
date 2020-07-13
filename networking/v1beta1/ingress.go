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

package v1beta1

import (
	"context"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchIngress(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*networking.Ingress) *networking.Ingress, opts metav1.PatchOptions) (*networking.Ingress, kutil.VerbType, error) {
	cur, err := c.NetworkingV1beta1().Ingresses(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Ingress %s/%s.", meta.Namespace, meta.Name)
		out, err := c.NetworkingV1beta1().Ingresses(meta.Namespace).Create(ctx, transform(&networking.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: networking.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchIngress(ctx, c, cur, transform, opts)
}

func PatchIngress(ctx context.Context, c kubernetes.Interface, cur *networking.Ingress, transform func(*networking.Ingress) *networking.Ingress, opts metav1.PatchOptions) (*networking.Ingress, kutil.VerbType, error) {
	return PatchIngressObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchIngressObject(ctx context.Context, c kubernetes.Interface, cur, mod *networking.Ingress, opts metav1.PatchOptions) (*networking.Ingress, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, networking.Ingress{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching Ingress %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.NetworkingV1beta1().Ingresses(cur.Namespace).Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateIngress(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*networking.Ingress) *networking.Ingress, opts metav1.UpdateOptions) (result *networking.Ingress, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.NetworkingV1beta1().Ingresses(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.NetworkingV1beta1().Ingresses(cur.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Ingress %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update Ingress %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
