package v1

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func CreateOrPatchSecret(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*core.Secret) *core.Secret) (*core.Secret, bool, error) {
	cur, err := c.CoreV1().Secrets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Secret %s/%s.", meta.Namespace, meta.Name)
		out, err := c.CoreV1().Secrets(meta.Namespace).Create(transform(&core.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: core.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, true, err
	} else if err != nil {
		return nil, false, err
	}
	return PatchSecret(c, cur, transform)
}

func PatchSecret(c kubernetes.Interface, cur *core.Secret, transform func(*core.Secret) *core.Secret) (*core.Secret, bool, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, false, err
	}

	modJson, err := json.Marshal(transform(cur.DeepCopy()))
	if err != nil {
		return nil, false, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, core.Secret{})
	if err != nil {
		return nil, false, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, false, nil
	}
	glog.V(3).Infof("Patching Secret %s/%s", cur.Namespace, cur.Name)
	out, err := c.CoreV1().Secrets(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
	return out, true, err
}

func TryPatchSecret(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*core.Secret) *core.Secret) (result *core.Secret, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().Secrets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, _, e2 = PatchSecret(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch Secret %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to patch Secret %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func TryUpdateSecret(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*core.Secret) *core.Secret) (result *core.Secret, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().Secrets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.CoreV1().Secrets(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Secret %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Secret %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
