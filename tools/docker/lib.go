package docker

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	manifestV2 "github.com/docker/distribution/manifest/schema2"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/golang/glog"
	reg "github.com/heroku/docker-registry-client/registry"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kubernetes/pkg/credentialprovider"
	_ "k8s.io/kubernetes/pkg/credentialprovider/aws"
	_ "k8s.io/kubernetes/pkg/credentialprovider/azure"
	_ "k8s.io/kubernetes/pkg/credentialprovider/gcp"
	// _ "k8s.io/kubernetes/pkg/credentialprovider/rancher" // enable in Kube 1.10
	"k8s.io/kubernetes/pkg/util/parsers"
)

var (
	defaultKeyring = credentialprovider.NewDockerKeyring()
)

func MakeDockerKeyring(pullSecrets []v1.Secret) (credentialprovider.DockerKeyring, error) {
	return credentialprovider.MakeDockerKeyring(pullSecrets, defaultKeyring)
}

// PullManifest pulls an image manifest (v2 or v1) from remote registry using the supplied secrets if necessary.
// ref: https://github.com/kubernetes/kubernetes/blob/release-1.9/pkg/kubelet/kuberuntime/kuberuntime_image.go#L29
func PullManifest(img string, keyring credentialprovider.DockerKeyring) (*reg.Registry, *dockertypes.AuthConfig, interface{}, error) {
	repoToPull, tag, digest, err := parsers.ParseImageName(img)
	if err != nil {
		return nil, nil, nil, err
	}
	parts := strings.SplitN(repoToPull, "/", 2)
	registry := parts[0]
	repo := parts[1]
	ref := tag
	if ref == "" {
		ref = digest
	}

	if strings.HasPrefix(registry, "docker.io") || strings.HasPrefix(registry, "index.docker.io") {
		registry = "registry-1.docker.io"
	}
	if !strings.HasPrefix(registry, "https://") && !strings.HasPrefix(registry, "http://") {
		registry = "https://" + registry
	}
	_, err = url.Parse(registry)
	if err != nil {
		return nil, nil, nil, err
	}

	creds, withCredentials := keyring.Lookup(repoToPull)
	if !withCredentials {
		glog.V(3).Infof("Pulling image %q without credentials", img)
		auth := &dockertypes.AuthConfig{ServerAddress: registry}
		hub, mf, err := pullManifest(repo, ref, auth)
		return hub, auth, mf, err
	}

	var pullErrs []error
	for _, currentCreds := range creds {
		authConfig := credentialprovider.LazyProvide(currentCreds)
		auth := &dockertypes.AuthConfig{
			Username:      authConfig.Username,
			Password:      authConfig.Password,
			Auth:          authConfig.Auth,
			ServerAddress: authConfig.ServerAddress,
		}
		if auth.ServerAddress == "" {
			auth.ServerAddress = registry
		}

		hub, mf, err := pullManifest(repo, ref, auth)
		// If there was no error, return success
		if err == nil {
			return hub, auth, mf, nil
		}
		pullErrs = append(pullErrs, err)
	}
	return nil, nil, nil, utilerrors.NewAggregate(pullErrs)
}

func pullManifest(repo, ref string, auth *dockertypes.AuthConfig) (*reg.Registry, interface{}, error) {
	hub := &reg.Registry{
		URL: auth.ServerAddress,
		Client: &http.Client{
			Transport: reg.WrapTransport(http.DefaultTransport, auth.ServerAddress, auth.Username, auth.Password),
		},
		Logf: reg.Log,
	}
	mf, err := hub.ManifestVx(repo, ref)
	return hub, mf, err
}

// GetLabels returns the labels of docker image. The image name should how it is presented to a Kubernetes container.
// If image is found it returns tuple {labels, err=nil}, otherwise it returns tuple {label=nil, err}
func GetLabels(hub *reg.Registry, img string, mf interface{}) (map[string]string, error) {
	repoToPull, _, _, err := parsers.ParseImageName(img)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(repoToPull, "/", 2)
	repo := parts[1]

	switch manifest := mf.(type) {
	case *manifestV2.DeserializedManifest:
		resp, err := hub.DownloadLayer(repo, manifest.Config.Digest)
		if err != nil {
			return nil, err
		}
		defer resp.Close()

		var cfg dockertypes.ImageInspect
		err = json.NewDecoder(resp).Decode(&cfg)
		if err != nil {
			return nil, err
		}
		return cfg.Config.Labels, nil
	}
	return nil, errors.New("image manifest must of v2 format")
}
