package doctor

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ClusterInfo struct {
	Version      *VersionInfo                     `json:"version"`
	ClientConfig RestConfig                       `json:"clientConfig"`
	Capabilities Capabilities                     `json:"capabilities"`
	APIServers   APIServers                       `json:"apiServers"`
	AuthConfig   ExtensionApiserverAuthentication `json:"extension-apiserver-authentication"`
}

type VersionInfo struct {
	Minor      string `json:"minor"`
	Patch      string `json:"patch"`
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	Platform   string `json:"platform"`
}

type RestConfig struct {
	Host     string
	CABundle string `json:"caBundle"`
	Insecure bool   `json:"insecure"`
}

type Capabilities struct {
	APIVersion                 string `json:"apiVersion"`
	AggregateAPIServer         bool   `json:"aggregateAPIServer"`
	MutatingAdmissionWebhook   bool   `json:"mutatingAdmissionWebhook"`
	ValidatingAdmissionWebhook bool   `json:"validatingAdmissionWebhook"`
	PodSecurityPolicy          bool   `json:"podSecurityPolicy"`
	Initializers               bool   `json:"initializers"`
}

type APIServerConfig struct {
	PodName                   string
	NodeName                  string
	PodIP                     string
	HostIP                    string
	AdmissionControl          []string
	ClientCAData              string
	RequestheaderClientCAData string
	AllowPrivileged           bool
	AuthorizationMode         []string
	RuntimeConfig             RuntimeConfig
}

type RuntimeConfig struct {
	Enabled  []string
	Disabled []string
}

type APIServers []APIServerConfig

type ExtensionApiserverAuthentication struct {
	ClientCA      string
	RequestHeader *RequestHeaderConfig `json:"requestHeaderConfig"`
}

type RequestHeaderConfig struct {
	// UsernameHeaders are the headers to check (in order, case-insensitively) for an identity. The first header with a value wins.
	UsernameHeaders []string
	// GroupHeaders are the headers to check (case-insensitively) for a group names.  All values will be used.
	GroupHeaders []string
	// ExtraHeaderPrefixes are the head prefixes to check (case-insentively) for filling in
	// the user.Info.Extra.  All values of all matching headers will be added.
	ExtraHeaderPrefixes []string
	// ClientCA points to CA bundle file which is used verify the identity of the front proxy
	ClientCA string
	// AllowedClientNames is a list of common names that may be presented by the authenticating front proxy.  Empty means: accept any.
	AllowedClientNames []string
}

func (c ClusterInfo) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (c ClusterInfo) Validate() error {
	var errs []error

	{
		if c.ClientConfig.Insecure {
			errs = append(errs, errors.New("Admission webhooks can't be used when kube apiserver is accesible without verifying its TLS certificate (insecure-skip-tls-verify : true)."))
		} else {
			if c.AuthConfig.ClientCA == "" {
				errs = append(errs, errors.Errorf(`"%s/%s" configmap is missing "client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
			} else if c.ClientConfig.CABundle != c.AuthConfig.ClientCA {
				errs = append(errs, errors.Errorf(`"%s/%s" configmap has mismatched "client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
			}

			for _, pod := range c.APIServers {
				if pod.ClientCAData != c.ClientConfig.CABundle {
					errs = append(errs, errors.Errorf(`pod "%s"" has mismatched "client-ca-file".`, pod.PodName))
				}
			}
		}
	}
	{
		if len(c.APIServers) == 0 {
			errs = append(errs, errors.New(`failed to detect kube apiservers. Please file a bug at: https://github.com/appscode/kutil/issues/new .`))
		}
	}
	{
		if c.AuthConfig.RequestHeader == nil {
			errs = append(errs, errors.Errorf(`"%s/%s" configmap is missing "requestheader-client-ca-file" key.`, authenticationConfigMapNamespace, authenticationConfigMapName))
		}
		for _, pod := range c.APIServers {
			if pod.RequestheaderClientCAData != c.AuthConfig.RequestHeader.ClientCA {
				errs = append(errs, errors.Errorf(`pod "%s"" has mismatched "requestheader-client-ca-file".`, pod.PodName))
			}
		}
	}
	{
		for _, pod := range c.APIServers {
			modes := sets.NewString(pod.AuthorizationMode...)
			if !modes.Has("RBAC") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable RBAC authorization mode.`, pod.PodName))
			}
		}
	}
	{
		for _, pod := range c.APIServers {
			adms := sets.NewString(pod.AdmissionControl...)
			if !adms.Has("MutatingAdmissionWebhook") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable MutatingAdmissionWebhook admission controller.`, pod.PodName))
			}
			if !adms.Has("ValidatingAdmissionWebhook") {
				errs = append(errs, errors.Errorf(`pod "%s"" does not enable ValidatingAdmissionWebhook admission controller.`, pod.PodName))
			}
		}
	}
	return utilerrors.NewAggregate(errs)
}

func (servers APIServers) UsesAdmissionControl(name string) (bool, error) {
	e := 0
	for _, s := range servers {
		adms := sets.NewString(s.AdmissionControl...)
		if adms.Has(name) {
			e++
		}
	}

	switch {
	case e == 0:
		return false, nil
	case e == len(servers):
		return true, nil
	default:
		return false, errors.Errorf("admission control %s is enabled in %s api server, expected %d", name, e, len(servers))
	}
}
