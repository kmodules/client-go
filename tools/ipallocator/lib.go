package ipallocator

import (
	"strconv"
	"strings"
)

type IPAllocator struct {
	serviceSubnet string
	services      map[string]int64
	useDNS        bool // all services must be in the same namespace
}

func New(serviceSubnet string, services []string) *IPAllocator {
	ipa := &IPAllocator{
		serviceSubnet: serviceSubnet,
		services:      map[string]int64{},
	}
	for i, svc := range services {
		ipa.services[svc] = int64(i + 1)
	}
	return ipa
}

func (ipa IPAllocator) ClusterIP(svc string) string {
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
	if ipa.useDNS {
		return svc
	}
	return ipa.ClusterIP(svc)
}
