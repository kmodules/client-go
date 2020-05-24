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

package v1beta1

import (
	"context"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchCustomResourceDefinition(
	ctx context.Context,
	c cs.Interface,
	name string,
	transform func(in *api.CustomResourceDefinition) *api.CustomResourceDefinition,
	opts metav1.PatchOptions,
) (*api.CustomResourceDefinition, kutil.VerbType, error) {
	cur, err := c.ApiextensionsV1beta1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating CustomResourceDefinition %s.", name)
		out, err := c.ApiextensionsV1beta1().CustomResourceDefinitions().Create(ctx, transform(&api.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchCustomResourceDefinition(ctx, c, cur, transform, opts)
}

func PatchCustomResourceDefinition(
	ctx context.Context,
	c cs.Interface,
	cur *api.CustomResourceDefinition,
	transform func(*api.CustomResourceDefinition) *api.CustomResourceDefinition,
	opts metav1.PatchOptions,
) (*api.CustomResourceDefinition, kutil.VerbType, error) {
	return PatchCustomResourceDefinitionObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchCustomResourceDefinitionObject(
	ctx context.Context,
	c cs.Interface,
	cur, mod *api.CustomResourceDefinition,
	opts metav1.PatchOptions,
) (*api.CustomResourceDefinition, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, api.CustomResourceDefinition{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching CustomResourceDefinition %s with %s.", cur.Name, string(patch))
	out, err := c.ApiextensionsV1beta1().CustomResourceDefinitions().Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateCustomResourceDefinition(
	ctx context.Context,
	c cs.Interface,
	name string,
	transform func(*api.CustomResourceDefinition) *api.CustomResourceDefinition,
	opts metav1.UpdateOptions,
) (result *api.CustomResourceDefinition, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.ApiextensionsV1beta1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.ApiextensionsV1beta1().CustomResourceDefinitions().Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update CustomResourceDefinition %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update CustomResourceDefinition %s after %d attempts due to %v", name, attempt, err)
	}
	return
}
