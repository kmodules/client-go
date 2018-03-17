package v1

import (
	"github.com/appscode/kutil/meta"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var json = jsoniter.ConfigFastest

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	return rbac.SchemeGroupVersion.WithKind(meta.GetKind(v))
}

func AssignTypeKind(v interface{}) error {
	_, err := conversion.EnforcePtr(v)
	if err != nil {
		return err
	}

	switch u := v.(type) {
	case *rbac.Role:
		u.APIVersion = rbac.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	case *rbac.RoleBinding:
		u.APIVersion = rbac.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	case *rbac.ClusterRole:
		u.APIVersion = rbac.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	case *rbac.ClusterRoleBinding:
		u.APIVersion = rbac.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	}
	return errors.New("unknown v1beta1 object type")
}
