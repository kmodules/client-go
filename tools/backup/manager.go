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

package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

const (
	timestampFormat = "20060102T150405"
)

type ItemList struct {
	Items []map[string]interface{} `json:"items,omitempty"`
}

type BackupManager struct {
	cluster  string
	config   *rest.Config
	mapper   meta.RESTMapper
	sanitize bool
}

func NewBackupManager(cluster string, config *rest.Config, sanitize bool) BackupManager {
	hc, err := rest.HTTPClientFor(config)
	if err != nil {
		panic(err)
	}
	mapper, err := apiutil.NewDynamicRESTMapper(config, hc)
	if err != nil {
		panic(err)
	}
	return BackupManager{
		cluster:  cluster,
		config:   config,
		mapper:   mapper,
		sanitize: sanitize,
	}
}

type processorFunc func(relPath string, data []byte) error

func (mgr BackupManager) snapshotPrefix(t time.Time) string {
	if mgr.cluster == "" {
		return "snapshot-" + t.UTC().Format(timestampFormat)
	}
	return mgr.cluster + "-" + t.UTC().Format(timestampFormat)
}

func (mgr BackupManager) BackupToDir(backupDir string) (string, error) {
	snapshotDir := mgr.snapshotPrefix(time.Now())
	p := func(relPath string, data []byte) error {
		absPath := filepath.Join(backupDir, snapshotDir, relPath)
		dir := filepath.Dir(absPath)
		err := os.MkdirAll(dir, 0o777)
		if err != nil {
			return err
		}
		return os.WriteFile(absPath, data, 0o644)
	}
	return snapshotDir, mgr.Backup(p)
}

func (mgr BackupManager) BackupToTar(backupDir string) (string, error) {
	err := os.MkdirAll(backupDir, 0o777)
	if err != nil {
		return "", err
	}

	t := time.Now()
	prefix := mgr.snapshotPrefix(t)

	fileName := filepath.Join(backupDir, prefix+".tar.gz")
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// set up the gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	p := func(relPath string, data []byte) error {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = relPath
		header.Size = int64(len(data))
		header.Mode = 0o666
		header.ModTime = t
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, bytes.NewReader(data)); err != nil {
			return err
		}
		return nil
	}
	return fileName, mgr.Backup(p)
}

func (mgr BackupManager) Backup(process processorFunc) error {
	// ref: https://github.com/kubernetes/ingress-nginx/blob/0dab51d9eb1e5a9ba3661f351114825ac8bfc1af/pkg/ingress/controller/launch.go#L252
	mgr.config.QPS = 1e6
	mgr.config.Burst = 1e6
	if err := rest.SetKubernetesDefaults(mgr.config); err != nil {
		return err
	}
	mgr.config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	if mgr.config.UserAgent == "" {
		mgr.config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	disClient, err := discovery.NewDiscoveryClientForConfig(mgr.config)
	if err != nil {
		return err
	}
	resourceLists, err := disClient.ServerPreferredResources()
	if err != nil {
		return err
	}
	resourceListBytes, err := yaml.Marshal(resourceLists)
	if err != nil {
		return err
	}
	err = process("resource_lists.yaml", resourceListBytes)
	if err != nil {
		return err
	}

	for _, list := range resourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			return err
		}
		for _, r := range list.APIResources {
			if strings.ContainsRune(r.Name, '/') {
				continue // skip subresource
			}
			if !sets.NewString(r.Verbs...).HasAll("list", "get") {
				continue
			}

			klog.V(3).Infof("Taking backup of %s apiVersion:%s kind:%s", list.GroupVersion, r.Name, r.Kind)
			mgr.config.GroupVersion = &gv
			mgr.config.APIPath = "/apis"
			if gv.Group == core.GroupName {
				mgr.config.APIPath = "/api"
			}
			client, err := rest.RESTClientFor(mgr.config)
			if err != nil {
				return err
			}
			request := client.Get().Resource(r.Name).Param("pretty", "true")
			resp, err := request.DoRaw(context.TODO())
			if err != nil {
				return err
			}
			items := &ItemList{}
			err = yaml.Unmarshal(resp, &items)
			if err != nil {
				return err
			}
			for _, item := range items.Items {
				var path string
				item["apiVersion"] = list.GroupVersion
				item["kind"] = r.Kind

				md, ok := item["metadata"]
				if ok {
					path = getPathFromSelfLink(mgr.mapper, item)
					if mgr.sanitize {
						cleanUpObjectMeta(md)
					}
				}
				if mgr.sanitize {
					if spec, ok := item["spec"].(map[string]interface{}); ok {
						switch r.Kind {
						case "Pod":
							item["spec"], err = cleanUpPodSpec(spec)
							if err != nil {
								return err
							}
						case "StatefulSet", "Deployment", "ReplicaSet", "DaemonSet", "ReplicationController", "Job":
							template, ok := spec["template"].(map[string]interface{})
							if ok {
								podSpec, ok := template["spec"].(map[string]interface{})
								if ok {
									template["spec"], err = cleanUpPodSpec(podSpec)
									if err != nil {
										return err
									}
								}
							}
						}
					}
					delete(item, "status")
				}
				data, err := yaml.Marshal(item)
				if err != nil {
					return err
				}
				err = process(path, data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cleanUpObjectMeta(md interface{}) {
	meta, ok := md.(map[string]interface{})
	if !ok {
		return
	}
	delete(meta, "creationTimestamp")
	delete(meta, "resourceVersion")
	delete(meta, "uid")
	delete(meta, "generateName")
	delete(meta, "generation")
	annotation, ok := meta["annotations"]
	if !ok {
		return
	}
	annotations, ok := annotation.(map[string]string)
	if !ok {
		return
	}
	cleanUpDecorators(annotations)
}

func cleanUpDecorators(m map[string]string) {
	delete(m, "controller-uid")
	delete(m, "deployment.kubernetes.io/desired-replicas")
	delete(m, "deployment.kubernetes.io/max-replicas")
	delete(m, "deployment.kubernetes.io/revision")
	delete(m, "pod-template-hash")
	delete(m, "pv.kubernetes.io/bind-completed")
	delete(m, "pv.kubernetes.io/bound-by-controller")
}

func cleanUpPodSpec(in map[string]interface{}) (map[string]interface{}, error) {
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	spec := &core.PodSpec{}
	err = yaml.Unmarshal(b, spec)
	if err != nil {
		return in, nil // Not a podSpec
	}
	spec.DNSPolicy = core.DNSPolicy("")
	spec.NodeName = ""
	if spec.ServiceAccountName == "default" {
		spec.ServiceAccountName = ""
	}
	spec.TerminationGracePeriodSeconds = nil
	for i, c := range spec.Containers {
		c.TerminationMessagePath = ""
		spec.Containers[i] = c
	}
	for i, c := range spec.InitContainers {
		c.TerminationMessagePath = ""
		spec.InitContainers[i] = c
	}
	b, err = yaml.Marshal(spec)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	err = yaml.Unmarshal(b, &out)
	return out, err
}

func getPathFromSelfLink(mapper meta.RESTMapper, obj map[string]interface{}) string {
	u := unstructured.Unstructured{Object: obj}
	gvk := u.GetObjectKind().GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		panic(err)
	}
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		return fmt.Sprintf("%s/%s/namespaces/%s/%s/%s.yaml", gvk.Group, gvk.Version, u.GetNamespace(), mapping.Resource.Resource, u.GetName())
	}
	return fmt.Sprintf("%s/%s/%s/%s.yaml", gvk.Group, gvk.Version, mapping.Resource.Resource, u.GetName())
}
