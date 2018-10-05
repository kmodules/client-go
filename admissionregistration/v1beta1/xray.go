package v1beta1

import (
	"github.com/appscode/kutil"
	"github.com/appscode/kutil/discovery"
	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/api/admissionregistration/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ValidatingWebhookXray struct {
	config    *rest.Config
	obj       runtime.Object
	op        v1beta1.OperationType
	transform func(_ runtime.Object)
}

func NewCreateValidatingWebhookXray(config *rest.Config, obj runtime.Object) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:    config,
		obj:       obj,
		op:        v1beta1.Create,
		transform: nil,
	}
}

func NewUpdateValidatingWebhookXray(config *rest.Config, obj runtime.Object, transform func(_ runtime.Object)) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:    config,
		obj:       obj,
		op:        v1beta1.Update,
		transform: transform,
	}
}

func NewDeleteValidatingWebhookXray(config *rest.Config, obj runtime.Object, transform func(_ runtime.Object)) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:    config,
		obj:       obj,
		op:        v1beta1.Delete,
		transform: transform,
	}
}

var ErrMissingKind = errors.New("test object missing kind")
var ErrMissingVersion = errors.New("test object missing version")
var ErrInactiveWebhook = errors.New("webhook is inactive")

func (d ValidatingWebhookXray) IsActive() (bool, error) {
	kc, err := kubernetes.NewForConfig(d.config)
	if err != nil {
		return false, err
	}

	dc, err := dynamic.NewForConfig(d.config)
	if err != nil {
		return false, err
	}

	gvk := d.obj.GetObjectKind().GroupVersionKind()
	if gvk.Version == "" {
		return false, ErrMissingVersion
	}
	if gvk.Kind == "" {
		return false, ErrMissingKind
	}
	glog.Infof("testing ValidatingWebhook using an object with GVK = %s", gvk.String())

	gvr, err := discovery.ResourceForGVK(kc.Discovery(), gvk)
	if err != nil {
		return false, err
	}
	glog.Infof("testing ValidatingWebhook using an object with GVR = %s", gvr.String())

	accessor, err := meta.Accessor(d.obj)
	if err != nil {
		return false, err
	}

	var ri dynamic.ResourceInterface
	if accessor.GetNamespace() != "" {
		ri = dc.Resource(gvr).Namespace(accessor.GetNamespace())
	} else {
		ri = dc.Resource(gvr)
	}

	objJson, err := json.Marshal(d.obj)
	if err != nil {
		return false, err
	}

	u := unstructured.Unstructured{}
	_, _, err = unstructured.UnstructuredJSONScheme.Decode(objJson, nil, &u)
	if err != nil {
		return false, err
	}

	if d.op == v1beta1.Create {
		_, err := ri.Create(&u)
		if kerr.IsForbidden(err) {
			glog.Infof("failed to create invalid test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		err = ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return false, ErrInactiveWebhook
	} else if d.op == v1beta1.Update {
		_, err := ri.Create(&u)
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else {
			return false, err
		}

		mod := d.obj.DeepCopyObject()
		d.transform(mod)
		modJson, err := json.Marshal(mod)
		if err != nil {
			return false, err
		}

		patch, err := jsonpatch.CreateMergePatch(objJson, modJson)
		if err != nil {
			return false, err
		}

		_, err = ri.Patch(accessor.GetName(), types.MergePatchType, patch)
		defer ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})

		if kerr.IsForbidden(err) {
			glog.Infof("failed to update test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		return false, ErrInactiveWebhook
	} else if d.op == v1beta1.Delete {
		_, err := ri.Create(&u)
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else {
			return false, err
		}

		err = ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
		if kerr.IsForbidden(err) {
			defer func() {
				// update to make it valid
				mod := d.obj.DeepCopyObject()
				d.transform(mod)
				modJson, err := json.Marshal(mod)
				if err != nil {
					return
				}

				patch, err := jsonpatch.CreateMergePatch(objJson, modJson)
				if err != nil {
					return
				}

				ri.Patch(accessor.GetName(), types.MergePatchType, patch)

				// delete
				ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
			}()

			glog.Infof("failed to delete test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return false, ErrInactiveWebhook
	}

	return false, nil
}
