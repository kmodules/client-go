/*
Copyright AppsCode Inc. and Contributors

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

package v1beta1

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	reg "k8s.io/api/admissionregistration/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchValidatingWebhookConfiguration(ctx context.Context, c kubernetes.Interface, name string, transform func(*reg.ValidatingWebhookConfiguration) *reg.ValidatingWebhookConfiguration, opts metav1.PatchOptions) (*reg.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	cur, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating ValidatingWebhookConfiguration %s.", name)
		out, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(ctx, transform(&reg.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ValidatingWebhookConfiguration",
				APIVersion: reg.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchValidatingWebhookConfiguration(ctx, c, cur, transform, opts)
}

func PatchValidatingWebhookConfiguration(ctx context.Context, c kubernetes.Interface, cur *reg.ValidatingWebhookConfiguration, transform func(*reg.ValidatingWebhookConfiguration) *reg.ValidatingWebhookConfiguration, opts metav1.PatchOptions) (*reg.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	return PatchValidatingWebhookConfigurationObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchValidatingWebhookConfigurationObject(ctx context.Context, c kubernetes.Interface, cur, mod *reg.ValidatingWebhookConfiguration, opts metav1.PatchOptions) (*reg.ValidatingWebhookConfiguration, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, reg.ValidatingWebhookConfiguration{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching ValidatingWebhookConfiguration %s with %s.", cur.Name, string(patch))
	out, err := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateValidatingWebhookConfiguration(ctx context.Context, c kubernetes.Interface, name string, transform func(*reg.ValidatingWebhookConfiguration) *reg.ValidatingWebhookConfiguration, opts metav1.UpdateOptions) (result *reg.ValidatingWebhookConfiguration, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update ValidatingWebhookConfiguration %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update ValidatingWebhookConfiguration %s after %d attempts due to %v", name, attempt, err)
	}
	return
}

func UpdateValidatingWebhookCABundle(config *rest.Config, webhookConfigName string, extraConditions ...watchtools.ConditionFunc) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, kutil.ReadinessTimeout)
	defer cancel()

	err := rest.LoadTLSFiles(config)
	if err != nil {
		return err
	}

	kc := kubernetes.NewForConfigOrDie(config)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, webhookConfigName).String()
			return kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, webhookConfigName).String()
			return kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Watch(ctx, options)
		},
	}

	var conditions = append([]watchtools.ConditionFunc{
		func(event watch.Event) (bool, error) {
			switch event.Type {
			case watch.Deleted:
				return false, nil
			case watch.Error:
				return false, errors.New("error watching")
			case watch.Added, watch.Modified:
				cur := event.Object.(*reg.ValidatingWebhookConfiguration)
				_, _, err := PatchValidatingWebhookConfiguration(context.TODO(), kc, cur, func(in *reg.ValidatingWebhookConfiguration) *reg.ValidatingWebhookConfiguration {
					for i := range in.Webhooks {
						in.Webhooks[i].ClientConfig.CABundle = config.CAData
					}
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					glog.Warning(err)
				}
				return err == nil, err
			default:
				return false, fmt.Errorf("unexpected event type: %v", event.Type)
			}
		},
	}, extraConditions...)

	_, err = watchtools.UntilWithSync(ctx,
		lw,
		&reg.ValidatingWebhookConfiguration{},
		nil,
		conditions...,
	)
	return err
}

func SyncValidatingWebhookCABundle(config *rest.Config, webhookConfigName string) (cancel context.CancelFunc, err error) {
	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	err = rest.LoadTLSFiles(config)
	if err != nil {
		return
	}

	kc := kubernetes.NewForConfigOrDie(config)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, webhookConfigName).String()
			return kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fields.OneTermEqualSelector(kutil.ObjectNameField, webhookConfigName).String()
			return kc.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Watch(ctx, options)
		},
	}

	go func() {
		_, err := watchtools.UntilWithSync(
			ctx,
			lw,
			&reg.ValidatingWebhookConfiguration{},
			nil,
			func(event watch.Event) (bool, error) {
				switch event.Type {
				case watch.Deleted:
					return false, nil
				case watch.Error:
					return false, errors.New("error watching")
				case watch.Added, watch.Modified:
					cur := event.Object.(*reg.ValidatingWebhookConfiguration)
					_, _, err := PatchValidatingWebhookConfiguration(context.TODO(), kc, cur, func(in *reg.ValidatingWebhookConfiguration) *reg.ValidatingWebhookConfiguration {
						for i := range in.Webhooks {
							in.Webhooks[i].ClientConfig.CABundle = config.CAData
						}
						return in
					}, metav1.PatchOptions{})
					if err != nil {
						glog.Warning(err)
					}
					return false, nil // continue
				default:
					return false, fmt.Errorf("unexpected event type: %v", event.Type)
				}
			})
		utilruntime.Must(err)
	}()
	return
}
