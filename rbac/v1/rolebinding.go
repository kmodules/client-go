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

package v1

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchRoleBinding(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*rbac.RoleBinding) *rbac.RoleBinding) (*rbac.RoleBinding, kutil.VerbType, error) {
	cur, err := c.RbacV1().RoleBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating RoleBinding %s/%s.", meta.Namespace, meta.Name)
		out, err := c.RbacV1().RoleBindings(meta.Namespace).Create(transform(&rbac.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: rbac.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchRoleBinding(c, cur, transform)
}

func PatchRoleBinding(c kubernetes.Interface, cur *rbac.RoleBinding, transform func(*rbac.RoleBinding) *rbac.RoleBinding) (*rbac.RoleBinding, kutil.VerbType, error) {
	return PatchRoleBindingObject(c, cur, transform(cur.DeepCopy()))
}

func PatchRoleBindingObject(c kubernetes.Interface, cur, mod *rbac.RoleBinding) (*rbac.RoleBinding, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, rbac.RoleBinding{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching RoleBinding %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.RbacV1().RoleBindings(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateRoleBinding(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*rbac.RoleBinding) *rbac.RoleBinding) (result *rbac.RoleBinding, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.RbacV1().RoleBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.RbacV1().RoleBindings(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update RoleBinding %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update Role %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func WaitUntillRoleBindingDeleted(kubeClient kubernetes.Interface, meta metav1.ObjectMeta) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.GCTimeout, func() (bool, error) {
		_, err := kubeClient.RbacV1().RoleBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if err != nil && kerr.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
}
