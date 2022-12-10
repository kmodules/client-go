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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const listType = "List"

type ClientBuilder struct {
	// rawGVK schema.GroupVersionKind
	rawObj  client.Object
	duckObj Object
}

func NewClient() *ClientBuilder {
	return &ClientBuilder{}
}

func (b *ClientBuilder) ForDuckType(obj Object) *ClientBuilder {
	b.duckObj = obj
	return b
}

func (b *ClientBuilder) WithUnderlyingType(obj client.Object) *ClientBuilder {
	b.rawObj = obj
	return b
}

func (b *ClientBuilder) Build(c client.Client) (client.Client, error) {
	duckGVK, err := apiutil.GVKForObject(b.duckObj, c.Scheme())
	if err != nil {
		return nil, err
	}

	_, isUnstructured := b.rawObj.(*unstructured.Unstructured)
	if isUnstructured {
		return &unstructuredClient{
			c:       c,
			duckGVK: duckGVK,
			rawGVK:  b.rawObj.GetObjectKind().GroupVersionKind(),
		}, nil
	}

	return &typedClient{
		c:       c,
		duckGVK: duckGVK,
		rawGVK:  b.rawObj.GetObjectKind().GroupVersionKind(),
	}, nil
}
