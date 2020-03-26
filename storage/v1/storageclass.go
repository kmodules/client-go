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
	"context"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	storage "k8s.io/api/storage/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchStorageClass(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*storage.StorageClass) *storage.StorageClass) (*storage.StorageClass, kutil.VerbType, error) {
	cur, err := c.StorageV1().StorageClasses().Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating StorageClass %s.", meta.Name)
		out, err := c.StorageV1().StorageClasses().Create(ctx, transform(&storage.StorageClass{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StorageClass",
				APIVersion: storage.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchStorageClass(ctx, c, cur, transform)
}

func PatchStorageClass(ctx context.Context, c kubernetes.Interface, cur *storage.StorageClass, transform func(*storage.StorageClass) *storage.StorageClass) (*storage.StorageClass, kutil.VerbType, error) {
	return PatchStorageClassObject(ctx, c, cur, transform(cur.DeepCopy()))
}

func PatchStorageClassObject(ctx context.Context, c kubernetes.Interface, cur, mod *storage.StorageClass) (*storage.StorageClass, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, storage.StorageClass{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching StorageClass %s with %s.", cur.Name, string(patch))
	out, err := c.StorageV1().StorageClasses().Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return out, kutil.VerbPatched, err
}

func TryUpdateStorageClass(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*storage.StorageClass) *storage.StorageClass) (result *storage.StorageClass, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.StorageV1().StorageClasses().Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.StorageV1().StorageClasses().Update(ctx, transform(cur.DeepCopy()), metav1.UpdateOptions{})
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update StorageClass %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update StorageClass %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}
