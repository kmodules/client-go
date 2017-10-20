package v1beta1

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/appscode/kutil"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	return extensions.SchemeGroupVersion.WithKind(kutil.GetKind(v))
}

func AssignTypeKind(v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("%v must be a pointer", v)
	}

	switch u := v.(type) {
	case *extensions.Ingress:
		u.APIVersion = extensions.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	case *extensions.DaemonSet:
		u.APIVersion = extensions.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	case *extensions.ReplicaSet:
		u.APIVersion = extensions.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	case *extensions.Deployment:
		u.APIVersion = extensions.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	case *extensions.ThirdPartyResource:
		u.APIVersion = extensions.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	}
	return errors.New("unknown api object type")
}

func IsOwnedByDeployment(rs *extensions.ReplicaSet) bool {
	for _, ref := range rs.OwnerReferences {
		if ref.Kind == "Deployment" && ref.Name != "" {
			return true
		}
	}
	return false
}
