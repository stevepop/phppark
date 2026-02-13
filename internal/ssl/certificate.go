package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CertificatePaths holds paths to certificate files
type CertificatePaths struct {
	CertFile string // .crt file
	KeyFile  string // .key file
}

// GenerateSelfSignedCert generates a self-signed SSL certificate
func GenerateSelfSignedCert(siteName, domain, certDir string) (*CertificatePaths, error) {
	// Ensure certificate directory exists
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serverName := fmt.Sprintf("%s.%s", siteName, domain)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"PHPark Development"},
			CommonName:   serverName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{serverName, "localhost"},
		IPAddresses:           nil,
	}

	// Create self-signed certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Certificate file path
	certPath := filepath.Join(certDir, siteName+".crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	// Write certificate
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return nil, fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Private key file path
	keyPath := filepath.Join(certDir, siteName+".key")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	// Write private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	// Set permissions (private key should be read-only by owner)
	if err := os.Chmod(keyPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set key permissions: %w", err)
	}

	return &CertificatePaths{
		CertFile: certPath,
		KeyFile:  keyPath,
	}, nil
}

// CertificateExists checks if certificates exist for a site
func CertificateExists(siteName, certDir string) bool {
	certPath := filepath.Join(certDir, siteName+".crt")
	keyPath := filepath.Join(certDir, siteName+".key")

	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)

	return certErr == nil && keyErr == nil
}

// RemoveCertificate removes certificate files for a site
func RemoveCertificate(siteName, certDir string) error {
	certPath := filepath.Join(certDir, siteName+".crt")
	keyPath := filepath.Join(certDir, siteName+".key")

	// Remove certificate file
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate: %w", err)
	}

	// Remove key file
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove key: %w", err)
	}

	return nil
}
