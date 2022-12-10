/*
Copyright AppsCode Inc. and Contributors

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

package duck

import (
	"context"
	"fmt"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// client is a client.Client that reads and writes directly from/to an API server.  It lazily initializes
// new clients at the time they are used, and caches the client.
type unstructuredClient struct {
	c       client.Client
	duckGVK schema.GroupVersionKind
	rawGVK  schema.GroupVersionKind
}

var (
	_ client.Reader       = &unstructuredClient{}
	_ client.Writer       = &unstructuredClient{}
	_ client.StatusClient = &unstructuredClient{}
)

// Scheme returns the scheme this client is using.
func (d *unstructuredClient) Scheme() *runtime.Scheme {
	return d.c.Scheme()
}

// RESTMapper returns the rest this client is using.
func (d *unstructuredClient) RESTMapper() apimeta.RESTMapper {
	return d.c.RESTMapper()
}

// Create implements client.Client.
func (uc *unstructuredClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.Create(ctx, obj, opts...)
	}
	return fmt.Errorf("create not supported for duck type %+v", uc.duckGVK)
}

// Update implements client.Client.
func (uc *unstructuredClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", uc.duckGVK)
}

// Delete implements client.Client.
func (uc *unstructuredClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.Delete(ctx, obj, opts...)
	}

	var llo unstructured.Unstructured
	llo.GetObjectKind().SetGroupVersionKind(uc.rawGVK)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return uc.c.Delete(ctx, &llo, opts...)
}

// DeleteAllOf implements client.Client.
func (uc *unstructuredClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.DeleteAllOf(ctx, obj, opts...)
	}

	var llo unstructured.Unstructured
	llo.GetObjectKind().SetGroupVersionKind(uc.rawGVK)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return uc.c.DeleteAllOf(ctx, &llo, opts...)
}

// Patch implements client.Client.
func (uc *unstructuredClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.Patch(ctx, obj, patch, opts...)
	}

	var llo unstructured.Unstructured
	llo.GetObjectKind().SetGroupVersionKind(uc.rawGVK)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return uc.c.Patch(ctx, &llo, patch, opts...)
}

// Get implements client.Client.
func (uc *unstructuredClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.Get(ctx, key, obj, opts...)
	}

	var llo unstructured.Unstructured
	llo.GetObjectKind().SetGroupVersionKind(uc.rawGVK)
	err := uc.c.Get(ctx, key, &llo, opts...)
	if err != nil {
		return err
	}

	dd := obj.(Object)
	return dd.Duckify(&llo)
}

// List implements client.Client.
func (uc *unstructuredClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	u, ok := list.(*unstructured.UnstructuredList)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", list)
	}

	gvk := u.GroupVersionKind()
	if gvk != uc.duckGVK {
		return uc.c.List(ctx, list, opts...)
	}

	listGVK := uc.rawGVK
	listGVK.Kind += listType

	var llo unstructured.UnstructuredList
	llo.GetObjectKind().SetGroupVersionKind(listGVK)
	err := uc.c.List(ctx, &llo, opts...)
	if err != nil {
		return err
	}

	list.SetResourceVersion(llo.GetResourceVersion())
	list.SetContinue(llo.GetContinue())
	list.SetSelfLink(llo.GetSelfLink())
	list.SetRemainingItemCount(llo.GetRemainingItemCount())

	items := make([]runtime.Object, 0, apimeta.LenList(&llo))
	err = apimeta.EachListItem(&llo, func(object runtime.Object) error {
		d2, err := uc.c.Scheme().New(uc.duckGVK)
		if err != nil {
			return err
		}
		dd := d2.(Object)
		err = dd.Duckify(object)
		if err != nil {
			return err
		}
		items = append(items, d2)
		return nil
	})
	if err != nil {
		return err
	}
	return apimeta.SetList(list, items)
}

func (uc *unstructuredClient) Status() client.StatusWriter {
	return &unstructuredStatusWriter{client: uc}
}

// unstructuredStatusWriter is client.StatusWriter that writes status subresource.
type unstructuredStatusWriter struct {
	client *unstructuredClient
}

// ensure unstructuredStatusWriter implements client.StatusWriter.
var _ client.StatusWriter = &unstructuredStatusWriter{}

func (sw *unstructuredStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	gvk, err := apiutil.GVKForObject(obj, sw.client.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != sw.client.duckGVK {
		return sw.client.c.Status().Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", sw.client.duckGVK)
}

func (sw *unstructuredStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	gvk, err := apiutil.GVKForObject(obj, sw.client.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != sw.client.duckGVK {
		return sw.client.c.Status().Patch(ctx, obj, patch, opts...)
	}

	var llo unstructured.Unstructured
	llo.GetObjectKind().SetGroupVersionKind(sw.client.rawGVK)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return sw.client.c.Status().Patch(ctx, &llo, patch, opts...)
}
