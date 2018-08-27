package meta

import (
	"hash"
	"hash/fnv"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerationHash(in metav1.ObjectMeta) string {
	obj := make(map[string]interface{}, 3)
	obj["generation"] = in.Generation
	if len(in.Labels) > 0 {
		obj["labels"] = in.Labels
	}
	if len(in.Annotations) > 0 {
		data := make(map[string]string, len(in.Annotations))
		for k, v := range in.Annotations {
			if k != lastAppliedConfiguration {
				data[k] = v
			}
		}
		obj["annotations"] = data
	}
	h := fnv.New64a()
	DeepHashObject(h, obj)
	return strconv.FormatUint(h.Sum64(), 10)
}

// deepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
