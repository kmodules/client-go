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

package converter

import (
	"bytes"
	"path/filepath"
	"unicode"

	"kmodules.xyz/client-go/meta"

	gomime "github.com/cubewise-code/go-mime"
	"github.com/gabriel-vasile/mimetype"
	ylib "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

func Convert(name string, data []byte, format meta.DataFormat) ([]byte, string, error) {
	switch format {
	case meta.JsonFormat:
		out, err := ylib.ToJSON(data)
		if err != nil {
			return nil, "", err
		}
		return out, "application/json", nil
	case meta.YAMLFormat:
		if hasJSONPrefix(data) {
			out, err := yaml.JSONToYAML(data)
			if err != nil {
				return nil, "", err
			}
			return out, "text/yaml", nil
		}
	}
	ext := filepath.Ext(name)
	if ct := gomime.TypeByExtension(ext); ct != "" {
		return data, ct, nil
	}

	ct := mimetype.Detect(data)
	return data, ct.String(), nil
}

// ref: https://github.com/kubernetes/apimachinery/tree/v0.18.9/pkg/util/yaml

var jsonPrefix = []byte("{")

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is
// prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}
