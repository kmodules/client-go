package v1alpha1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/kutil"
	aci "github.com/appscode/searchlight/api"
	tcs "github.com/appscode/searchlight/client/clientset"
	"github.com/golang/glog"
	"github.com/mattbaird/jsonpatch"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func CreateOrPatchClusterAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(alert *aci.ClusterAlert) *aci.ClusterAlert) (*aci.ClusterAlert, error) {
	cur, err := c.ClusterAlerts(meta.Namespace).Get(meta.Name)
	if kerr.IsNotFound(err) {
		return c.ClusterAlerts(meta.Namespace).Create(transform(&aci.ClusterAlert{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchClusterAlert(c, cur, transform)
}

func PatchClusterAlert(c tcs.ExtensionInterface, cur *aci.ClusterAlert, transform func(*aci.ClusterAlert) *aci.ClusterAlert) (*aci.ClusterAlert, error) {
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
	glog.V(5).Infof("Patching ClusterAlert %s@%s with %s.", cur.Name, cur.Namespace, string(pb))
	result, err := c.ClusterAlerts(cur.Namespace).Patch(cur.Name, types.JSONPatchType, pb)
	return result, err
}

func TryPatchClusterAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.ClusterAlert) *aci.ClusterAlert) (*aci.ClusterAlert, error) {
	attempt := 0
	for ; attempt < kutil.MaxAttempts; attempt = attempt + 1 {
		cur, err := c.ClusterAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(err) {
			return cur, err
		} else if err == nil {
			return PatchClusterAlert(c, cur, transform)
		}
		glog.Errorf("Attempt %d failed to patch ClusterAlert %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(kutil.RetryInterval)
	}
	return nil, fmt.Errorf("Failed to patch ClusterAlert %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}

func TryUpdateClusterAlert(c tcs.ExtensionInterface, meta metav1.ObjectMeta, transform func(*aci.ClusterAlert) *aci.ClusterAlert) (*aci.ClusterAlert, error) {
	attempt := 0
	for ; attempt < kutil.MaxAttempts; attempt = attempt + 1 {
		cur, err := c.ClusterAlerts(meta.Namespace).Get(meta.Name)
		if kerr.IsNotFound(err) {
			return cur, err
		} else if err == nil {
			oJson, err := json.Marshal(cur)
			if err != nil {
				return nil, err
			}
			modified := transform(cur)
			mJson, err := json.Marshal(modified)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(oJson, mJson) {
				return cur, err
			}

			result, err := c.ClusterAlerts(cur.Namespace).Update(transform(cur))
			return result, err
		}
		glog.Errorf("Attempt %d failed to update ClusterAlert %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(kutil.RetryInterval)
	}
	return nil, fmt.Errorf("Failed to update ClusterAlert %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}
