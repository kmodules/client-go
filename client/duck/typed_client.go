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
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type typedClient struct {
	c       client.Client
	duckGVK schema.GroupVersionKind
	rawGVK  schema.GroupVersionKind
}

var (
	_ client.Reader       = &typedClient{}
	_ client.Writer       = &typedClient{}
	_ client.StatusClient = &typedClient{}
)

// Scheme returns the scheme this client is using.
func (d *typedClient) Scheme() *runtime.Scheme {
	return d.c.Scheme()
}

// RESTMapper returns the rest this client is using.
func (d *typedClient) RESTMapper() apimeta.RESTMapper {
	return d.c.RESTMapper()
}

func (d *typedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Get(ctx, key, obj, opts...)
	}

	ll, err := d.c.Scheme().New(d.rawGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.Object)
	err = d.c.Get(ctx, key, llo, opts...)
	if err != nil {
		return err
	}

	dd := obj.(Object)
	return dd.Duckify(llo)
}

func (d *typedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, d.c.Scheme())
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, listType) && apimeta.IsListType(list) {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	if gvk != d.duckGVK {
		return d.c.List(ctx, list, opts...)
	}

	listGVK := d.rawGVK
	listGVK.Kind += listType

	ll, err := d.c.Scheme().New(listGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.ObjectList)
	err = d.c.List(ctx, llo, opts...)
	if err != nil {
		return err
	}

	list.SetResourceVersion(llo.GetResourceVersion())
	list.SetContinue(llo.GetContinue())
	list.SetSelfLink(llo.GetSelfLink())
	list.SetRemainingItemCount(llo.GetRemainingItemCount())

	items := make([]runtime.Object, 0, apimeta.LenList(llo))
	err = apimeta.EachListItem(llo, func(object runtime.Object) error {
		d2, err := d.c.Scheme().New(d.duckGVK)
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

func (d *typedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Create(ctx, obj, opts...)
	}
	return fmt.Errorf("create not supported for duck type %+v", d.duckGVK)
}

func (d *typedClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Delete(ctx, obj, opts...)
	}

	ll, err := d.c.Scheme().New(d.rawGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.Object)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return d.c.Delete(ctx, llo, opts...)
}

func (d *typedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", d.duckGVK)
}

func (d *typedClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Patch(ctx, obj, patch, opts...)
	}

	ll, err := d.c.Scheme().New(d.rawGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.Object)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return d.c.Patch(ctx, llo, patch, opts...)
}

func (d *typedClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.DeleteAllOf(ctx, obj, opts...)
	}

	ll, err := d.c.Scheme().New(d.rawGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.Object)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return d.c.DeleteAllOf(ctx, llo, opts...)
}

func (d *typedClient) Status() client.StatusWriter {
	return &typedStatusWriter{client: d}
}

// typedStatusWriter is client.StatusWriter that writes status subresource.
type typedStatusWriter struct {
	client *typedClient
}

// ensure typedStatusWriter implements client.StatusWriter.
var _ client.StatusWriter = &typedStatusWriter{}

func (sw *typedStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	gvk, err := apiutil.GVKForObject(obj, sw.client.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != sw.client.duckGVK {
		return sw.client.c.Status().Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", sw.client.duckGVK)
}

func (sw *typedStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	gvk, err := apiutil.GVKForObject(obj, sw.client.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != sw.client.duckGVK {
		return sw.client.c.Status().Patch(ctx, obj, patch, opts...)
	}

	ll, err := sw.client.c.Scheme().New(sw.client.rawGVK)
	if err != nil {
		return err
	}
	llo := ll.(client.Object)
	llo.SetNamespace(obj.GetNamespace())
	llo.SetName(obj.GetName())
	llo.SetLabels(obj.GetLabels())
	return sw.client.c.Status().Patch(ctx, llo, patch, opts...)
}
