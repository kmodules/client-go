package v1beta1

import (
	"github.com/appscode/kutil"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	apps "k8s.io/api/admissionregistration/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func CreateOrPatchValidatingWebhookConfiguration(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*apps.ValidatingWebhookConfiguration) *apps.ValidatingWebhookConfiguration) (*apps.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	cur, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating ValidatingWebhookConfiguration %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(transform(&apps.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ValidatingWebhookConfiguration",
				APIVersion: apps.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchValidatingWebhookConfiguration(c, cur, transform)
}

func PatchValidatingWebhookConfiguration(c kubernetes.Interface, cur *apps.ValidatingWebhookConfiguration, transform func(*apps.ValidatingWebhookConfiguration) *apps.ValidatingWebhookConfiguration) (*apps.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	return PatchValidatingWebhookConfigurationObject(c, cur, transform(cur.DeepCopy()))
}

func PatchValidatingWebhookConfigurationObject(c kubernetes.Interface, cur, mod *apps.ValidatingWebhookConfiguration) (*apps.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, apps.ValidatingWebhookConfiguration{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching ValidatingWebhookConfiguration %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Patch(cur.Name, types.StrategicMergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateValidatingWebhookConfiguration(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*apps.ValidatingWebhookConfiguration) *apps.ValidatingWebhookConfiguration) (result *apps.ValidatingWebhookConfiguration, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update ValidatingWebhookConfiguration %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update ValidatingWebhookConfiguration %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
