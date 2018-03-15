package v1

import (
	"reflect"
	"testing"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoveOwnerReference(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       "Deployment",
				Name:       "dep-0",
				APIVersion: "v1",
				UID:        "0",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-1",
				APIVersion: "v1",
				UID:        "1",
			},
			{
				Kind:       "Deployment",
				Name:       "dep-2",
				APIVersion: "v1",
				UID:        "2",
			},
		},
	}

	ref := core.ObjectReference{
		Kind:       "Deployment",
		Name:       "dep-3",
		APIVersion: "v1",
		UID:        "3",
	}

	meta := RemoveOwnerReference(objectMeta, &ref)
	if !reflect.DeepEqual(meta, objectMeta) {
		t.Errorf("Remove of owner Reference is not successful, expected: %v. But Got: %v", objectMeta, meta)
	}

	appendedMeta := objectMeta
	appendedMeta.OwnerReferences = append(meta.OwnerReferences, metav1.OwnerReference{
		UID:        ref.UID,
		APIVersion: ref.APIVersion,
		Name:       ref.Name,
		Kind:       ref.Kind,
	})

	meta = RemoveOwnerReference(appendedMeta, &ref)
	if !reflect.DeepEqual(meta, objectMeta) {
		t.Errorf("Remove of owner Reference is not successful, expected: %v. But Got: %v", objectMeta, meta)
	}

}
