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
	"github.com/pkg/errors"
	"gomodules.xyz/version"
)

func (d *Doctor) extractVersion(info *ClusterInfo) error {
	v, err := d.kc.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	info.Version = &VersionInfo{
		GitVersion: v.GitVersion,
		GitCommit:  v.GitCommit,
		BuildDate:  v.BuildDate,
		Platform:   v.Platform,
	}

	gv, err := version.NewVersion(v.GitVersion)
	if err != nil {
		return errors.Wrapf(err, "invalid version %s", v.GitVersion)
	}
	mv := gv.ToMutator().ResetMetadata().ResetPrerelease()
	info.Version.Patch = mv.String()
	info.Version.Minor = mv.ResetPatch().String()

	return err
}
