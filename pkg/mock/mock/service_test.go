package mock

import (
	"github.com/aldas/xroad-mock-proxy/pkg/mock/config"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	test_test "github.com/aldas/xroad-mock-proxy/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestNewService(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	service := NewService(&logger, testStorage{})

	assert.Implements(t, (*Service)(nil), service)
}

func TestMockNoMatchingRules(t *testing.T) {
	service := createTestService(config.RuleConfigs{})

	dataBytes := test_test.LoadBytes(t, "rr.rr456.v1/rr456.paring.xml")
	bytes, status := service.mock(dataBytes)

	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, string(bytes), "Rule not found\n")
}

func TestMockMatchingRule(t *testing.T) {
	dataBytes := test_test.LoadBytes(t, "rr.rr456.v1/rr456.paring.xml")

	rules := config.RuleConfigs{
		config.RuleConf{
			Service:        "rr.rr456.v1",
			Priority:       1,
			MatcherRegex:   []string{"(?mi)<isikukood>3821102\\d{4}<\\/isikukood>"},
			IdentityRegex:  "(?mi)<isikukood>(\\d{11})<\\/isikukood>",
			TemplateFile:   "../../../test/testdata/rr.rr456.v1/response.xml",
			Timeout:        "0s",
			ResponseStatus: 200,
		},
	}

	service := createTestService(rules)
	bytes, status := service.mock(dataBytes)

	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(bytes), "<Isik.Isikukood>38211020380</Isik.Isikukood>")
	assert.Contains(t, string(bytes), "<Isik.Eesnimi>Foxtrot</Isik.Eesnimi>")
	assert.Contains(t, string(bytes), "<Isik.Perenimi>Kilo</Isik.Perenimi>")
	assert.Contains(t, string(bytes), "<Isik.Sugu>M</Isik.Sugu>")
}

func createTestService(rules config.RuleConfigs) service {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	r, err := domain.ConvertRules(rules)
	if err != nil {
		logger.Error().Err(err).Msg("failed to ConvertRules")
	}

	service := service{
		logger: &logger,
		storage: testStorage{
			r,
		},
	}

	return service
}

type testStorage struct {
	Rules domain.Rules
}

func (s testStorage) GetAll() domain.Rules {
	return s.Rules
}

func (s testStorage) GetRule(ID int64) (domain.Rule, bool) {
	return domain.Rule{}, false
}
