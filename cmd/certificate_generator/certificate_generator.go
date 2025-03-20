package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

const (
	rsaKeyBits = 2048
	fileMode   = 0o600
)

func main() {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"url_shortener"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		log.Fatalf("Can generate rsa key: %s", err)
	}

	certificate, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Can not create certificate: %s", err)
	}

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	err = os.WriteFile("cert.pem", certBytes, fileMode)
	if err != nil {
		log.Fatalf("Can not write certificate to file: %s", err)
	}

	err = os.WriteFile("key.pem", keyBytes, fileMode)
	if err != nil {
		log.Fatalf("Can not write certificate to file: %s", err)
	}
}
