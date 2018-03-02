package v1beta1

import (
	"github.com/appscode/kutil/meta"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	return apps.SchemeGroupVersion.WithKind(meta.GetKind(v))
}

func AssignTypeKind(v interface{}) error {
	_, err := conversion.EnforcePtr(v)
	if err != nil {
		return err
	}

	switch u := v.(type) {
	case *apps.StatefulSet:
		u.APIVersion = apps.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	case *apps.Deployment:
		u.APIVersion = apps.SchemeGroupVersion.String()
		u.Kind = meta.GetKind(v)
		return nil
	}
	return errors.New("unknown api object type")
}
