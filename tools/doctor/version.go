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
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
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

	gvPtr, err := semver.NewVersion(v.GitVersion)
	if err != nil {
		return errors.Wrapf(err, "invalid version %s", v.GitVersion)
	}
	gv := *gvPtr
	gv, _ = gv.SetPrerelease("")
	gv, _ = gv.SetMetadata("")
	info.Version.Patch = gv.Original()
	info.Version.Minor = fmt.Sprintf("%s%d.%d.0", originalVPrefix(gv), gv.Major(), gv.Minor())

	return err
}

func originalVPrefix(v semver.Version) string {
	// Note, only lowercase v is supported as a prefix by the parser.
	if v.Original() != "" && v.Original()[:1] == "v" {
		return v.Original()[:1]
	}
	return ""
}
