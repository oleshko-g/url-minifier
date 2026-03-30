package http

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"
	"os"
	"time"
)

const (
	certFile, keyFile string = "cert.der", "key.der"
)

func TLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		err = writeX509KeyPair()
		if err != nil {
			return nil
		}

		cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil
		}
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}
}

// createCertificate creates [x509.Certificate] and writes it in the [certFile]
func writeX509KeyPair() error {
	template := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"minifier"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0), // add 1 year
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// self-signed certificate
	parent := template

	certData, err := x509.CreateCertificate(
		rand.Reader,
		template,
		parent,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return err
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certData,
	})
	if err != nil {
		return err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	if err = os.WriteFile(certFile, certPEM.Bytes(), 0o600); err != nil {
		return err
	}

	if err = os.WriteFile(keyFile, privateKeyPEM.Bytes(), 0o600); err != nil {
		return err
	}

	return nil
}
