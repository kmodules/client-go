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

package configreader

import (
	"context"
	"fmt"
	"reflect"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/pager"
)

var _ v1.ConfigMapNamespaceLister = &configmapNamespaceLister{}

// configmapNamespaceLister implements the NamespaceLister interface.
type configmapNamespaceLister struct {
	dc        kubernetes.Interface
	namespace string
}

// List lists all resources in the indexer for a given namespace.
func (l *configmapNamespaceLister) List(selector labels.Selector) (ret []*core.ConfigMap, err error) {
	fn := func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		return l.dc.CoreV1().ConfigMaps(l.namespace).List(ctx, opts)
	}
	opts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	err = pager.New(fn).EachListItem(context.TODO(), opts, func(obj runtime.Object) error {
		o, ok := obj.(*core.ConfigMap)
		if !ok {
			return fmt.Errorf("expected *core.ConfigMap, found %s", reflect.TypeOf(obj))
		}
		ret = append(ret, o)
		return nil
	})
	return ret, err
}

// Get retrieves a resource from the indexer for a given namespace and name.
func (l *configmapNamespaceLister) Get(name string) (*core.ConfigMap, error) {
	obj, err := l.dc.CoreV1().ConfigMaps(l.namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
