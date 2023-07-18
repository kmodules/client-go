/*
Copyright 2020 The Kubernetes Authors.

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

package committer

import (
	"context"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Patcher is just the Patch API with a generic to keep use sites type safe
type Patcher[R any] interface {
	Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error
}

type NamespacedPatcher[R any] func(ns string) Patcher[R]

type MyInterface[Sp any, St any] interface {
	GetStatus() St
	client.Object
}

func NewStatusCommitter[R any, Sp any, St any](patcher NamespacedPatcher[R]) func(context.Context, MyInterface[Sp, St], MyInterface[Sp, St]) error {
	focusType := fmt.Sprintf("%T", *new(R))
	return func(ctx context.Context, old, obj MyInterface[Sp, St]) error {
		logger := klog.FromContext(ctx)
		statusChanged := !equality.Semantic.DeepEqual(old.GetStatus(), obj.GetStatus())
		if !statusChanged {
			return nil
		}

		ns := old.GetNamespace()
		name := old.GetName()

		oldData, err := json.Marshal(old.GetStatus())
		if err != nil {
			return fmt.Errorf("failed to Marshal old data for %s %s/%s: %w", focusType, ns, name, err)
		}

		newData, err := json.Marshal(obj.GetStatus())
		if err != nil {
			return fmt.Errorf("failed to Marshal new data for %s %s/%s: %w", focusType, ns, name, err)
		}

		patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
		if err != nil {
			return fmt.Errorf("failed to create patch for %s %s/%s: %w", focusType, ns, name, err)
		}

		logger.Info(fmt.Sprintf("patching %s", focusType), "patch", string(patchBytes))
		patch := client.MergeFrom(old)
		return patcher(ns).Patch(ctx, obj, patch)
	}
}
