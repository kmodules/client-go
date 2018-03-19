package v1beta1

import (
	"bytes"
	"fmt"
	"sync"

	jp "github.com/appscode/jsonpatch"
	"github.com/appscode/kutil/admission"
	"github.com/appscode/kutil/runtime/serializer/versioning"
	workload "github.com/appscode/kutil/workload/v1"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
)

// WorkloadWebhook avoids the bidirectional conversion needed for GenericWebhooks. Only supports workload types.
type WorkloadWebhook struct {
	plural   schema.GroupVersionResource
	singular string

	srcGroups sets.String
	target    schema.GroupVersionKind
	factory   GetterFactory
	get       GetFunc
	handler   admission.ResourceHandler

	initialized bool
	lock        sync.RWMutex
}

var _ AdmissionHook = &WorkloadWebhook{}

func NewWorkloadWebhook(
	plural schema.GroupVersionResource,
	singular string,
	target schema.GroupVersionKind,
	factory GetterFactory,
	handler admission.ResourceHandler) *WorkloadWebhook {
	return &WorkloadWebhook{
		plural:    plural,
		singular:  singular,
		srcGroups: sets.NewString(core.GroupName, appsv1.GroupName, extensions.GroupName, batchv1.GroupName),
		target:    target,
		factory:   factory,
		handler:   handler,
	}
}

func (h *WorkloadWebhook) Resource() (schema.GroupVersionResource, string) {
	return h.plural, h.singular
}

func (h *WorkloadWebhook) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.initialized = true

	var err error
	if h.factory != nil {
		h.get, err = h.factory.New(config)
	}
	return err
}

