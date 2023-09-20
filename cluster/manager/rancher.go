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

package manager

import (
	"context"
	"errors"

	"kmodules.xyz/client-go/apis/management/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const LabelKeyRancherProjectId = "field.cattle.io/projectId"

func IsRancherManaged(mapper meta.RESTMapper) bool {
	if _, err := mapper.RESTMappings(schema.GroupKind{
		Group: "management.cattle.io",
		Kind:  "Cluster",
	}); err == nil {
		return true
	}
	return false
}

func IsRancherSystemResource(kc client.Client, key types.NamespacedName) (bool, error) {
	if !IsRancherManaged(kc.RESTMapper()) {
		return false, errors.New("not a Rancher managed cluster")
	}

	if key.Namespace == metav1.NamespaceSystem {
		return true, nil
	}

	var ns core.Namespace
	err := kc.Get(context.TODO(), client.ObjectKey{Name: key.Namespace}, &ns)
	if err != nil {
		return false, err
	}
	projectId, exists := ns.Labels[LabelKeyRancherProjectId]
	if !exists {
		return false, nil
	}

	var sysNS core.Namespace
	err = kc.Get(context.TODO(), client.ObjectKey{Name: metav1.NamespaceSystem}, &sysNS)
	if err != nil {
		return false, err
	}

	sysProjectId, exists := ns.Labels[LabelKeyRancherProjectId]
	if !exists {
		return false, nil
	}
	return projectId == sysProjectId, nil
}

func ListRancherProjects(kc client.Client) ([]v1alpha1.Project, error) {
	var list core.NamespaceList
	err := kc.List(context.TODO(), &list)
	if meta.IsNoMatchError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	projects := map[string]v1alpha1.Project{}
	for _, ns := range list.Items {
		projectId, exists := ns.Labels[LabelKeyRancherProjectId]
		if !exists {
			continue
		}

		project, exists := projects[projectId]
		if !exists {
			project = v1alpha1.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name: projectId,
				},
				Spec: v1alpha1.ProjectSpec{
					Type:       v1alpha1.ProjectUser,
					Namespaces: nil,
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							LabelKeyRancherProjectId: projectId,
						},
					},
					// Quota: core.ResourceRequirements{},
				},
			}
		}
		if ns.Name == metav1.NamespaceDefault {
			project.Spec.Type = v1alpha1.ProjectDefault
		} else if ns.Name == metav1.NamespaceSystem {
			project.Spec.Type = v1alpha1.ProjectSystem
		}
		project.Spec.Namespaces = append(project.Spec.Namespaces, ns.Name)

		projects[projectId] = project
	}

	result := make([]v1alpha1.Project, 0, len(projects))
	for _, p := range projects {
		result = append(result, p)
	}
	return result, nil
}

func GetRancherProject(kc client.Client, name string) (*v1alpha1.Project, error) {
	var list core.NamespaceList
	err := kc.List(context.TODO(), &list, client.MatchingLabels{
		LabelKeyRancherProjectId: name,
	})
	if err != nil {
		return nil, err
	}

	project := v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ProjectSpec{
			Type:       v1alpha1.ProjectUser,
			Namespaces: nil,
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelKeyRancherProjectId: name,
				},
			},
			// Quota: core.ResourceRequirements{},
		},
	}
	for _, ns := range list.Items {
		if ns.Name == metav1.NamespaceDefault {
			project.Spec.Type = v1alpha1.ProjectDefault
		} else if ns.Name == metav1.NamespaceSystem {
			project.Spec.Type = v1alpha1.ProjectSystem
		}
		project.Spec.Namespaces = append(project.Spec.Namespaces, ns.Name)
	}

	return &project, nil
}
