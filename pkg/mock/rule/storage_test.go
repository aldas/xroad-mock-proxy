package rule

import (
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/bluele/gcache"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSaveExistingRule(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	storage := cacheStorage{
		logger: &logger,
		cache:  gcache.New(2).Simple().Build(),
	}

	rule := domain.Rule{
		ID:         0,
		IsReadOnly: false,
	}
	r, err := storage.Save(rule)
	if err != nil {
		t.Fatalf("new rule save failed: %v", err)
	}
	assert.NotEqual(t, int64(0), r.ID)

	r2, err := storage.Save(r)
	if err != nil {
		t.Fatalf("existing rule save failed: %v", err)
	}
	assert.Equal(t, r.ID, r2.ID)
	assert.ObjectsAreEqual(r, r2)
}
