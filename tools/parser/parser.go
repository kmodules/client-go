package parser

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ylib "k8s.io/apimachinery/pkg/util/yaml"
)

type ResourceFn func(obj *unstructured.Unstructured) error

func ProcessResources(data []byte, fn ResourceFn) error {
	reader := ylib.NewYAMLOrJSONDecoder(bytes.NewReader(data), 2048)
	for {
		var obj unstructured.Unstructured
		err := reader.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if obj.IsList() {
			if err := obj.EachListItem(func(item runtime.Object) error {
				return fn(item.(*unstructured.Unstructured))
			}); err != nil {
				return err
			}
		} else {
			if err := fn(&obj); err != nil {
				return err
			}
		}
	}
	return nil
}

func ProcessDir(dir string, fn ResourceFn) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fmt.Println(">>> ", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		return ProcessResources(data, fn)
	})
}

func ListResources(data []byte) ([]*unstructured.Unstructured, error) {
	var resources []*unstructured.Unstructured

	err := ProcessResources(data, func(obj *unstructured.Unstructured) error {
		if obj.GetNamespace() == "" {
			obj.SetNamespace(core.NamespaceDefault)
		}
		resources = append(resources, obj)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(resources, func(i, j int) bool {
		if resources[i].GetAPIVersion() == resources[j].GetAPIVersion() {
			return resources[i].GetKind() < resources[j].GetKind()
		}
		return resources[i].GetAPIVersion() < resources[j].GetAPIVersion()
	})

	return resources, nil
}

var empty = struct{}{}

func ExtractComponents(data []byte) (map[metav1.GroupKind]struct{}, map[string]string, error) {
	components := map[metav1.GroupKind]struct{}{}
	commonLabels := map[string]string{}
	init := false

	err := ProcessResources(data, func(obj *unstructured.Unstructured) error {
		gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
		if err != nil {
			return err
		}
		components[metav1.GroupKind{Group: gv.Group, Kind: obj.GetKind()}] = empty

		if !init {
			commonLabels = obj.GetLabels()
			init = true
		} else {
			for k, v := range obj.GetLabels() {
				if existing, found := commonLabels[k]; found && existing != v {
					delete(commonLabels, k)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return components, commonLabels, err
}