func (h *WorkloadWebhook) Admit(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	status := &v1beta1.AdmissionResponse{}

	if h.handler == nil ||
		(req.Operation != v1beta1.Create && req.Operation != v1beta1.Update && req.Operation != v1beta1.Delete) ||
		len(req.SubResource) != 0 ||
		!h.srcGroups.Has(req.Kind.Group) ||
		req.Kind.Kind != h.target.Kind {
		status.Allowed = true
		return status
	}

	h.lock.RLock()
	defer h.lock.RUnlock()
	if !h.initialized {
		return StatusUninitialized()
	}

	codec := versioning.Serializer
	gvk := schema.GroupVersionKind{Group: req.Kind.Group, Version: req.Kind.Version, Kind: req.Kind.Kind}

	switch req.Operation {
	case v1beta1.Delete:
		if h.get == nil {
			break
		}
		// req.Object.Raw = nil, so read from kubernetes
		obj, err := h.get(req.Namespace, req.Name)
		if err != nil && !kerr.IsNotFound(err) {
			return StatusInternalServerError(err)
		} else if err == nil {
			err2 := h.handler.OnDelete(obj)
			if err2 != nil {
				return StatusBadRequest(err)
			}
		}
	case v1beta1.Create:
		obj, kind, err := codec.Decode(req.Object.Raw, &gvk, nil)
		if err != nil {
			return StatusBadRequest(err)
		}
		obj.GetObjectKind().SetGroupVersionKind(*kind)
		w, err := h.convertToWorkload(obj)
		if err != nil {
			return StatusBadRequest(err)
		}

		mod, err := h.handler.OnCreate(w)
		if err != nil {
			return StatusForbidden(err)
		} else if mod != nil {
			err = h.applyWorkload(obj, mod.(*workload.Workload))
			if err != nil {
				return StatusForbidden(err)
			}

			var buf bytes.Buffer
			err = codec.Encode(obj, &buf)
			if err != nil {
				return StatusBadRequest(err)
			}
			ops, err := jp.CreatePatch(req.Object.Raw, buf.Bytes())
			if err != nil {
				return StatusBadRequest(err)
			}
			patch, err := json.Marshal(ops)
			if err != nil {
				return StatusInternalServerError(err)
			}
			status.Patch = patch
			patchType := v1beta1.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	case v1beta1.Update:
		obj, kind, err := codec.Decode(req.Object.Raw, &gvk, nil)
		if err != nil {
			return StatusBadRequest(err)
		}
		obj.GetObjectKind().SetGroupVersionKind(*kind)
		w, err := h.convertToWorkload(obj)
		if err != nil {
			return StatusBadRequest(err)
		}

		oldObj, kind, err := codec.Decode(req.OldObject.Raw, &gvk, nil)
		if err != nil {
			return StatusBadRequest(err)
		}
		oldObj.GetObjectKind().SetGroupVersionKind(*kind)
		ow, err := h.convertToWorkload(oldObj)
		if err != nil {
			return StatusBadRequest(err)
		}

		mod, err := h.handler.OnUpdate(ow, w)
		if err != nil {
			return StatusForbidden(err)
		} else if mod != nil {
			err = h.applyWorkload(obj, mod.(*workload.Workload))
			if err != nil {
				return StatusForbidden(err)
			}

			var buf bytes.Buffer
			err = codec.Encode(obj, &buf)
			if err != nil {
				return StatusBadRequest(err)
			}
			ops, err := jp.CreatePatch(req.Object.Raw, buf.Bytes())
			if err != nil {
				return StatusBadRequest(err)
			}
			patch, err := json.Marshal(ops)
			if err != nil {
				return StatusInternalServerError(err)
			}
			status.Patch = patch
			patchType := v1beta1.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	}

	status.Allowed = true
	return status
}

// ref: https://github.com/kubernetes/kubernetes/blob/4f083dee54539b0ca24ddc55d53921f5c2efc0b9/pkg/kubectl/cmd/util/factory_client_access.go#L221
func (h *WorkloadWebhook) convertToWorkload(obj runtime.Object) (*workload.Workload, error) {
	switch t := obj.(type) {
	case *core.Pod:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec), nil
		// ReplicationController
	case *core.ReplicationController:
		if t.Spec.Template == nil {
			t.Spec.Template = &core.PodTemplateSpec{}
		}
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// Deployment
	case *extensions.Deployment:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1beta1.Deployment:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1beta2.Deployment:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1.Deployment:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// DaemonSet
	case *extensions.DaemonSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1beta2.DaemonSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1.DaemonSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// ReplicaSet
	case *extensions.ReplicaSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1beta2.ReplicaSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1.ReplicaSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// StatefulSet
	case *appsv1beta1.StatefulSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1beta2.StatefulSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
	case *appsv1.StatefulSet:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// Job
	case *batchv1.Job:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.Template.Spec), nil
		// CronJob
	case *batchv1beta1.CronJob:
		return newWorkload(t.TypeMeta, t.ObjectMeta, t.Spec.JobTemplate.Spec.Template.Spec), nil
	default:
		return nil, fmt.Errorf("the object is not a pod or does not have a pod template")
	}
}

func newWorkload(t metav1.TypeMeta, o metav1.ObjectMeta, spec core.PodSpec) *workload.Workload {
	return &workload.Workload{
		TypeMeta:   t,
		ObjectMeta: o,
		Spec:       spec,
	}
}

func (h *WorkloadWebhook) applyWorkload(obj runtime.Object, w *workload.Workload) error {
	switch t := obj.(type) {
	case *core.Pod:
		t.ObjectMeta = w.ObjectMeta
		t.Spec = w.Spec
		// ReplicationController
	case *core.ReplicationController:
		if t.Spec.Template == nil {
			t.Spec.Template = &core.PodTemplateSpec{}
		}
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// Deployment
	case *extensions.Deployment:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1beta1.Deployment:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1beta2.Deployment:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1.Deployment:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// DaemonSet
	case *extensions.DaemonSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1beta2.DaemonSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1.DaemonSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// ReplicaSet
	case *extensions.ReplicaSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1beta2.ReplicaSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1.ReplicaSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// StatefulSet
	case *appsv1beta1.StatefulSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1beta2.StatefulSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
	case *appsv1.StatefulSet:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// Job
	case *batchv1.Job:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.Template.Spec = w.Spec
		// CronJob
	case *batchv1beta1.CronJob:
		t.ObjectMeta = w.ObjectMeta
		t.Spec.JobTemplate.Spec.Template.Spec = w.Spec
	default:
		return fmt.Errorf("the object is not a pod or does not have a pod template")
	}
	return nil
}
