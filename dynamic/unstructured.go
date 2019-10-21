package dynamic

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatch(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	meta metav1.ObjectMeta,
	transform func(*unstructured.Unstructured) *unstructured.Unstructured,
) (*unstructured.Unstructured, kutil.VerbType, error) {
	var ri dynamic.ResourceInterface
	if meta.Namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(meta.Namespace)
	}

	cur, err := ri.Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating %s %s/%s.", gvr.String(), meta.Namespace, meta.Name)
		u := &unstructured.Unstructured{}
		u.SetName(meta.Name)
		u.SetNamespace(meta.Namespace)
		out, err := ri.Create(transform(u), metav1.CreateOptions{})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return Patch(c, gvr, cur, transform)
}

func Patch(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	cur *unstructured.Unstructured,
	transform func(*unstructured.Unstructured) *unstructured.Unstructured,
) (*unstructured.Unstructured, kutil.VerbType, error) {
	return PatchObject(c, gvr, cur, transform(cur.DeepCopy()))
}

func PatchObject(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	cur, mod *unstructured.Unstructured,
) (*unstructured.Unstructured, kutil.VerbType, error) {
	var ri dynamic.ResourceInterface
	if cur.GetNamespace() == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(cur.GetNamespace())
	}

	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(curJson, modJson, curJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching %s %s/%s with %s.", gvr.String(), cur.GetNamespace(), cur.GetName(), string(patch))
	out, err := ri.Patch(cur.GetName(), types.MergePatchType, patch, metav1.PatchOptions{})
	return out, kutil.VerbPatched, err
}

func TryUpdate(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	meta metav1.ObjectMeta,
	transform func(*unstructured.Unstructured) *unstructured.Unstructured,
) (result *unstructured.Unstructured, err error) {
	var ri dynamic.ResourceInterface
	if meta.Namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(meta.Namespace)
	}

	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := ri.Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = ri.Update(transform(cur.DeepCopy()), metav1.UpdateOptions{})
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update %s %s/%s due to %v.", attempt, gvr.String(), cur.GetNamespace(), cur.GetName(), e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update %s %s/%s after %d attempts due to %v", gvr.String(), meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateStatus(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	in *unstructured.Unstructured,
	transform func(*unstructured.Unstructured) *unstructured.Unstructured,
) (result *unstructured.Unstructured, err error) {
	var ri dynamic.ResourceInterface
	if in.GetNamespace() == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(in.GetNamespace())
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = ri.UpdateStatus(transform(cur), metav1.UpdateOptions{})
		if kerr.IsConflict(e2) {
			latest, e3 := ri.Get(in.GetName(), metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of %s %s/%s after %d attempts due to %v", gvr.String(), in.GetNamespace(), in.GetName(), attempt, err)
	}
	return
}
