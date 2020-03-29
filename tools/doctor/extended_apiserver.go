/*
Copyright The Kmodules Authors.

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

package doctor

import (
	"context"
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	authenticationConfigMapNamespace = metav1.NamespaceSystem
	// authenticationConfigMapName is the name of ConfigMap in the kube-system namespace holding the root certificate
	// bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified
	// by --requestheader-username-headers. This is created in the cluster by the kube-apiserver.
	// "WARNING: generally do not depend on authorization being already done for incoming requests.")
	authenticationConfigMapName = "extension-apiserver-authentication"
)

func (d *Doctor) extractExtendedAPIServerInfo(info *ClusterInfo) error {
	authConfigMap, err := d.kc.CoreV1().ConfigMaps(authenticationConfigMapNamespace).Get(context.TODO(), authenticationConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	clientCA, ok := authConfigMap.Data["client-ca-file"]
	if ok {
		info.ExtensionServerConfig.ClientCAData = strings.TrimSpace(clientCA)
	}

	requestHeaderCA, ok := authConfigMap.Data["requestheader-client-ca-file"]
	if !ok {
		return nil
	}

	usernameHeaders, err := deserializeStrings(authConfigMap.Data["requestheader-username-headers"])
	if err != nil {
		return err
	}
	groupHeaders, err := deserializeStrings(authConfigMap.Data["requestheader-group-headers"])
	if err != nil {
		return err
	}
	extraHeaderPrefixes, err := deserializeStrings(authConfigMap.Data["requestheader-extra-headers-prefix"])
	if err != nil {
		return err
	}
	allowedNames, err := deserializeStrings(authConfigMap.Data["requestheader-allowed-names"])
	if err != nil {
		return err
	}

	info.ExtensionServerConfig.RequestHeader = &RequestHeaderConfig{
		UsernameHeaders:     usernameHeaders,
		GroupHeaders:        groupHeaders,
		ExtraHeaderPrefixes: extraHeaderPrefixes,
		CAData:              strings.TrimSpace(requestHeaderCA),
		AllowedClientNames:  allowedNames,
	}

	return nil
}

func deserializeStrings(in string) ([]string, error) {
	if len(in) == 0 {
		return nil, nil
	}
	var ret []string
	if err := json.Unmarshal([]byte(in), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
