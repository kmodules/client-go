package meta

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func IPv6Enabled(kc kubernetes.Interface) (bool, error) {
	svc, err := kc.CoreV1().Services(metav1.NamespaceDefault).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	clusterIPs := []string{svc.Spec.ClusterIP}
	for _, ip := range clusterIPs {
		if strings.ContainsRune(ip, ':') {
			return true, nil
		}
	}
	return false, nil
}
