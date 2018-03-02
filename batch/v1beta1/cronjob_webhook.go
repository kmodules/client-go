package v1beta1

import (
	"net/http"
	"sync"

	"github.com/appscode/kutil/admission/api"
	"github.com/appscode/kutil/meta"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/api/batch/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CronJobWebhook struct {
	client   kubernetes.Interface
	handler  api.ResourceHandler
	plural   schema.GroupVersionResource
	singular string

	initialized bool
	lock        sync.RWMutex
}

var _ api.AdmissionHook = &CronJobWebhook{}

func (a *CronJobWebhook) Resource() (plural schema.GroupVersionResource, singular string) {
	return plural, singular
}

func (a *CronJobWebhook) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	a.client, err = kubernetes.NewForConfig(config)
	return err
}

func (a *CronJobWebhook) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if a.handler == nil ||
		(req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		(req.Kind.Group != v1beta1.GroupName) ||
		req.Kind.Kind != "CronJob" {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		status.Allowed = false
		status.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusInternalServerError, Reason: metav1.StatusReasonInternalError,
			Message: "not initialized",
		}
		return status
	}
	gv := schema.GroupVersion{Group: req.Kind.Group, Version: req.Kind.Version}

	switch req.Operation {
	case admission.Delete:
		// req.Object.Raw = nil, so read from kubernetes
		obj, err := a.client.BatchV1beta1().CronJobs(req.Namespace).Get(req.Name, metav1.GetOptions{})
		if err == nil {
			err2 := a.handler.OnDelete(obj)
			if err2 != nil {
				status.Allowed = false
				status.Result = &metav1.Status{
					Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
					Message: err.Error(),
				}
				return status
			}
		}
	case admission.Create:
		v1Obj, originalObj, err := convert_to_v1_cronjob(gv, req.Object.Raw)
		if err != nil {
			status.Allowed = false
			status.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: err.Error(),
			}
			return status
		}

		v1Mod, err := a.handler.OnCreate(v1Obj)
		if err != nil {
			status.Allowed = false
			status.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusForbidden, Reason: metav1.StatusReasonForbidden,
				Message: err.Error(),
			}
			return status
		} else if v1Mod != nil {
			patch, err := create_cronjob_patch(gv, originalObj, v1Mod)
			if err != nil {
				status.Allowed = false
				status.Result = &metav1.Status{
					Status: metav1.StatusFailure, Code: http.StatusInternalServerError, Reason: metav1.StatusReasonInternalError,
					Message: err.Error(),
				}
				return status
			}
			status.Patch = patch
			patchType := admission.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	case admission.Update:
		v1Obj, originalObj, err := convert_to_v1_cronjob(gv, req.Object.Raw)
		if err != nil {
			status.Allowed = false
			status.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: err.Error(),
			}
			return status
		}
		v1OldObj, _, err := convert_to_v1_cronjob(gv, req.OldObject.Raw)
		if err != nil {
			status.Allowed = false
			status.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: err.Error(),
			}
			return status
		}

		v1Mod, err := a.handler.OnUpdate(v1OldObj, v1Obj)
		if err != nil {
			status.Allowed = false
			status.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusForbidden, Reason: metav1.StatusReasonForbidden,
				Message: err.Error(),
			}
			return status
		} else if v1Mod != nil {
			patch, err := create_cronjob_patch(gv, originalObj, v1Mod)
			if err != nil {
				status.Allowed = false
				status.Result = &metav1.Status{
					Status: metav1.StatusFailure, Code: http.StatusInternalServerError, Reason: metav1.StatusReasonInternalError,
					Message: err.Error(),
				}
				return status
			}
			status.Patch = patch
			patchType := admission.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	}

	status.Allowed = true
	return status
}

func convert_to_v1_cronjob(gv schema.GroupVersion, raw []byte) (*v1beta1.CronJob, runtime.Object, error) {
	switch gv {
	case v1beta1.SchemeGroupVersion:
		v1Obj, err := meta.UnmarshalToJSON(raw, v1beta1.SchemeGroupVersion)
		if err != nil {
			return nil, nil, err
		}
		return v1Obj.(*v1beta1.CronJob), v1Obj, nil
	}
	return nil, nil, errors.New("unknown")
}

func create_cronjob_patch(gv schema.GroupVersion, originalObj, v1Mod interface{}) ([]byte, error) {
	switch gv {
	case v1beta1.SchemeGroupVersion:
		return meta.CreateJSONMergePatch(originalObj.(runtime.Object), v1Mod.(runtime.Object))
	}
	return nil, errors.New("unknown")
}
