package tools

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	netz "github.com/appscode/go/net"
	"k8s.io/client-go/util/cert"
)

type CertManager struct {
	caKey  *rsa.PrivateKey
	caCert *x509.Certificate

	dir    string
	Expiry time.Duration
}

func NewCertStore(rootDir string) (*CertManager, error) {
	dir := filepath.Join(rootDir, "pki")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir `%s`. Reason: %v", dir, err)
	}
	cm := &CertManager{dir: dir}
	if cm.PairExists("ca") {
		cm.caCert, cm.caKey, err = cm.Read("ca")
		if err == nil {
			return cm, nil
		}
	}

	key, err := cert.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	cfg := cert.Config{
		CommonName:   "ca",
		Organization: []string{"AppsCode", "Eng"},
		AltNames: cert.AltNames{
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		},
	}
	crt, err := cert.NewSelfSignedCACert(cfg, key)
	if err != nil {
		return nil, err
	}
	err = cm.Write("ca", crt, key)
	if err != nil {
		return nil, err
	}

	cm.caCert = crt
	cm.caKey = key
	return cm, nil
}

func (cm *CertManager) Location() string {
	return cm.dir
}

func (cm *CertManager) CACert() []byte {
	return cert.EncodeCertPEM(cm.caCert)
}

func (cm *CertManager) CAKey() []byte {
	return cert.EncodePrivateKeyPEM(cm.caKey)
}

func (cm *CertManager) NewHostCertPair() ([]byte, []byte, error) {
	var sans cert.AltNames
	publicIPs, privateIPs, _ := netz.HostIPs()
	for _, ip := range publicIPs {
		sans.IPs = append(sans.IPs, net.ParseIP(ip))
	}
	for _, ip := range privateIPs {
		sans.IPs = append(sans.IPs, net.ParseIP(ip))
	}
	return cm.NewServerCertPair("127.0.0.1", sans)
}

func (cm *CertManager) NewServerCertPair(cn string, sans cert.AltNames) ([]byte, []byte, error) {
	cfg := cert.Config{
		CommonName:   cn,
		Organization: []string{"AppsCode", "Eng"},
		AltNames:     sans,
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	key, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	crt, err := cert.NewSignedCert(cfg, key, cm.caCert, cm.caKey)
	if err != nil {
		return nil, nil, err
	}
	return cert.EncodeCertPEM(crt), cert.EncodePrivateKeyPEM(key), nil
}

func (cm *CertManager) NewClientCertPair(cn string) ([]byte, []byte, error) {
	cfg := cert.Config{
		CommonName:   cn,
		Organization: []string{"AppsCode", "Eng"},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	key, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	crt, err := cert.NewSignedCert(cfg, key, cm.caCert, cm.caKey)
	if err != nil {
		return nil, nil, err
	}
	return cert.EncodeCertPEM(crt), cert.EncodePrivateKeyPEM(key), nil
}

func (cm *CertManager) IsExists(name string) bool {
	if _, err := os.Stat(cm.CertFile(name)); err == nil {
		return true
	}
	if _, err := os.Stat(cm.KeyFile(name)); err == nil {
		return true
	}
	return false
}

func (cm *CertManager) PairExists(name string) bool {
	if _, err := os.Stat(cm.CertFile(name)); err == nil {
		if _, err := os.Stat(cm.KeyFile(name)); err == nil {
			return true
		}
	}
	return false
}

func (cm *CertManager) CertFile(name string) string {
	return filepath.Join(cm.dir, strings.ToLower(name)+".crt")
}

func (cm *CertManager) KeyFile(name string) string {
	return filepath.Join(cm.dir, strings.ToLower(name)+".key")
}

func (cm *CertManager) Write(name string, crt *x509.Certificate, key *rsa.PrivateKey) error {
	if err := ioutil.WriteFile(cm.CertFile(name), cert.EncodeCertPEM(crt), 0644); err != nil {
		return fmt.Errorf("failed to write `%cm`. Reason: %v", cm.CertFile(name), err)
	}
	if err := ioutil.WriteFile(cm.KeyFile(name), cert.EncodePrivateKeyPEM(key), 0600); err != nil {
		return fmt.Errorf("failed to write `%cm`. Reason: %v", cm.KeyFile(name), err)
	}
	return nil
}

func (cm *CertManager) WriteBytes(name string, crt, key []byte) error {
	if err := ioutil.WriteFile(cm.CertFile(name), crt, 0644); err != nil {
		return fmt.Errorf("failed to write `%cm`. Reason: %v", cm.CertFile(name), err)
	}
	if err := ioutil.WriteFile(cm.KeyFile(name), key, 0600); err != nil {
		return fmt.Errorf("failed to write `%cm`. Reason: %v", cm.KeyFile(name), err)
	}
	return nil
}

func (cm *CertManager) Read(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	crtBytes, err := ioutil.ReadFile(cm.CertFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate `%cm`. Reason: %v", cm.CertFile(name), err)
	}
	crt, err := cert.ParseCertsPEM(crtBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate `%cm`. Reason: %v", cm.CertFile(name), err)
	}

	keyBytes, err := ioutil.ReadFile(cm.KeyFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key `%cm`. Reason: %v", cm.KeyFile(name), err)
	}
	key, err := cert.ParsePrivateKeyPEM(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key `%cm`. Reason: %v", cm.KeyFile(name), err)
	}
	return crt[0], key.(*rsa.PrivateKey), nil
}

func (cm *CertManager) ReadBytes(name string) ([]byte, []byte, error) {
	crtBytes, err := ioutil.ReadFile(cm.CertFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate `%cm`. Reason: %v", cm.CertFile(name), err)
	}

	keyBytes, err := ioutil.ReadFile(cm.KeyFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key `%cm`. Reason: %v", cm.KeyFile(name), err)
	}
	return crtBytes, keyBytes, nil
}
