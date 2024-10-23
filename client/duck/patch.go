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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RawPatch struct {
	pt   types.PatchType
	data []byte
}

var _ client.Patch = &RawPatch{}

func NewRawPatch(obj client.Object, patch client.Patch) (client.Patch, error) {
	data, err := patch.Data(obj)
	if err != nil {
		return nil, err
	}
	return &RawPatch{
		pt:   patch.Type(),
		data: data,
	}, nil
}

func (r *RawPatch) Type() types.PatchType {
	return r.pt
}

func (r *RawPatch) Data(obj client.Object) ([]byte, error) {
	return r.data, nil
}
