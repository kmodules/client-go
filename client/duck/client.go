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

type Client struct {
	c       client.Client // reader?
	obj     Object
	duckGVK schema.GroupVersionKind
	rawGVK  schema.GroupVersionKind
}

var (
	_ client.Reader       = &Client{}
	_ client.Writer       = &Client{}
	_ client.StatusClient = &Client{}
)

type ClientBuilder struct {
	cc *Client
}

func NewClient() *ClientBuilder {
	return &ClientBuilder{
		cc: new(Client),
	}
}

func (b *ClientBuilder) ForDuckType(obj Object) *ClientBuilder {
	b.cc.obj = obj
	return b
}

func (b *ClientBuilder) WithUnderlyingType(rawGVK schema.GroupVersionKind) *ClientBuilder {
	b.cc.rawGVK = rawGVK
	return b
}

func (b *ClientBuilder) Build(c client.Client) (client.Client, error) {
	b.cc.c = c
	gvk, err := apiutil.GVKForObject(b.cc.obj, c.Scheme())
	if err != nil {
		return nil, err
	}
	b.cc.duckGVK = gvk
	return b.cc, nil
}

// Scheme returns the scheme this client is using.
func (d *Client) Scheme() *runtime.Scheme {
	return d.c.Scheme()
}

// RESTMapper returns the rest this client is using.
func (d *Client) RESTMapper() apimeta.RESTMapper {
	return d.c.RESTMapper()
}

func (d *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
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

func (d *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, d.c.Scheme())
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") && apimeta.IsListType(list) {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	if gvk != d.duckGVK {
		return d.c.List(ctx, list, opts...)
	}

	listGVK := d.rawGVK
	listGVK.Kind += "List"

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

func (d *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Create(ctx, obj, opts...)
	}
	return fmt.Errorf("create not supported for duck type %+v", d.duckGVK)
}

func (d *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
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

func (d *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	gvk, err := apiutil.GVKForObject(obj, d.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != d.duckGVK {
		return d.c.Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", d.duckGVK)
}

func (d *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
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

func (d *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
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

func (d *Client) Status() client.StatusWriter {
	return &statusWriter{client: d}
}

// statusWriter is client.StatusWriter that writes status subresource.
type statusWriter struct {
	client *Client
}

// ensure statusWriter implements client.StatusWriter.
var _ client.StatusWriter = &statusWriter{}

func (sw *statusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	gvk, err := apiutil.GVKForObject(obj, sw.client.c.Scheme())
	if err != nil {
		return err
	}
	if gvk != sw.client.duckGVK {
		return sw.client.c.Status().Update(ctx, obj, opts...)
	}
	return fmt.Errorf("update not supported for duck type %+v", sw.client.duckGVK)
}

func (sw *statusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
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
