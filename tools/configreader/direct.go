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
	"sync"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
)

type directImpl struct {
	dc kubernetes.Interface

	lock         sync.RWMutex
	cfgLister    v1.ConfigMapNamespaceLister
	secretLister v1.SecretNamespaceLister
}

var _ ConfigReader = &directImpl{}

func (i *directImpl) ConfigMaps(namespace string) v1.ConfigMapNamespaceLister {
	i.lock.RLock()
	defer i.lock.RUnlock()
	if i.cfgLister != nil {
		return i.cfgLister
	}

	i.cfgLister = &configmapNamespaceLister{dc: i.dc, namespace: namespace}
	return i.cfgLister
}

func (i *directImpl) Secrets(namespace string) v1.SecretNamespaceLister {
	i.lock.RLock()
	defer i.lock.RUnlock()
	if i.secretLister != nil {
		return i.secretLister
	}

	i.secretLister = &secretNamespaceLister{dc: i.dc, namespace: namespace}
	return i.secretLister
}
