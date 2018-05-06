package doctor

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

type Doctor struct {
	config *rest.Config
	kc     kubernetes.Interface
}

func New(config *rest.Config) (*Doctor, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Doctor{config, client}, nil
}

func (d *Doctor) GetClusterInfo() (*ClusterInfo, error) {
	var info ClusterInfo
	var err error

	err = d.extractKubeCA(&info)
	if err != nil {
		return nil, err
	}

	err = d.extractVersion(&info)
	if err != nil {
		return nil, err
	}

	err = d.extractExtendedAPIServerInfo(&info)
	if err != nil {
		return nil, err
	}

	err = d.extractMasterArgs(&info)
	if err != nil {
		return nil, err
	}

	{
		info.Capabilities.AggregateAPIServer = info.AuthConfig.RequestHeader != nil
	}
	{
		enabled, err := info.APIServers.UsesAdmissionControl("MutatingAdmissionWebhook")
		if err != nil {
			return nil, err
		}
		info.Capabilities.MutatingAdmissionWebhook = !info.ClientConfig.Insecure && enabled
	}
	{
		enabled, err := info.APIServers.UsesAdmissionControl("ValidatingAdmissionWebhook")
		if err != nil {
			return nil, err
		}
		info.Capabilities.MutatingAdmissionWebhook = !info.ClientConfig.Insecure && enabled

	}
	{
		enabled, err := info.APIServers.UsesAdmissionControl("PodSecurityPolicy")
		if err != nil {
			return nil, err
		}
		info.Capabilities.PodSecurityPolicy = !info.ClientConfig.Insecure && enabled

	}
	{
		enabled, err := info.APIServers.UsesAdmissionControl("Initializers")
		if err != nil {
			return nil, err
		}
		info.Capabilities.Initializers = !info.ClientConfig.Insecure && enabled

	}

	return &info, nil
}
