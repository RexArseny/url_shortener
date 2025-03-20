package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("successful server creation", func(t *testing.T) {
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

		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		certificate, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
		assert.NoError(t, err)

		certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
		keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

		err = os.WriteFile("cert.pem", certBytes, 0o600)
		assert.NoError(t, err)

		err = os.WriteFile("key.pem", keyBytes, 0o600)
		assert.NoError(t, err)

		t.Setenv("PUBLIC_KEY_PATH", "../../public.pem")
		t.Setenv("PRIVATE_KEY_PATH", "../../private.pem")
		t.Setenv("ENABLE_HTTPS", "true")
		t.Setenv("CERTIFICATE_PATH", "cert.pem")
		t.Setenv("CERTIFICATE_KEY_PATH", "key.pem")

		defer func() {
			err = os.Remove("cert.pem")
			assert.NoError(t, err)
			err = os.Remove("key.pem")
			assert.NoError(t, err)
		}()

		go func() {
			err := NewServer()
			assert.NoError(t, err)
		}()

		time.Sleep(time.Second * 5)

		t.SkipNow()
	})
}
