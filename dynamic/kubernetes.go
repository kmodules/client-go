/*
Copyright The Kmodules Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dynamic

import (
	"context"
	"fmt"
	"time"

	v1 "kmodules.xyz/client-go/core/v1"
	discovery_util "kmodules.xyz/client-go/discovery"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	kutil "kmodules.xyz/client-go"
)

func WaitUntilDeleted(ri dynamic.ResourceInterface, stopCh <-chan struct{}, name string, subresources ...string) error {
	err := ri.Delete(name, &metav1.DeleteOptions{}, subresources...)
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	// delete operation was successful, now wait for obj to be removed(eg: objects with finalizers)
	return wait.PollImmediateUntil(kutil.RetryInterval, func() (bool, error) {
		_, e2 := ri.Get(name, metav1.GetOptions{}, subresources...)
		if kerr.IsNotFound(e2) {
			return true, nil
		} else if e2 != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return false, nil
	}, stopCh)
}

func UntilHasLabel(config *rest.Config, gvk schema.GroupVersionKind, namespace, name string, key string, value *string, timeout time.Duration) (out string, err error) {
	return untilHasKey(config, gvk, namespace, name, func(obj metav1.Object) map[string]string { return obj.GetLabels() }, key, value, timeout)
}

func UntilHasAnnotation(config *rest.Config, gvk schema.GroupVersionKind, namespace, name string, key string, value *string, timeout time.Duration) (out string, err error) {
	return untilHasKey(config, gvk, namespace, name, func(obj metav1.Object) map[string]string { return obj.GetAnnotations() }, key, value, timeout)
}

func untilHasKey(
	config *rest.Config,
	gvk schema.GroupVersionKind,
	namespace, name string,
	fn func(metav1.Object) map[string]string,
	key string, value *string,
	timeout time.Duration) (out string, err error) {

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	kc := kubernetes.NewForConfigOrDie(config)
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return
	}

	gvr, err := discovery_util.ResourceForGVK(kc.Discovery(), gvk)
	if err != nil {
		return
	}

	var ri dynamic.ResourceInterface
	if namespace != "" {
		ri = dc.Resource(gvr).Namespace(namespace)
	} else {
		ri = dc.Resource(gvr)
	}

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, name).String()
			return ri.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, name).String()
			return ri.Watch(options)
		},
	}

	_, err = watchtools.UntilWithSync(ctx,
		lw,
		&unstructured.Unstructured{},
		nil,
		func(event watch.Event) (bool, error) {
			switch event.Type {
			case watch.Deleted:
				return false, nil
			case watch.Error:
				return false, errors.Wrap(err, "error watching")
			case watch.Added, watch.Modified:
				m, e2 := meta.Accessor(event.Object)
				if e2 != nil {
					return false, e2
				}
				var ok bool
				if out, ok = fn(m)[key]; ok && (value == nil || *value == out) {
					return true, nil
				}
				return false, nil // continue
			default:
				return false, fmt.Errorf("unexpected event type: %v", event.Type)
			}
		},
	)
	return
}

func DetectWorkload(config *rest.Config, resource schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, schema.GroupVersionResource, error) {
	kc := kubernetes.NewForConfigOrDie(config)
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, resource, err
	}

	var ri dynamic.ResourceInterface
	if namespace != "" {
		ri = dc.Resource(resource).Namespace(namespace)
	} else {
		ri = dc.Resource(resource)
	}

	obj, err := ri.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, resource, err
	}
	return findWorkload(kc, dc, resource, obj)
}

func findWorkload(kc kubernetes.Interface, dc dynamic.Interface, resource schema.GroupVersionResource, obj *unstructured.Unstructured) (*unstructured.Unstructured, schema.GroupVersionResource, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, resource, err
	}
	for _, ref := range m.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
			ar, err := discovery_util.APIResourceForGVK(kc.Discovery(), gvk)
			if err != nil {
				return nil, schema.GroupVersionResource{}, err
			}
			gvr := schema.GroupVersionResource{
				Group:    ar.Group,
				Version:  ar.Version,
				Resource: ar.Name,
			}
			var ri dynamic.ResourceInterface
			if ar.Namespaced {
				ri = dc.Resource(gvr).Namespace(m.GetNamespace())
			} else {
				ri = dc.Resource(gvr)
			}
			parent, err := ri.Get(ref.Name, metav1.GetOptions{})
			if err != nil {
				return nil, schema.GroupVersionResource{}, err
			}
			return findWorkload(kc, dc, gvr, parent)
		}
	}
	return obj, resource, nil
}

func RemoveOwnerReferenceForItems(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	items []string,
	ref *core.ObjectReference,
) error {
	var ri dynamic.ResourceInterface
	if namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(namespace)
	}

	var errs []error
	for _, name := range items {
		item, err := ri.Get(name, metav1.GetOptions{})
		if err != nil {
			if !kerr.IsNotFound(err) {
				errs = append(errs, err)
			}
			continue
		}
		if _, _, err := Patch(c, gvr, item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.RemoveOwnerReference(in, ref)
			return in
		}); err != nil && !kerr.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func RemoveOwnerReferenceForSelector(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	selector labels.Selector,
	ref *core.ObjectReference,
) error {
	var ri dynamic.ResourceInterface
	if namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(namespace)
	}

	list, err := ri.List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}

	var errs []error
	for _, item := range list.Items {
		if _, _, err := Patch(c, gvr, &item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.RemoveOwnerReference(in, ref)
			return in
		}); err != nil && !kerr.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func EnsureOwnerReferenceForItems(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	items []string,
	ref *core.ObjectReference,
) error {
	var ri dynamic.ResourceInterface
	if namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(namespace)
	}

	var errs []error
	for _, name := range items {
		item, err := ri.Get(name, metav1.GetOptions{})
		if err != nil {
			if !kerr.IsNotFound(err) {
				errs = append(errs, err)
			}
			continue
		}
		if _, _, err := Patch(c, gvr, item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.EnsureOwnerReference(in, ref)
			return in
		}); err != nil && !kerr.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func EnsureOwnerReferenceForSelector(
	c dynamic.Interface,
	gvr schema.GroupVersionResource,
	namespace string,
	selector labels.Selector,
	ref *core.ObjectReference,
) error {
	var ri dynamic.ResourceInterface
	if namespace == "" {
		ri = c.Resource(gvr)
	} else {
		ri = c.Resource(gvr).Namespace(namespace)
	}
	list, err := ri.List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}

	var errs []error
	for _, item := range list.Items {
		if _, _, err := Patch(c, gvr, &item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.EnsureOwnerReference(in, ref)
			return in
		}); err != nil && !kerr.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}
