package discovery

import (
	"fmt"

	version "github.com/appscode/go-version"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

func GetVersion(client discovery.DiscoveryInterface) (string, error) {
	info, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	gv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return "", err
	}
	return gv.ToMutator().ResetMetadata().ResetPrerelease().String(), nil
}

func GetBaseVersion(client discovery.DiscoveryInterface) (string, error) {
	info, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	gv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return "", err
	}
	return gv.ToMutator().ResetMetadata().ResetPrerelease().ResetPatch().String(), nil
}

func CheckAPIVersion(client discovery.DiscoveryInterface, constraint string) (bool, error) {
	info, err := client.ServerVersion()
	if err != nil {
		return false, err
	}
	cond, err := version.NewConstraint(constraint)
	if err != nil {
		return false, err
	}
	v, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return false, err
	}
	return cond.Check(v.ToMutator().ResetPrerelease().ResetMetadata().Done()), nil
}

func IsPreferredAPIResource(client discovery.DiscoveryInterface, groupVersion, kind string) bool {
	if resourceList, err := client.ServerPreferredResources(); discovery.IsGroupDiscoveryFailedError(err) || err == nil {
		for _, resources := range resourceList {
			if resources.GroupVersion != groupVersion {
				continue
			}
			for _, resource := range resources.APIResources {
				if resource.Kind == kind {
					return true
				}
			}
		}
	}
	return false
}

type KnownBug struct {
	URL string
	Fix string
}

func (e *KnownBug) Error() string {
	return "Bug: " + e.URL + ". To fix, " + e.Fix
}

var err62649_K1_9 = &KnownBug{URL: "https://github.com/kubernetes/kubernetes/pull/62649", Fix: "upgrade to Kubernetes 1.9.8 or later."}
var err62649_K1_10 = &KnownBug{URL: "https://github.com/kubernetes/kubernetes/pull/62649", Fix: "upgrade to Kubernetes 1.10.2 or later."}

var (
	DefaultConstraint                     = ">= 1.9.0"
	DefaultBlackListedVersions            map[string]error
	DefaultBlackListedMultiMasterVersions = map[string]error{
		"1.9.0":  err62649_K1_9,
		"1.9.1":  err62649_K1_9,
		"1.9.2":  err62649_K1_9,
		"1.9.3":  err62649_K1_9,
		"1.9.4":  err62649_K1_9,
		"1.9.5":  err62649_K1_9,
		"1.9.6":  err62649_K1_9,
		"1.9.7":  err62649_K1_9,
		"1.10.0": err62649_K1_10,
		"1.10.1": err62649_K1_10,
	}
)

func IsDefaultSupportedVersion(kc kubernetes.Interface) error {
	return IsSupportedVersion(
		kc,
		DefaultConstraint,
		DefaultBlackListedVersions,
		DefaultBlackListedMultiMasterVersions)
}

func IsSupportedVersion(kc kubernetes.Interface, constraint string, blackListedVersions map[string]error, blackListedMultiMasterVersions map[string]error) error {
	info, err := kc.Discovery().ServerVersion()
	if err != nil {
		return err
	}
	glog.Infof("Kubernetes version: %#v\n", info)

	gv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return err
	}
	v := gv.ToMutator().ResetMetadata().ResetPrerelease().Done()

	nodes, err := kc.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master",
	})
	if err != nil {
		return err
	}
	multiMaster := len(nodes.Items) > 1

	return checkVersion(v, multiMaster, constraint, blackListedVersions, blackListedMultiMasterVersions)
}

func checkVersion(v *version.Version, multiMaster bool, constraint string, blackListedVersions map[string]error, blackListedMultiMasterVersions map[string]error) error {
	vs := v.String()

	if constraint != "" {
		c, err := version.NewConstraint(constraint)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return fmt.Errorf("kubernetes version %s fails constraint %s", vs, constraint)
		}
	}

	if e, ok := blackListedVersions[v.Original()]; ok {
		return errors.Wrapf(e, "kubernetes version %s is blacklisted", v.Original())
	}
	if e, ok := blackListedVersions[vs]; ok {
		return errors.Wrapf(e, "kubernetes version %s is blacklisted", vs)
	}

	if multiMaster {
		if e, ok := blackListedMultiMasterVersions[v.Original()]; ok {
			return errors.Wrapf(e, "kubernetes version %s is blacklisted for multi-master cluster", v.Original())
		}
		if e, ok := blackListedMultiMasterVersions[vs]; ok {
			return errors.Wrapf(e, "kubernetes version %s is blacklisted for multi-master cluster", vs)
		}
	}
	return nil
}
