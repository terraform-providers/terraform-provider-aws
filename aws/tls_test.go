package aws

import (
	"strings"
	"testing"
)

func TestTlsRsaPrivateKeyPem(t *testing.T) {
	key := tlsRsaPrivateKeyPem(2048)

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("key does not contain RSA PRIVATE KEY: %s", key)
	}
}

func TestTlsRsaPublicKeyPem(t *testing.T) {
	privateKey := tlsRsaPrivateKeyPem(2048)
	publicKey := tlsRsaPublicKeyPem(privateKey)

	if !strings.Contains(publicKey, pemBlockTypePublicKey) {
		t.Errorf("key does not contain PUBLIC KEY: %s", publicKey)
	}
}

func TestTlsRsaX509LocallySignedCertificatePem(t *testing.T) {
	caKey := tlsRsaPrivateKeyPem(2048)
	caCertificate := tlsRsaX509SelfSignedCaCertificatePem(caKey)
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509LocallySignedCertificatePem(caKey, caCertificate, key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTlsRsaX509SelfSignedCaCertificatePem(t *testing.T) {
	caKey := tlsRsaPrivateKeyPem(2048)
	caCertificate := tlsRsaX509SelfSignedCaCertificatePem(caKey)

	if !strings.Contains(caCertificate, pemBlockTypeCertificate) {
		t.Errorf("CA certificate does not contain CERTIFICATE: %s", caCertificate)
	}
}

func TestTlsRsaX509SelfSignedCertificatePem(t *testing.T) {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	if !strings.Contains(certificate, pemBlockTypeCertificate) {
		t.Errorf("certificate does not contain CERTIFICATE: %s", certificate)
	}
}

func TestTlsRsaX509CertificateRequestPem(t *testing.T) {
	csr, key := tlsRsaX509CertificateRequestPem(2048, "example.com")

	if !strings.Contains(csr, pemBlockTypeCertificateRequest) {
		t.Errorf("certificate does not contain CERTIFICATE REQUEST: %s", csr)
	}

	if !strings.Contains(key, pemBlockTypeRsaPrivateKey) {
		t.Errorf("certificate does not contain RSA PRIVATE KEY: %s", key)
	}
}
