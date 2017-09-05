package v2alpha1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime/schema"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
)

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	switch v.(type) {
	case *batchv2alpha1.CronJob:
		return batchv2alpha1.SchemeGroupVersion.WithKind("CronJob")
	default:
		return schema.GroupVersionKind{}
	}
}

func AssignTypeKind(v interface{}) error {
	switch u := v.(type) {
	case *batchv2alpha1.CronJob:
		u.APIVersion = batchv2alpha1.SchemeGroupVersion.String()
		u.Kind = "CronJob"
		return nil
	}
	return errors.New("Unknown api object type")
}
