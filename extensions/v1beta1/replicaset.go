package v1beta1

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	. "github.com/appscode/go/types"
	"github.com/appscode/jsonpatch"
	"github.com/appscode/kutil"
	"github.com/cenkalti/backoff"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func EnsureReplicaSet(c clientset.Interface, meta metav1.ObjectMeta, transform func(*extensions.ReplicaSet) *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	return CreateOrPatchReplicaSet(c, meta, transform)
}

func CreateOrPatchReplicaSet(c clientset.Interface, meta metav1.ObjectMeta, transform func(*extensions.ReplicaSet) *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	cur, err := c.ExtensionsV1beta1().ReplicaSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return c.ExtensionsV1beta1().ReplicaSets(meta.Namespace).Create(transform(&extensions.ReplicaSet{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchReplicaSet(c, cur, transform)
}

func PatchReplicaSet(c clientset.Interface, cur *extensions.ReplicaSet, transform func(*extensions.ReplicaSet) *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	modJson, err := json.Marshal(transform(cur))
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreatePatch(curJson, modJson)
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 {
		return cur, nil
	}
	pb, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("Patching ReplicaSet %s@%s with %s.", cur.Name, cur.Namespace, string(pb))
	return c.ExtensionsV1beta1().ReplicaSets(cur.Namespace).Patch(cur.Name, types.JSONPatchType, pb)
}

func TryPatchReplicaSet(c clientset.Interface, meta metav1.ObjectMeta, transform func(*extensions.ReplicaSet) *extensions.ReplicaSet) (result *extensions.ReplicaSet, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.ExtensionsV1beta1().ReplicaSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = PatchReplicaSet(c, cur, transform)
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to patch ReplicaSet %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to patch ReplicaSet %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func TryUpdateReplicaSet(c clientset.Interface, meta metav1.ObjectMeta, transform func(*extensions.ReplicaSet) *extensions.ReplicaSet) (result *extensions.ReplicaSet, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.ExtensionsV1beta1().ReplicaSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = c.ExtensionsV1beta1().ReplicaSets(cur.Namespace).Update(transform(cur))
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to update ReplicaSet %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to update ReplicaSet %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func WaitUntilReplicaSetReady(c clientset.Interface, meta metav1.ObjectMeta) error {
	return backoff.Retry(func() error {
		if obj, err := c.ExtensionsV1beta1().ReplicaSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{}); err == nil {
			if Int32(obj.Spec.Replicas) == obj.Status.ReadyReplicas {
				return nil
			}
		}
		return errors.New("check again")
	}, backoff.NewConstantBackOff(2*time.Second))
}
