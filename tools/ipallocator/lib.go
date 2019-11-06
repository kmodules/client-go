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
package ipallocator

import (
	"strconv"
	"strings"
)

type DiscoverVia string

const (
	// DiscoverViaIP is a replacement for kube-dns. It uses a predefined map for service name to cluster ip.
	// This reduces Kubernetes related overhead in a cluster.
	DiscoverViaIP  DiscoverVia = "ip"
	DiscoverViaDNS DiscoverVia = "dns"
)

type IPAllocator struct {
	serviceSubnet string
	services      map[string]int64
	discoverVia   DiscoverVia // all services must be in the same namespace
}

func New(serviceSubnet string, services []string, discoverVia DiscoverVia) *IPAllocator {
	ipa := &IPAllocator{
		serviceSubnet: serviceSubnet,
		services:      map[string]int64{},
		discoverVia:   discoverVia,
	}
	for i, svc := range services {
		ipa.services[svc] = int64(i + 1)
	}
	return ipa
}

func (ipa IPAllocator) ClusterIP(svc string) string {
	if ipa.discoverVia == DiscoverViaDNS {
		return ""
	}
	seq, ok := ipa.services[svc]
	if !ok {
		return ""
	}
	octets := strings.Split(ipa.serviceSubnet, ".")
	p, _ := strconv.ParseInt(octets[3], 10, 64)
	p = p + seq
	octets[3] = strconv.FormatInt(p, 10)
	return strings.Join(octets, ".")
}

func (ipa IPAllocator) ServiceAddress(svc string) string {
	switch ipa.discoverVia {
	case DiscoverViaDNS:
		return svc
	case DiscoverViaIP:
		return ipa.ClusterIP(svc)
	default:
		panic("unknown discovery mechanism " + ipa.discoverVia)
	}
}

func (ipa IPAllocator) HostNetwork() bool {
	return ipa.discoverVia != DiscoverViaDNS
}
