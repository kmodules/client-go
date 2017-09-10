package v1alpha1

import (
	"errors"

	"github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	switch v.(type) {
	case *v1alpha1.Postgres:
		return v1alpha1.SchemeGroupVersion.WithKind("Postgres")
	case *v1alpha1.Elasticsearch:
		return v1alpha1.SchemeGroupVersion.WithKind("Elasticsearch")
	case *v1alpha1.Snapshot:
		return v1alpha1.SchemeGroupVersion.WithKind("Snapshot")
	case *v1alpha1.DormantDatabase:
		return v1alpha1.SchemeGroupVersion.WithKind("DormantDatabase")
	default:
		return schema.GroupVersionKind{}
	}
}

func AssignTypeKind(v interface{}) error {
	switch u := v.(type) {
	case *v1alpha1.Postgres:
		u.APIVersion = v1alpha1.SchemeGroupVersion.String()
		u.Kind = "Postgres"
		return nil
	case *v1alpha1.Elasticsearch:
		u.APIVersion = v1alpha1.SchemeGroupVersion.String()
		u.Kind = "Elasticsearch"
		return nil
	case *v1alpha1.Snapshot:
		u.APIVersion = v1alpha1.SchemeGroupVersion.String()
		u.Kind = "Snapshot"
		return nil
	case *v1alpha1.DormantDatabase:
		u.APIVersion = v1alpha1.SchemeGroupVersion.String()
		u.Kind = "DormantDatabase"
		return nil
	}
	return errors.New("Unknown api object type")
}
