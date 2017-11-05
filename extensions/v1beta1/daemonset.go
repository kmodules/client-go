package v1beta1

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	extensions "k8s.io/api/extensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func CreateOrPatchDaemonSet(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*extensions.DaemonSet) *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	cur, err := c.ExtensionsV1beta1().DaemonSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating DaemonSet %s/%s.", meta.Namespace, meta.Name)
		return c.ExtensionsV1beta1().DaemonSets(meta.Namespace).Create(transform(&extensions.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DaemonSet",
				APIVersion: extensions.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
	} else if err != nil {
		return nil, err
	}
	return PatchDaemonSet(c, cur, transform)
}

func PatchDaemonSet(c kubernetes.Interface, cur *extensions.DaemonSet, transform func(*extensions.DaemonSet) *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	modJson, err := json.Marshal(transform(cur.DeepCopy()))
	if err != nil {
		return nil, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, extensions.DaemonSet{})
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, nil
	}
	glog.V(3).Infof("Patching DaemonSet %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	result, err := c.ExtensionsV1beta1().DaemonSets(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
	if ok, err := kutil.CheckAPIVersion(c, "<= 1.5"); err == nil && ok {
		// https://kubernetes.io/docs/tasks/manage-daemon/update-daemon-set/
		core_util.RestartPods(c, cur.Namespace, cur.Spec.Selector)
	}
	return result, err
}

func TryPatchDaemonSet(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*extensions.DaemonSet) *extensions.DaemonSet) (result *extensions.DaemonSet, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.ExtensionsV1beta1().DaemonSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = PatchDaemonSet(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch DaemonSet %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to patch DaemonSet %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func TryUpdateDaemonSet(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*extensions.DaemonSet) *extensions.DaemonSet) (result *extensions.DaemonSet, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.ExtensionsV1beta1().DaemonSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.ExtensionsV1beta1().DaemonSets(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update DaemonSet %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update DaemonSet %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func WaitUntilDaemonSetReady(kubeClient kubernetes.Interface, meta metav1.ObjectMeta) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		if obj, err := kubeClient.ExtensionsV1beta1().DaemonSets(meta.Namespace).Get(meta.Name, metav1.GetOptions{}); err == nil {
			return obj.Status.DesiredNumberScheduled == obj.Status.NumberReady, nil
		}
		return false, nil
	})
}
