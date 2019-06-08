package proxy

import (
	"bytes"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	test_test "github.com/aldas/xroad-mock-proxy/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestProxyMatchRule(t *testing.T) {
	var receivedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()

		w.Header().Add("Content-Type", "text/xml;charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		w.Write(test_test.LoadBytes(t, "rr.rr456.v1/response.xml"))
	}))
	defer ts.Close()

	requestBody := bytes.NewBuffer(test_test.LoadBytes(t, "rr.rr456.v1/rr456.paring.xml"))
	req, err := http.NewRequest("POST", XroadDefaulURL, requestBody)
	if err != nil {
		t.Fatal(err)
	}

	serverConfigs := config.ProxyServerConfigs{
		config.ProxyServerConf{
			Address:   "http://localhost:7000",
			Name:      "default",
			IsDefault: true,
		},
		config.ProxyServerConf{
			Address:   ts.URL,
			Name:      "xroad",
			IsDefault: false,
		},
	}

	ruleConfigs := config.RuleConfigs{
		config.RuleConf{
			Server:   "xroad",
			Service:  "rr.RR456.v1",
			Priority: 100,
		},
	}

	recorder := serveWithProxy(t, req, serverConfigs, ruleConfigs)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, string(receivedBody), "<Isikukood>38211020380</Isikukood>")

	response := recorder.Body.String()
	assert.Contains(t, response, "<Isik.Isikukood>{{.Identity}}</Isik.Isikukood>")
}

func serveWithProxy(
	t *testing.T,
	req *http.Request,
	serverConfigs config.ProxyServerConfigs,
	ruleConfigs config.RuleConfigs,
) *httptest.ResponseRecorder {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	servers, err := domain.ConvertProxyServers(serverConfigs)
	if err != nil {
		t.Fatal(err)
	}

	rules, err := domain.ConvertRules(ruleConfigs)
	if err != nil {
		t.Fatal(err)
	}

	service := NewService(&logger, servers, ruleMockService{Rules: rules})

	proxy, err := NewProxyHandler(&logger, service, ruleMockCache{})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	proxy.ServeHTTP(recorder, req)

	return recorder
}

type ruleMockService struct {
	Rules []domain.Rule
}

func (s ruleMockService) GetAll() []domain.Rule {
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
