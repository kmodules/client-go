package dynamic

import (
	"github.com/appscode/kutil/core/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func RemoveOwnerReferenceForItems(
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
	for _, item := range list.Items {
		if _, _, err := Patch(c, gvr, &item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.RemoveOwnerReference(in, ref)
			return in
		}); err != nil {
			return err
		}
	}
	return nil
}

func EnsureOwnerReferenceForItems(
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
	for _, item := range list.Items {
		if _, _, err := Patch(c, gvr, &item, func(in *unstructured.Unstructured) *unstructured.Unstructured {
			v1.EnsureOwnerReference(in, ref)
			return in
		}); err != nil {
			return err
		}
	}
	return nil
}
