/*
Copyright The Kmodules Authors.

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

package v1beta1_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"kmodules.xyz/client-go/apiextensions/v1beta1"

	"github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/kube-openapi/pkg/common"
)

func TestConvertSchematoJsonProp(t *testing.T) {
	ref := new(string)
	*ref = "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"
	schema := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata",
			Ref:         v1beta1.OpenAPIRefCallBack(*ref),
		},
	}

	expected := extensionsobj.JSONSchemaProps{
		Description: "Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata",
		Ref:         ref,
	}
	var def map[string]common.OpenAPIDefinition
	props := v1beta1.SchemaPropsToJSONProps(&schema, def, false)

	if props.Description != expected.Description {
		t.Errorf("Description: expected %s, got %s", schema.Description, expected.Description)
	}

	if *props.Ref != schema.Ref.String() {
		t.Errorf("Ref: expected '%s', got '%s'", schema.Ref.String(), *props.Ref)
	}
}

func TestConvertFullSchematoJsonProp(t *testing.T) {
	schema := spec.Schema{SchemaProps: spec.SchemaProps{
		Description: "Describes an Alertmanager cluster.",
		Properties: map[string]spec.Schema{
			"kind": {
				SchemaProps: spec.SchemaProps{
					Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
					Type:        []string{"string"},
					Format:      "",
				},
			},
			"items": {
				SchemaProps: spec.SchemaProps{
					Description: "List of Alertmanagers",
					Type:        []string{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: v1beta1.OpenAPIRefCallBack("github.com/coreos/prometheus-operator/pkg/client/monitoring/v1.Alertmanager"),
							},
						},
					},
				},
			},
			"metadata": {
				SchemaProps: spec.SchemaProps{
					Description: "Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata",
					Ref:         v1beta1.OpenAPIRefCallBack("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
				},
			},
		},
	},
	}
	var def map[string]common.OpenAPIDefinition
	props := v1beta1.SchemaPropsToJSONProps(&schema, def, false)
	jsonBytes, err := json.MarshalIndent(props, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(jsonBytes)

}
