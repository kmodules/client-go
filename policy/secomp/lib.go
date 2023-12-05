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

package secomp

import (
	"fmt"
	"sync"

	"kmodules.xyz/client-go/discovery"
	meta_util "kmodules.xyz/client-go/meta"

	"github.com/spf13/pflag"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	profile string
	seccomp *core.SeccompProfile
	once    sync.Once
)

func init() {
	pflag.StringVar(&profile, "default-seccomp-profile-type", profile, "Default seccomp profile")
}

func DefaultSeccompProfile() *core.SeccompProfile {
	once.Do(func() {
		if meta_util.PossiblyInCluster() {
			cfg, err := rest.InClusterConfig()
			if err != nil {
				panic(err)
			}
			kc := kubernetes.NewForConfigOrDie(cfg)
			yes, err := discovery.CheckAPIVersion(kc.Discovery(), ">= 1.27")
			if err != nil {
				panic(err)
			}
			if yes && profile == "" {
				profile = string(core.SeccompProfileTypeRuntimeDefault)
			}
		}
		if profile == "" {
			seccomp = nil
		} else if profile != string(core.SeccompProfileTypeUnconfined) &&
			profile != string(core.SeccompProfileTypeRuntimeDefault) &&
			profile != string(core.SeccompProfileTypeLocalhost) {
			panic(fmt.Errorf("unknown seccomp profile type %s", profile))
		}
		seccomp = &core.SeccompProfile{
			Type: core.SeccompProfileType(profile),
		}
	})
	return seccomp
}
