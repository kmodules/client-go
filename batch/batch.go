package batch

import (
	"context"

	v1 "kmodules.xyz/client-go/batch/v1"
	"kmodules.xyz/client-go/batch/v1beta1"
	"kmodules.xyz/client-go/discovery"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchCronJob(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*batchv1.CronJob) *batchv1.CronJob, opts metav1.PatchOptions) (*batchv1.CronJob, kutil.VerbType, error) {
	if ok, err := discovery.CheckAPIVersion(c.Discovery(), ">= 1.21"); err == nil && ok {
		return v1.CreateOrPatchCronJob(ctx, c, meta, transform, opts)
	}

	p, vt, err := v1beta1.CreateOrPatchCronJob(
		ctx,
		c,
		meta,
		func(in *batchv1beta1.CronJob) *batchv1beta1.CronJob {
			out := convert_v1_to_v1beta1(transform(convert_v1beta1_to_v1(in)))
			out.Status = in.Status
			return out
		},
		opts,
	)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return convert_v1beta1_to_v1(p), vt, nil
}

func DeleteCronJob(ctx context.Context, c kubernetes.Interface, meta types.NamespacedName) error {
	if ok, err := discovery.CheckAPIVersion(c.Discovery(), ">= 1.21"); err == nil && ok {
		return c.BatchV1().CronJobs(meta.Namespace).Delete(ctx, meta.Name, metav1.DeleteOptions{})
	}
	return c.BatchV1beta1().CronJobs(meta.Namespace).Delete(ctx, meta.Name, metav1.DeleteOptions{})
}

func convert_v1beta1_to_v1(in *batchv1beta1.CronJob) *batchv1.CronJob {
	return &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       in.Kind,
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: in.ObjectMeta,
		Spec: batchv1.CronJobSpec{
			Schedule:                in.Spec.Schedule,
			TimeZone:                in.Spec.TimeZone,
			StartingDeadlineSeconds: in.Spec.StartingDeadlineSeconds,
			ConcurrencyPolicy:       batchv1.ConcurrencyPolicy(in.Spec.ConcurrencyPolicy),
			Suspend:                 in.Spec.Suspend,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: in.Spec.JobTemplate.ObjectMeta,
				Spec:       in.Spec.JobTemplate.Spec,
			},
			SuccessfulJobsHistoryLimit: in.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     in.Spec.FailedJobsHistoryLimit,
		},
	}
}

func convert_v1_to_v1beta1(in *batchv1.CronJob) *batchv1beta1.CronJob {
	return &batchv1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       in.Kind,
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: in.ObjectMeta,
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                in.Spec.Schedule,
			TimeZone:                in.Spec.TimeZone,
			StartingDeadlineSeconds: in.Spec.StartingDeadlineSeconds,
			ConcurrencyPolicy:       batchv1beta1.ConcurrencyPolicy(in.Spec.ConcurrencyPolicy),
			Suspend:                 in.Spec.Suspend,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: in.Spec.JobTemplate.ObjectMeta,
				Spec:       in.Spec.JobTemplate.Spec,
			},
			SuccessfulJobsHistoryLimit: in.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     in.Spec.FailedJobsHistoryLimit,
		},
	}
}
