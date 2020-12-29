package portforward

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_getFirstSelectedPod(t *testing.T) {
	testCases := []struct {
		title           string
		resource        string
		resourceName    string
		sampleResource  runtime.Object
		selectedPodName string
		expectedErr     error
	}{
		{
			title:           "Targeted resource type is 'pods'",
			resource:        "pods",
			resourceName:    "foo-pod",
			selectedPodName: "foo-pod",
			sampleResource: newSamplePod(func(in *core.Pod) {
				in.Name = "another-foo-pod"
				in.Status.Phase = core.PodPending
			}),
		},
		{
			title:           "Targeted resource type is 'deployments'",
			resource:        "deployments",
			resourceName:    "foo-dpl",
			selectedPodName: "foo-pod",
			sampleResource:  newSampleDeployment(),
		},
		{
			title:           "Targeted resource type is 'daemonsets'",
			resource:        "daemonsets",
			resourceName:    "foo-dmn",
			selectedPodName: "foo-pod",
			sampleResource:  newSampleDaemonSet(),
		},
		{
			title:           "Targeted resource type is 'statefulsets'",
			resource:        "statefulsets",
			resourceName:    "foo-sts",
			selectedPodName: "foo-pod",
			sampleResource:  newSampleStatefulSet(),
		},
		{
			title:           "Targeted resource type is 'services'",
			resource:        "services",
			resourceName:    "foo-svc",
			selectedPodName: "foo-pod",
			sampleResource:  newSampleService(),
		},
		{
			title:          "Unknown resource type",
			resource:       "bar",
			resourceName:   "foo-bar",
			sampleResource: newSampleService(),
			expectedErr:    fmt.Errorf("unknown resource type: bar"),
		},
		{
			title:        "Selector doesn't match",
			resource:     "deployments",
			resourceName: "foo-dpl",
			sampleResource: newSampleDeployment(func(in *appsv1.Deployment) {
				in.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"select": "nothing",
					},
				}
			}),
			expectedErr: fmt.Errorf("no Pod found for deployments/foo-dpl"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			tunnel := NewTunnel(nil, nil, tc.resource, "default", tc.resourceName, 1234)
			fakeClient := fake.NewSimpleClientset(newSamplePod(), tc.sampleResource)

			pod, err := tunnel.getFirstSelectedPod(fakeClient)
			if err != nil {
				if tc.expectedErr != nil && tc.expectedErr.Error() == err.Error() {
					return
				}
				t.Error(err)
				return
			}
			if pod.Name != tc.selectedPodName {
				t.Errorf("Expect selected Pod name to be: %v Found: %v", tc.selectedPodName, pod.Name)
			}

		})
	}
}

func Test_translateRemotePort(t *testing.T) {
	testCases := []struct {
		title          string
		sampleResource runtime.Object
		remotePort     int
		expectedErr    error
		translatedPort int
	}{
		{
			title:          "Remote port does not exist in the Service",
			remotePort:     80,
			sampleResource: newSampleService(),
			expectedErr:    fmt.Errorf("remote port: 80 does not exist in Service: foo-svc"),
		},
		{
			title:          "TargetPort hasn't been specified in the Service",
			remotePort:     1234,
			sampleResource: newSampleService(),
			translatedPort: 1234,
		},
		{
			title:      "Port name has been specified in targetPort field of the Service",
			remotePort: 80,
			sampleResource: newSampleService(func(in *core.Service) {
				in.Spec.Ports = []core.ServicePort{
					{
						Name:       "foo",
						Port:       80,
						TargetPort: intstr.FromString("foo"),
					},
				}
			}),
			translatedPort: 1234,
		},
		{
			title:      "Port number has been specified in targetPort field of the Service",
			remotePort: 80,
			sampleResource: newSampleService(func(in *core.Service) {
				in.Spec.Ports = []core.ServicePort{
					{
						Name:       "foo",
						Port:       80,
						TargetPort: intstr.FromInt(1234),
					},
				}
			}),
			translatedPort: 1234,
		},
		{
			title:      "Remote port does not exist in the selected Pod",
			remotePort: 80,
			sampleResource: newSampleService(func(in *core.Service) {
				in.Spec.Ports = []core.ServicePort{
					{
						Name:       "foo",
						Port:       80,
						TargetPort: intstr.FromInt(80),
					},
				}
			}),
			expectedErr: fmt.Errorf("remote port: 80 does not match with any container port of the selected Pod: foo-pod"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			tunnel := NewTunnel(nil, nil, "services", "default", "foo-svc", tc.remotePort)
			fakeClient := fake.NewSimpleClientset(tc.sampleResource)

			err := tunnel.translateRemotePort(fakeClient, newSamplePod())
			if err != nil {
				if tc.expectedErr != nil && tc.expectedErr.Error() == err.Error() {
					return
				}
				t.Error(err)
				return
			}
			if tunnel.Remote != tc.translatedPort {
				t.Errorf("Expect translated port to be: %v Found: %v", tc.translatedPort, tunnel.Remote)
			}

		})
	}
}

func newSamplePod(transformFuncs ...func(in *core.Pod)) *core.Pod {
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-pod",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Ports: []core.ContainerPort{
						{
							Name:          "foo",
							ContainerPort: 1234,
						},
					},
				},
			},
		},
		Status: core.PodStatus{
			Phase: core.PodRunning,
		},
	}
	for _, fn := range transformFuncs {
		fn(pod)
	}
	return pod
}

func newSampleService(transformFuncs ...func(in *core.Service)) *core.Service {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-svc",
			Namespace: "default",
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"foo": "bar",
			},
			Ports: []core.ServicePort{
				{
					Name: "foo",
					Port: 1234,
				},
			},
		},
	}
	for _, fn := range transformFuncs {
		fn(svc)
	}
	return svc
}

func newSampleDeployment(transformFuncs ...func(in *appsv1.Deployment)) *appsv1.Deployment {
	dpl := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-dpl",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	for _, fn := range transformFuncs {
		fn(dpl)
	}
	return dpl
}

func newSampleDaemonSet(transformFuncs ...func(in *appsv1.DaemonSet)) *appsv1.DaemonSet {
	dmn := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-dmn",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	for _, fn := range transformFuncs {
		fn(dmn)
	}
	return dmn
}

func newSampleStatefulSet(transformFuncs ...func(in *appsv1.StatefulSet)) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-sts",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	for _, fn := range transformFuncs {
		fn(sts)
	}
	return sts
}
