package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func configureTLSConfig(server *http.Server, config TLSConf) error {
	if config.CAFile == "" {
		return nil
	}

	certPool, err := certPool(config.CAFile)
	if err != nil {
		return errors.Wrap(err, "failed to get CA certificates")
	}

	certBytes, err := ioutil.ReadFile(config.CertFile)
	if err != nil {
		return errors.Wrap(err, "failed to read cert file")
	}

	keyBytes, err := ioutil.ReadFile(config.KeyFile)
	if err != nil {
		return errors.Wrap(err, "failed to read private key file")
	}

	cert, err := loadX509KeyPair(certBytes, keyBytes, config.KeyPassword)
	if err != nil {
		return err
	}

	// Create the TLS Config with the CA pool
	tlsConf := &tls.Config{
		ClientCAs:    certPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if config.ForceClientCertAuth {
		// force Client certificate authentication for extra security
		tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
	}

	server.TLSConfig = tlsConf
	server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

	return nil
}

func certPool(caRootPath string) (certPool *x509.CertPool, err error) {
	certPool = x509.NewCertPool()
	_, err = os.Lstat(caRootPath)
	if err != nil {
		return nil, errors.Wrap(err, "cert pool failed to stat file")
	}

	var buf []byte
	if buf, err = ioutil.ReadFile(caRootPath); err != nil {
		return nil, errors.Wrap(err, "cert pool failed to read file")
	}

	certPool.AppendCertsFromPEM(buf)

	return certPool, nil
}

// loadX509KeyPair load cert and key to keypair
func loadX509KeyPair(certBytes, keyBytes []byte, keyPassword string) (tls.Certificate, error) {
	keyPEMBlock, err := loadPrivateKey(keyBytes, keyPassword)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(certBytes, keyPEMBlock)
}

func loadPrivateKey(keyBytes []byte, password string) ([]byte, error) {
	var v *pem.Block
	var pkey []byte

	for {
		v, keyBytes = pem.Decode(keyBytes)
		if v == nil {
			break
		}
		if !strings.Contains(v.Type, "PRIVATE KEY") {
			continue
		}

		if x509.IsEncryptedPEMBlock(v) {
			pkey, err := x509.DecryptPEMBlock(v, []byte(password))
			if err != nil {
				return nil, errors.Wrap(err, "failed to decrypt private key with password.")
			}

			pkey = pem.EncodeToMemory(&pem.Block{
				Type:  v.Type,
				Bytes: pkey,
			})
		} else {
			pkey = pem.EncodeToMemory(v)
		}

		return pkey, nil
	}
	return nil, errors.New("private key file did not contain private key")
}

// ToTLSConfig creates TLS config
func ToTLSConfig(caCert []byte, certBytes []byte, keyBytes []byte, keyPassword string) (*tls.Config, error) {
	cert, err := loadX509KeyPair(certBytes, keyBytes, keyPassword)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
