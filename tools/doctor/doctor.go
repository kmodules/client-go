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

//nolint:goconst
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
		info.Capabilities.APIVersion = info.Version.Minor
	}
	{
		info.Capabilities.AggregateAPIServer = info.ExtensionServerConfig.RequestHeader != nil
	}
	{
		status, err := info.APIServers.AdmissionControl("MutatingAdmissionWebhook")
		if err != nil {
			return nil, err
		}
		if info.ClientConfig.Insecure {
			info.Capabilities.MutatingAdmissionWebhook = "false"
		} else {
			info.Capabilities.MutatingAdmissionWebhook = status
		}
	}
	{
		status, err := info.APIServers.AdmissionControl("ValidatingAdmissionWebhook")
		if err != nil {
			return nil, err
		}
		if info.ClientConfig.Insecure {
			info.Capabilities.ValidatingAdmissionWebhook = "false"
		} else {
			info.Capabilities.ValidatingAdmissionWebhook = status
		}
	}
	{
		status, err := info.APIServers.AdmissionControl("PodSecurityPolicy")
		if err != nil {
			return nil, err
		}
		info.Capabilities.PodSecurityPolicy = status

	}
	{
		status, err := info.APIServers.AdmissionControl("Initializers")
		if err != nil {
			return nil, err
		}
		info.Capabilities.Initializers = status
	}
	{
		status, err := info.APIServers.FeatureGate("CustomResourceSubresources")
		if err != nil {
			return nil, err
		}
		info.Capabilities.CustomResourceSubresources = status
	}

	return &info, nil
}
