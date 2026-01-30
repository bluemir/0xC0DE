package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	// YAML
	yamlContent := `
backend:
  auth:
    salt: "test-salt"
`
	yamlPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	conf, err := readCofigFile(yamlPath)
	assert.NoError(t, err)
	assert.Equal(t, "test-salt", conf.Backend.Auth.Salt)

	// HJSON
	hjsonContent := `
{
  backend: {
    auth: {
      salt: "test-salt-hjson"
    }
  }
}
`
	hjsonPath := filepath.Join(tempDir, "config.hjson")
	err = os.WriteFile(hjsonPath, []byte(hjsonContent), 0644)
	require.NoError(t, err)

	conf, err = readCofigFile(hjsonPath)
	assert.NoError(t, err)
	assert.Equal(t, "test-salt-hjson", conf.Backend.Auth.Salt)

	// Unknown Ext
	unknownPath := filepath.Join(tempDir, "config.txt")
	err = os.WriteFile(unknownPath, []byte(""), 0644)
	require.NoError(t, err)

	_, err = readCofigFile(unknownPath)
	assert.Error(t, err)
}

func TestGetTLSConfig(t *testing.T) {
	// Generate self-signed cert for testing
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certOut := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyOut := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "server.crt")
	keyPath := filepath.Join(tempDir, "server.key")

	err = os.WriteFile(certPath, certOut, 0644)
	require.NoError(t, err)
	err = os.WriteFile(keyPath, keyOut, 0600)
	require.NoError(t, err)

	// Test CertConfig.Load
	certConfig := CertConfig{
		CertFile: certPath,
		KeyFile:  keyPath,
	}
	cert, err := certConfig.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cert)

	// Test getTLSConfig
	tlsConf, err := getTLSConfig(cert, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tlsConf)
	assert.Len(t, tlsConf.Certificates, 1)
}
