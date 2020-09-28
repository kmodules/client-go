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

package conditions

import (
	"context"
	"encoding/json"

	kmapi "kmodules.xyz/client-go/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type DynamicOptions struct {
	Client    dynamic.Interface
	GVR       schema.GroupVersionResource
	Kind      string
	Name      string
	Namespace string
}

func (do *DynamicOptions) HasCondition(condType string) (bool, error) {
	_, conditions, err := do.ReadConditions()
	if err != nil {
		return false, err
	}
	return kmapi.HasCondition(conditions, condType), nil
}

func (do *DynamicOptions) GetCondition(condType string) (int, *kmapi.Condition, error) {
	_, conditions, err := do.ReadConditions()
	if err != nil {
		return -1, nil, err
	}
	idx, cond := kmapi.GetCondition(conditions, condType)
	return idx, cond, nil
}

func (do *DynamicOptions) SetCondition(newCond kmapi.Condition) error {
	res, conditions, err := do.ReadConditions()
	if err != nil {
		return err
	}
	conditions = kmapi.SetCondition(conditions, newCond)

	unstrConds := make([]interface{}, len(conditions))
	for i := range conditions {
		cond, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&conditions[i])
		if err != nil {
			return err
		}
		unstrConds[i] = cond
	}
	err = unstructured.SetNestedField(res.Object, unstrConds, "status", "conditions")
	if err != nil {
		return err
	}
	_, err = do.Client.Resource(do.GVR).Namespace(do.Namespace).UpdateStatus(context.TODO(), res, metav1.UpdateOptions{})
	return err
}

func (do *DynamicOptions) RemoveCondition(condType string) error {
	res, conditions, err := do.ReadConditions()
	if err != nil {
		return err
	}
	conditions = kmapi.RemoveCondition(conditions, condType)

	unstrConds := make([]interface{}, len(conditions))
	for i := range conditions {
		cond, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&conditions[i])
		if err != nil {
			return err
		}
		unstrConds[i] = cond
	}
	err = unstructured.SetNestedField(res.Object, unstrConds, "status", "conditions")
	if err != nil {
		return err
	}
	_, err = do.Client.Resource(do.GVR).Namespace(do.Namespace).UpdateStatus(context.TODO(), res, metav1.UpdateOptions{})
	return err
}

func (do *DynamicOptions) IsConditionTrue(condType string) (bool, error) {
	_, conditions, err := do.ReadConditions()
	if err != nil {
		return false, err
	}
	return kmapi.IsConditionTrue(conditions, condType), nil
}

func (do *DynamicOptions) ReadConditions() (*unstructured.Unstructured, []kmapi.Condition, error) {
	resp, err := do.Client.Resource(do.GVR).Namespace(do.Namespace).Get(context.TODO(), do.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	unstrConds, _, err := unstructured.NestedSlice(resp.Object, "status", "conditions")
	if err != nil {
		return nil, nil, err
	}
	var conditions []kmapi.Condition
	data, err := json.Marshal(unstrConds)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(data, &conditions)
	return resp, conditions, err
}
