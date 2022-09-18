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

package doctor

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

func (d *Doctor) extractKubeCA(info *ClusterInfo) error {
	info.ClientConfig.Host = d.config.Host
	info.ClientConfig.Insecure = d.config.Insecure

	if len(d.config.CAData) > 0 {
		info.ClientConfig.CAData = strings.TrimSpace(string(d.config.CAData))
	} else if len(d.config.CAFile) > 0 {
		data, err := os.ReadFile(d.config.CAFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load ca file %s", d.config.CAFile)
		}
		info.ClientConfig.CAData = strings.TrimSpace(string(data))
	}
	return nil
}
