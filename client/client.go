package client

import (
	"context"
	"strings"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TransformFunc func(client.Object) client.Object

func CreateOrPatch(c client.Client, obj client.Object, transform TransformFunc, opts ...client.PatchOption) (client.Object, kutil.VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(context.TODO(), key, obj)
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating %+v %s/%s.", obj.GetObjectKind().GroupVersionKind(), key.Namespace, key.Name)

		createOpts := make([]client.CreateOption, 0, len(opts))
		for i := range opts {
			if opt, ok := opts[i].(client.CreateOption); ok {
				createOpts = append(createOpts, opt)
			}
		}
		obj = transform(obj.DeepCopyObject().(client.Object))
		err := c.Create(context.TODO(), obj, createOpts...)
		return obj, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	var patch client.Patch
	if isOfficialTypes(obj.GetObjectKind().GroupVersionKind().Group) {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}

	obj = transform(obj.DeepCopyObject().(client.Object))
	err = c.Patch(context.TODO(), obj, patch, opts...)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return obj, kutil.VerbPatched, nil
}

func isOfficialTypes(group string) bool {
	return !strings.ContainsRune(group, '.')
}
