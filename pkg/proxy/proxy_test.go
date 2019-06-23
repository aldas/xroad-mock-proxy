package proxy

import (
	"bytes"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/api/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	test_test "github.com/aldas/xroad-mock-proxy/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

// These 2 cert are copied from net.http.internal package. They are used to serve 'httptest.NewTLSServer'
var LocalhostCert = `-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`

func TestProxyMatchRuleToHTTP(t *testing.T) {
	var receivedBody []byte
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()

		w.Header().Add("Content-Type", "text/xml;charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		w.Write(test_test.LoadBytes(t, "rr.rr456.v1/response.xml"))
	}))
	defer mockServer.Close()

	requestBody := bytes.NewBuffer(test_test.LoadBytes(t, "rr.rr456.v1/rr456.paring.xml"))
	req, err := http.NewRequest("POST", XroadDefaulURL, requestBody)
	if err != nil {
		t.Fatal(err)
	}

	servers, err := domain.ConvertProxyServers(config.ProxyServerConfigs{
		config.ProxyServerConf{
			Address:   "http://localhost:7000",
			Name:      "default",
			IsDefault: true,
		},
		config.ProxyServerConf{
			Address:   mockServer.URL,
			Name:      "xroad",
			IsDefault: false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ruleConfigs := config.RuleConfigs{
		config.RuleConf{
			Server:   "xroad",
			Service:  "rr.RR456.v1",
			Priority: 100,
		},
	}

	recorder := serveWithProxy(t, req, servers, ruleConfigs)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, string(receivedBody), "<Isikukood>38211020380</Isikukood>")

	response := recorder.Body.String()
	assert.Contains(t, response, "<Isik.Isikukood>{{.Identity}}</Isik.Isikukood>")
}

func TestProxyMatchRuleToSelfSignedHTTPS(t *testing.T) {
	var receivedBody []byte
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()

		w.Header().Add("Content-Type", "text/xml;charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		w.Write(test_test.LoadBytes(t, "rr.rr456.v1/response.xml"))
	}))
	defer mockServer.Close()

	requestBody := bytes.NewBuffer(test_test.LoadBytes(t, "rr.rr456.v1/rr456.paring.xml"))
	req, err := http.NewRequest("POST", XroadDefaulURL, requestBody)
	if err != nil {
		t.Fatal(err)
	}

	httpsServer, err := dto.ToProxyServer(dto.ServerDTO{
		Name:    "https",
		Address: mockServer.URL,
		TLS: &dto.TLSDTO{
			CACert: LocalhostCert,
			Cert:   LocalhostCert,
			Key:    LocalhostKey,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	servers := domain.ProxyServers{
		domain.ProxyServer{
			Address:   parseURL(t, "http://localhost:7000"),
			Name:      "default",
			IsDefault: true,
		},
		httpsServer,
	}

	ruleConfigs := config.RuleConfigs{
		config.RuleConf{
			Server:   "https",
			Service:  "rr.RR456.v1",
			Priority: 100,
		},
	}

	recorder := serveWithProxy(t, req, servers, ruleConfigs)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, string(receivedBody), "<Isikukood>38211020380</Isikukood>")

	response := recorder.Body.String()
	assert.Contains(t, response, "<Isik.Isikukood>{{.Identity}}</Isik.Isikukood>")
}

func serveWithProxy(
	t *testing.T,
	req *http.Request,
	servers domain.ProxyServers,
	ruleConfigs config.RuleConfigs,
) *httptest.ResponseRecorder {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	rules, err := domain.ConvertRules(ruleConfigs)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := NewProxyHandler(
		&logger,
		serverMockService{servers: servers},
		ruleMockService{Rules: rules},
		ruleMockCache{},
	)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	proxy.ServeHTTP(recorder, req)

	return recorder
}

func parseURL(t *testing.T, toURL string) url.URL {
	parsedURL, err := url.Parse(toURL)
	if err != nil {
		t.Fatal(err)
	}
	return *parsedURL
}

type serverMockService struct {
	servers domain.ProxyServers
}

func (s serverMockService) HostToProxyServer(host string) (domain.ProxyServer, bool) {
	return s.servers.FindByHost(host)
}

func (s serverMockService) DefaultServer() (domain.ProxyServer, bool) {
	return s.servers.Default()
}

func (s serverMockService) Find(name string) (domain.ProxyServer, bool) {
	return s.servers.Find(name)
}

func (s serverMockService) Servers() domain.ProxyServers {
	return s.servers
}

type ruleMockService struct {
	Rules []domain.Rule
}

func (s ruleMockService) GetAll() domain.Rules {
	return s.Rules
}

func (s ruleMockService) GetRule(ID int64) (domain.Rule, bool) {
	return s.Rules[0], true
}

func (s ruleMockService) Save(rule domain.Rule) (domain.Rule, error) {
	return domain.Rule{}, nil
}

func (s ruleMockService) Remove(ID int64) bool {
	return false
}

type ruleMockCache struct {
}

func (c ruleMockCache) Set(req domain.Request) {}
func (c ruleMockCache) Get(ID string) (domain.Request, bool) {
	return domain.Request{}, false
}
func (c ruleMockCache) GetAllIDs() []string {
	return nil
}
func (c ruleMockCache) GetAll() []domain.Request {
	return nil
}
func (c ruleMockCache) DeleteAll() {
	return
}
