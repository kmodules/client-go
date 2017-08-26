package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	actypes "github.com/appscode/go/types"
	"github.com/appscode/jsonpatch"
	"github.com/appscode/kutil"
	"github.com/cenkalti/backoff"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func EnsureRC(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.ReplicationController) *apiv1.ReplicationController) (*apiv1.ReplicationController, error) {
	return CreateOrPatchRC(c, meta, transform)
}

func CreateOrPatchRC(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.ReplicationController) *apiv1.ReplicationController) (*apiv1.ReplicationController, error) {
	cur, err := c.CoreV1().ReplicationControllers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return c.CoreV1().ReplicationControllers(meta.Namespace).Create(transform(&apiv1.ReplicationController{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchRC(c, cur, transform)
}

func PatchRC(c clientset.Interface, cur *apiv1.ReplicationController, transform func(*apiv1.ReplicationController) *apiv1.ReplicationController) (*apiv1.ReplicationController, error) {
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
	glog.V(5).Infof("Patching ReplicationController %s@%s with %s.", cur.Name, cur.Namespace, string(pb))
	return c.CoreV1().ReplicationControllers(cur.Namespace).Patch(cur.Name, types.JSONPatchType, pb)
}

func TryPatchRC(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.ReplicationController) *apiv1.ReplicationController) (result *apiv1.ReplicationController, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().ReplicationControllers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = PatchRC(c, cur, transform)
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to patch ReplicationController %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to patch ReplicationController %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func TryUpdateRC(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.ReplicationController) *apiv1.ReplicationController) (result *apiv1.ReplicationController, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().ReplicationControllers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = c.CoreV1().ReplicationControllers(cur.Namespace).Update(transform(cur))
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to update ReplicationController %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to update ReplicationController %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func WaitUntilRCReady(c clientset.Interface, meta metav1.ObjectMeta) error {
	return backoff.Retry(func() error {
		if obj, err := c.CoreV1().ReplicationControllers(meta.Namespace).Get(meta.Name, metav1.GetOptions{}); err == nil {
			if actypes.Int32(obj.Spec.Replicas) == obj.Status.ReadyReplicas {
				return nil
			}
		}
		return errors.New("check again")
	}, backoff.NewConstantBackOff(2*time.Second))
}
