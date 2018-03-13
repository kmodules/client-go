package converter

import (
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	"k8s.io/kubernetes/pkg/api/legacyscheme"

	apps_v1 "k8s.io/api/apps/v1"
	apps_v1beta1 "k8s.io/api/apps/v1beta1"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	batch_v1 "k8s.io/api/batch/v1"
	batch_v1beta1 "k8s.io/api/batch/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	extensions_v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	Version_v1                = "v1"
	Version_v1beta1           = "v1beta1"
	Version_v1beta2           = "v1beta2"
	KindDeployment            = "Deployment"
	KindDaemonSet             = "DaemonSet"
	KindReplicaSet            = "ReplicaSet"
	KindStatefulSet           = "StatefulSet"
	KindReplicationController = "ReplicationController"
	KindJob                   = "Job"
	KindCronJob               = "CronJob"
	KindPod                   = "Pod"
)

func Convert(sgvk schema.GroupVersionKind, in, out interface{}) error {
	var sourceObj, internalObj interface{}

	switch sgvk.Group {
	case apps.GroupName:
		switch sgvk.Version {
		case Version_v1:
			switch sgvk.Kind {
			case KindDeployment:
			case KindDaemonSet:
			case KindReplicaSet:
			case KindStatefulSet:
				sourceObj = &apps_v1.StatefulSet{}
				internalObj = &apps.StatefulSet{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		case Version_v1beta1:
			switch sgvk.Kind {
			case KindDeployment:
			case KindStatefulSet:
				sourceObj = &apps_v1beta1.StatefulSet{}
				internalObj = &apps.StatefulSet{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		case Version_v1beta2:
			switch sgvk.Kind {
			case KindDeployment:
			case KindDaemonSet:
			case KindReplicaSet:
			case KindStatefulSet:
				sourceObj = &apps_v1beta2.StatefulSet{}
				internalObj = &apps.StatefulSet{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		default:
			return fmt.Errorf("resource version unknown")
		}
	case batch.GroupName:
		switch sgvk.Version {
		case Version_v1:
			switch sgvk.Kind {
			case KindJob:
				sourceObj = &batch_v1.Job{}
				internalObj = &batch.Job{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		case Version_v1beta1:
			switch sgvk.Kind {
			case KindCronJob:
				sourceObj = &batch_v1beta1.CronJob{}
				internalObj = &batch.CronJob{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		default:
			return fmt.Errorf("resource version unknown")
		}
	case core.GroupName:
		switch sgvk.Version {
		case Version_v1:
			switch sgvk.Kind {
			case KindReplicationController:
				sourceObj = &core_v1.ReplicationController{}
				internalObj = &core.ReplicationController{}
			case KindPod:
				sourceObj = &core_v1.Pod{}
				internalObj = &core.Pod{}
			default:
				fmt.Errorf("resource kind unknown")
			}
		default:
			return fmt.Errorf("resource version unknown")
		}
	case extensions.GroupName:
		switch sgvk.Version {
		case Version_v1beta1:
			switch sgvk.Kind {
			case KindDeployment:
				sourceObj = &extensions_v1beta1.Deployment{}
				internalObj = &extensions.Deployment{}
			case KindDaemonSet:
				sourceObj = &extensions_v1beta1.DaemonSet{}
				internalObj = &extensions.DaemonSet{}
			case KindReplicaSet:
				sourceObj = &extensions_v1beta1.ReplicaSet{}
				internalObj = &extensions.ReplicaSet{}
			default:
				return fmt.Errorf("resource kind unknown")
			}
		default:
			return fmt.Errorf("resource version unknown")
		}
	default:
		return fmt.Errorf("resource group unknown")
	}

	if reflect.TypeOf(in).Kind() == reflect.Ptr { //in is not in []byte type. it is object
		sourceObj = in
	} else { // in is []byte type
		err := json.Unmarshal(in.([]byte), sourceObj)
		if err != nil {
			return err
		}
	}

	// convert to internal version
	err := legacyscheme.Scheme.Convert(sourceObj, internalObj, nil)
	if err != nil {
		return err
	}

	// now convert this internal version to required version
	err = legacyscheme.Scheme.Convert(internalObj, out, nil)
	if err != nil {
		return nil
	}
	return nil
}
