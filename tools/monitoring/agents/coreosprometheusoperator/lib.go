package coreosprometheusoperator

import (
	"errors"
	"reflect"

	"github.com/appscode/kutil/tools/monitoring/api"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	ecs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// PrometheusCoreosOperator creates `ServiceMonitor` so that CoreOS Prometheus operator can generate necessary config for Prometheus.
type PrometheusCoreosOperator struct {
	k8sClient  kubernetes.Interface
	promClient prom.MonitoringV1Interface
	extClient  ecs.ApiextensionsV1beta1Interface
}

func New(k8sClient kubernetes.Interface, extClient ecs.ApiextensionsV1beta1Interface, promClient prom.MonitoringV1Interface) api.Agent {
	return &PrometheusCoreosOperator{
		k8sClient:  k8sClient,
		extClient:  extClient,
		promClient: promClient,
	}
}

func (agent *PrometheusCoreosOperator) Add(sp api.StatsAccessor, spec *api.AgentSpec) error {
	return agent.Update(sp, spec)
}

func (agent *PrometheusCoreosOperator) Update(sp api.StatsAccessor, new *api.AgentSpec) error {
	if !agent.SupportsCoreOSOperator() {
		return errors.New("cluster does not support CoreOS Prometheus operator")
	}
	return agent.ensureServiceMonitor(sp, new)
}

func (agent *PrometheusCoreosOperator) Delete(sp api.StatsAccessor, spec *api.AgentSpec) error {
	if !agent.SupportsCoreOSOperator() {
		return errors.New("cluster does not support CoreOS Prometheus operator")
	}
	if err := agent.promClient.ServiceMonitors(spec.Prometheus.Namespace).Delete(sp.ServiceMonitorName(), nil); !kerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (agent *PrometheusCoreosOperator) SupportsCoreOSOperator() bool {
	_, err := agent.extClient.CustomResourceDefinitions().Get(prom.PrometheusName+"."+prom.Group, metav1.GetOptions{})
	if err != nil {
		return false
	}
	_, err = agent.extClient.CustomResourceDefinitions().Get(prom.ServiceMonitorName+"."+prom.Group, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (agent *PrometheusCoreosOperator) ensureServiceMonitor(sp api.StatsAccessor, new *api.AgentSpec) error {
	old, err := agent.promClient.ServiceMonitors(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			"name": sp.ServiceMonitorName(),
		}.String(),
	})

	oldItems := old.(*prom.ServiceMonitorList)

	for _, item := range oldItems.Items {
		if item != nil && (new == nil || item.Namespace != new.Prometheus.Namespace) {
			err := agent.promClient.ServiceMonitors(item.Namespace).Delete(sp.ServiceMonitorName(), nil)
			if err != nil && !kerr.IsNotFound(err) {
				return err
			}
			if new == nil {
				return nil
			}
		}
	}

	actual, err := agent.promClient.ServiceMonitors(new.Prometheus.Namespace).Get(sp.ServiceMonitorName(), metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return agent.createServiceMonitor(sp, new)
	} else if err != nil {
		return err
	}

	update := false
	if !reflect.DeepEqual(actual.Labels, new.Prometheus.Labels) {
		update = true
	}

	if !update {
		for _, e := range actual.Spec.Endpoints {
			if e.Interval != new.Prometheus.Interval {
				update = true
				break
			}
		}
	}

	if update {
		svc, err := agent.k8sClient.CoreV1().Services(sp.GetNamespace()).Get(sp.ServiceName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		var labels map[string]string
		labels = svc.Labels
		labels["name"] = sp.ServiceMonitorName()

		actual.Labels = new.Prometheus.Labels
		actual.Spec.Selector = metav1.LabelSelector{
			MatchLabels: labels,
		}
		actual.Spec.NamespaceSelector = prom.NamespaceSelector{
			MatchNames: []string{sp.GetNamespace()},
		}
		for i := range actual.Spec.Endpoints {
			actual.Spec.Endpoints[i].Interval = new.Prometheus.Interval
		}
		_, err = agent.promClient.ServiceMonitors(new.Prometheus.Namespace).Update(actual)
		return err
	}

	return nil
}

func (agent *PrometheusCoreosOperator) createServiceMonitor(sp api.StatsAccessor, spec *api.AgentSpec) error {
	svc, err := agent.k8sClient.CoreV1().Services(sp.GetNamespace()).Get(sp.ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	var portName string
	for _, p := range svc.Spec.Ports {
		if p.Port == spec.Prometheus.Port {
			portName = p.Name
		}
	}
	if portName == "" {
		return errors.New("no port found in stats service")
	}

	var labels map[string]string
	labels = svc.Labels
	labels["name"] = sp.ServiceMonitorName()

	sm := &prom.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sp.ServiceMonitorName(),
			Namespace: spec.Prometheus.Namespace,
			Labels:    spec.Prometheus.Labels,
		},
		Spec: prom.ServiceMonitorSpec{
			NamespaceSelector: prom.NamespaceSelector{
				MatchNames: []string{sp.GetNamespace()},
			},
			Endpoints: []prom.Endpoint{
				{
					Port:     portName,
					Interval: spec.Prometheus.Interval,
					Path:     sp.Path(),
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
	if _, err := agent.promClient.ServiceMonitors(spec.Prometheus.Namespace).Create(sm); !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}
