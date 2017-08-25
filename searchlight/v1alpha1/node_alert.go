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

func EnsureNodeAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(alert *aci.NodeAlert) *aci.NodeAlert) (*aci.NodeAlert, error) {
	return CreateOrPatchNodeAlert(c, meta, transform)
}

func CreateOrPatchNodeAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(alert *aci.NodeAlert) *aci.NodeAlert) (*aci.NodeAlert, error) {
	cur, err := c.NodeAlerts(meta.Namespace).Get(meta.Name)
	if kerr.IsNotFound(err) {
		return c.NodeAlerts(meta.Namespace).Create(transform(&aci.NodeAlert{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchNodeAlert(c, cur, transform)
}

func PatchNodeAlert(c tcs.ExtensionInterface, cur *aci.NodeAlert, transform func(*aci.NodeAlert) *aci.NodeAlert) (*aci.NodeAlert, error) {
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
	glog.V(5).Infof("Patching NodeAlert %s@%s with %s.", cur.Name, cur.Namespace, string(pb))
	result, err := c.NodeAlerts(cur.Namespace).Patch(cur.Name, types.JSONPatchType, pb)
	return result, err
}

func TryPatchNodeAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.NodeAlert) *aci.NodeAlert) (result *aci.NodeAlert, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.NodeAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = PatchNodeAlert(c, cur, transform)
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to patch NodeAlert %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to patch NodeAlert %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func TryUpdateNodeAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.NodeAlert) *aci.NodeAlert) (result *aci.NodeAlert, err error) {
	attempt := 0
	err = wait.Poll(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.NodeAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(e2) {
			return true, e2
		} else if e2 == nil {
			result, e2 = c.NodeAlerts(cur.Namespace).Update(transform(cur))
			return e2 == nil, e2
		}
		glog.Errorf("Attempt %d failed to update NodeAlert %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, e2
	})

	if err != nil {
		err = fmt.Errorf("Failed to update NodeAlert %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}
