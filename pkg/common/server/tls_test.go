package server

import (
	"crypto/tls"
	test_test "github.com/aldas/xroad-mock-proxy/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadX509KeyPairWithPassword(t *testing.T) {
	certBytes := test_test.LoadBytes(t, "certificates/password_proxy_server_cert.pem")
	keyBytes := test_test.LoadBytes(t, "certificates/password_proxy_server_key.pem")

	cert, err := loadX509KeyPair(certBytes, keyBytes, "SuperSecret1")
	assert.NoError(t, err)
	assert.IsType(t, tls.Certificate{}, cert)
}

func TestLoadX509KeyPairWithoutPassword(t *testing.T) {
	certBytes := test_test.LoadBytes(t, "certificates/passwordless_proxy_server_cert.pem")
	keyBytes := test_test.LoadBytes(t, "certificates/passwordless_proxy_server_key.pem")

	cert, err := loadX509KeyPair(certBytes, keyBytes, "")
	assert.NoError(t, err)
	assert.IsType(t, tls.Certificate{}, cert)
}
