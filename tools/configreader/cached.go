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
	"fmt"
	"reflect"
	"sync"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/listers/core/v1"
)

type cachedImpl struct {
	factory informers.SharedInformerFactory
	stopCh  <-chan struct{}

	lock         sync.RWMutex
	cfgLister    v1.ConfigMapLister
	secretLister v1.SecretLister
}

var _ ConfigReader = &cachedImpl{}

func (i *cachedImpl) ConfigMaps(namespace string) v1.ConfigMapNamespaceLister {
	i.lock.RLock()
	if i.cfgLister != nil {
		i.lock.RUnlock()
		return i.cfgLister.ConfigMaps(namespace)
	}
	i.lock.RUnlock()

	createLister := func() v1.ConfigMapLister {
		i.lock.Lock()
		defer i.lock.Unlock()
		if i.cfgLister != nil {
			return i.cfgLister
		}

		informerType := reflect.TypeOf(&core.ConfigMap{})
		informerDep, _ := i.factory.ForResource(core.SchemeGroupVersion.WithResource("configmaps"))
		i.factory.Start(i.stopCh)
		if synced := i.factory.WaitForCacheSync(i.stopCh); !synced[informerType] {
			panic(fmt.Sprintf("informer for %s hasn't synced", informerType))
		}
		i.cfgLister = v1.NewConfigMapLister(informerDep.Informer().GetIndexer())
		return i.cfgLister
	}
	return createLister().ConfigMaps(namespace)
}

func (i *cachedImpl) Secrets(namespace string) v1.SecretNamespaceLister {
	i.lock.RLock()
	if i.secretLister != nil {
		i.lock.RUnlock()
		return i.secretLister.Secrets(namespace)
	}
	i.lock.RUnlock()

	createLister := func() v1.SecretLister {
		i.lock.Lock()
		defer i.lock.Unlock()
		if i.secretLister != nil {
			return i.secretLister
		}

		informerType := reflect.TypeOf(&core.Secret{})
		informerDep, _ := i.factory.ForResource(core.SchemeGroupVersion.WithResource("secrets"))
		i.factory.Start(i.stopCh)
		if synced := i.factory.WaitForCacheSync(i.stopCh); !synced[informerType] {
			panic(fmt.Sprintf("informer for %s hasn't synced", informerType))
		}
		i.secretLister = v1.NewSecretLister(informerDep.Informer().GetIndexer())
		return i.secretLister
	}
	return createLister().Secrets(namespace)
}
