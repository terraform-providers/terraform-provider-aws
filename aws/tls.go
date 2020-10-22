package aws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"time"
)

const (
	pemBlockTypeCertificate   = `CERTIFICATE`
	pemBlockTypeRsaPrivateKey = `RSA PRIVATE KEY`
	pemBlockTypePublicKey     = `PUBLIC KEY`
)

var tlsX509CertificateSerialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)

// tlsRsaPrivateKeyPem generates a RSA private key PEM string.
// Wrap with tlsPemEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func tlsRsaPrivateKeyPem(bits int) string {
	key, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(key),
		Type:  pemBlockTypeRsaPrivateKey,
	}

	return string(pem.EncodeToMemory(block))
}

// tlsRsaPublicKeyPem generates a RSA public key PEM string.
// Wrap with tlsPemEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: public_key_pem = "%[1]s"
func tlsRsaPublicKeyPem(keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	block := &pem.Block{
		Bytes: publicKeyBytes,
		Type:  pemBlockTypePublicKey,
	}

	return string(pem.EncodeToMemory(block))
}

// tlsRsaX509LocallySignedCertificatePem generates a local CA x509 certificate PEM string.
// Wrap with tlsPemEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: certificate_pem = "%[1]s"
func tlsRsaX509LocallySignedCertificatePem(caKeyPem, caCertificatePem, keyPem, commonName string) string {
	caCertificateBlock, _ := pem.Decode([]byte(caCertificatePem))

	caCertificate, err := x509.ParseCertificate(caCertificateBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	caKeyBlock, _ := pem.Decode([]byte(caKeyPem))

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour),
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, caCertificate, &key.PublicKey, caKey)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// tlsRsaX509SelfSignedCaCertificatePem generates a x509 CA certificate PEM string.
// Wrap with tlsPemEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: root_certificate_pem = "%[1]s"
func tlsRsaX509SelfSignedCaCertificatePem(keyPem string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	publicKeyBytesSha1 := sha1.Sum(publicKeyBytes)

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour),
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   "ACME Root CA",
			Organization: []string{"ACME Examples, Inc"},
		},
		SubjectKeyId: publicKeyBytesSha1[:],
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

// tlsRsaX509SelfSignedCertificatePem generates a x509 certificate PEM string.
// Wrap with tlsPemEscapeNewlines() to allow simple fmt.Sprintf()
// configurations such as: private_key_pem = "%[1]s"
func tlsRsaX509SelfSignedCertificatePem(keyPem, commonName string) string {
	keyBlock, _ := pem.Decode([]byte(keyPem))

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	serialNumber, err := rand.Int(rand.Reader, tlsX509CertificateSerialNumberLimit)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificate := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(24 * time.Hour),
		NotBefore:             time.Now(),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ACME Examples, Inc"},
		},
	}

	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)

	if err != nil {
		//lintignore:R009
		panic(err)
	}

	certificateBlock := &pem.Block{
		Bytes: certificateBytes,
		Type:  pemBlockTypeCertificate,
	}

	return string(pem.EncodeToMemory(certificateBlock))
}

func tlsPemEscapeNewlines(pem string) string {
	return strings.ReplaceAll(pem, "\n", "\\n")
}
