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
	if format == meta.JsonFormat {
		out, err := ylib.ToJSON(data)
		if err != nil {
			return nil, "", err
		}
		return out, "application/json", nil
	} else if format == meta.YAMLFormat {
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
