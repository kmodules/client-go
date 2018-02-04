package discovery_test

import (
	"path/filepath"
	"testing"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/discovery"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestRestMapper(t *testing.T) {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	restmapper, err := discovery.LoadRestMapper(kc.Discovery())
	if err != nil {
		t.Fatal(err)
	}

	data := []struct {
		in  interface{}
		out schema.GroupVersionResource
	}{
		{&core.Pod{}, core.SchemeGroupVersion.WithResource("pods")},
		{&core.Service{}, core.SchemeGroupVersion.WithResource("services")},
	}

	for _, tt := range data {
		gvr, err := discovery.DetectResource(restmapper, tt.in)
		if err != nil {
			t.Error(err)
		}
		if gvr != tt.out {
			t.Errorf("Failed to DetectResource: expected %+v, got %+v", tt.out, gvr)
		}
	}
}
