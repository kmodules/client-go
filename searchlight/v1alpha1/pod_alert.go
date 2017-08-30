package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/jsonpatch"
	"github.com/appscode/kutil"
	aci "github.com/appscode/searchlight/api"
	tcs "github.com/appscode/searchlight/client/clientset"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func EnsurePodAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(alert *aci.PodAlert) *aci.PodAlert) (*aci.PodAlert, error) {
	return CreateOrPatchPodAlert(c, meta, transform)
}

func CreateOrPatchPodAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(alert *aci.PodAlert) *aci.PodAlert) (*aci.PodAlert, error) {
	cur, err := c.PodAlerts(meta.Namespace).Get(meta.Name)
	if kerr.IsNotFound(err) {
		return c.PodAlerts(meta.Namespace).Create(transform(&aci.PodAlert{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchPodAlert(c, cur, transform)
}

func PatchPodAlert(c tcs.ExtensionInterface, cur *aci.PodAlert, transform func(*aci.PodAlert) *aci.PodAlert) (*aci.PodAlert, error) {
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
	glog.V(5).Infof("Patching PodAlert %s@%s with %s.", cur.Name, cur.Namespace, string(pb))
	result, err := c.PodAlerts(cur.Namespace).Patch(cur.Name, types.JSONPatchType, pb)
	return result, err
}

func TryPatchPodAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.PodAlert) *aci.PodAlert) (result *aci.PodAlert, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.PodAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = PatchPodAlert(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch PodAlert %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("Failed to patch PodAlert %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func TryUpdatePodAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.PodAlert) *aci.PodAlert) (result *aci.PodAlert, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.PodAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.PodAlerts(cur.Namespace).Update(transform(cur))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update PodAlert %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("Failed to update PodAlert %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}
