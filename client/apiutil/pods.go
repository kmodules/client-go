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

package apiutil

import (
	"context"
	"strings"

	kmapi "kmodules.xyz/client-go/api/v1"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Container struct {
	Name  string
	Image string
}

func CollectImageInfo(kc client.Client, pod *core.Pod, images map[string]kmapi.ImageInfo) (map[string]kmapi.ImageInfo, error) {
	lineage, err := DetectLineage(context.TODO(), kc, pod)
	if err != nil {
		return images, err
	}

	refs := map[string][]string{}
	for _, c := range pod.Spec.Containers {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.Image}, FindContainerStatus(c.Name, pod.Status.ContainerStatuses))
		if err != nil {
			return images, err
		}
		refs[ref] = append(refs[ref], c.Name)
	}
	for _, c := range pod.Spec.InitContainers {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.Image}, FindContainerStatus(c.Name, pod.Status.InitContainerStatuses))
		if err != nil {
			return images, err
		}
		refs[ref] = append(refs[ref], c.Name)
	}
	for _, c := range pod.Spec.EphemeralContainers {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.Image}, nil)
		if err != nil {
			return images, err
		}
		refs[ref] = append(refs[ref], c.Name)
	}

	for ref, containers := range refs {
		iu, ok := images[ref]
		if !ok {
			iu = kmapi.ImageInfo{
				Image:    ref,
				Lineages: nil,
				PullSecrets: &kmapi.PullSecrets{
					Namespace: pod.Namespace,
					Refs:      pod.Spec.ImagePullSecrets,
				},
			}
		}
		iu.Lineages = append(iu.Lineages, kmapi.Lineage{
			Chain:      lineage,
			Containers: containers,
		})
		images[ref] = iu
	}

	return images, nil
}

func CollectPullSecrets(pod *core.Pod, refs map[string]kmapi.PullSecrets) (map[string]kmapi.PullSecrets, error) {
	for _, c := range pod.Status.ContainerStatuses {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.ImageID}, FindContainerStatus(c.Name, pod.Status.ContainerStatuses))
		if err != nil {
			return refs, err
		}
		refs[ref] = kmapi.PullSecrets{
			Namespace: pod.Namespace,
			Refs:      pod.Spec.ImagePullSecrets,
		}
	}
	for _, c := range pod.Status.InitContainerStatuses {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.ImageID}, FindContainerStatus(c.Name, pod.Status.InitContainerStatuses))
		if err != nil {
			return refs, err
		}
		refs[ref] = kmapi.PullSecrets{
			Namespace: pod.Namespace,
			Refs:      pod.Spec.ImagePullSecrets,
		}
	}
	for _, c := range pod.Status.EphemeralContainerStatuses {
		ref, err := GetImageRef(Container{Name: c.Name, Image: c.ImageID}, nil)
		if err != nil {
			return refs, err
		}
		refs[ref] = kmapi.PullSecrets{
			Namespace: pod.Namespace,
			Refs:      pod.Spec.ImagePullSecrets,
		}
	}

	return refs, nil
}

func GetImageRef(c Container, status *core.ContainerStatus) (string, error) {
	var img string

	if strings.ContainsRune(c.Image, '@') {
		img = c.Image
	} else if strings.ContainsRune(status.Image, '@') {
		img = status.Image
	} else {
		// take the hash from status.ImageID and add to c.Image
		imageID := status.ImageID
		if strings.Contains(imageID, "://") {
			imageID = imageID[strings.Index(imageID, "://")+3:] // remove docker-pullable://
		}
		_, digest, ok := strings.Cut(imageID, "@")
		if !ok {
			img = c.Image
			// return "", fmt.Errorf("missing digest in pod %s container %s imageID %s", pod, status.Name, status.ImageID)
		} else {
			img = c.Image + "@" + digest
		}
	}

	ref, err := name.ParseReference(img)
	if err != nil {
		return "", errors.Wrapf(err, "ref=%s", img)
	}
	id := ref.Identifier()
	if strings.HasPrefix(id, "sha256:") {
		return ref.Context().String() + "@" + id, nil
	}
	return ref.Name(), nil
}

func FindContainerStatus(name string, statuses []core.ContainerStatus) *core.ContainerStatus {
	for i := range statuses {
		if statuses[i].Name == name {
			return &statuses[i]
		}
	}
	return nil
}

func DetectLineage(ctx context.Context, kc client.Client, obj client.Object) ([]kmapi.ObjectInfo, error) {
	return findLineage(ctx, kc, obj, nil)
}

func findLineage(ctx context.Context, kc client.Client, obj client.Object, result []kmapi.ObjectInfo) ([]kmapi.ObjectInfo, error) {
	ref := metav1.GetControllerOfNoCopy(obj)
	if ref != nil {
		var owner unstructured.Unstructured
		owner.SetAPIVersion(ref.APIVersion)
		owner.SetKind(ref.Kind)
		if err := kc.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: ref.Name}, &owner); client.IgnoreNotFound(err) != nil {
			return result, err
		} else if err == nil { // ignore not found error, owner might be already deleted
			var err error
			result, err = findLineage(ctx, kc, &owner, result)
			if err != nil {
				return result, err
			}
		}
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping, err := kc.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	result = append(result, kmapi.ObjectInfo{
		Resource: *kmapi.NewResourceID(mapping),
		Ref: kmapi.ObjectReference{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
	})
	return result, nil
}
