package meta

import (
	"hash"
	"hash/fnv"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerationHash(m metav1.ObjectMeta) string {
	data := make(map[string]interface{}, 3)
	data["generation"] = m.Generation
	if len(m.Labels) > 0 {
		data["labels"] = m.Labels
	}
	if len(m.Annotations) > 0 {
		data["annotations"] = m.Annotations
	}
	h := fnv.New64a()
	deepHashObject(h, data)
	return strconv.FormatUint(h.Sum64(), 10)
}

// deepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func deepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
