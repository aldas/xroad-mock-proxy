package mock

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockService struct {
	Payload []byte
}

func (s *mockService) mock(requestBody []byte) ([]byte, int) {
	s.Payload = requestBody
	return []byte("SOAP"), 200
}

func TestMock(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("PAYLOAD"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(XroadDefaulURL)

	service := mockService{}
	h := &controller{&service}

	if assert.NoError(t, h.mock(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []string{"text/xml;charset=UTF-8"}, rec.Header()["Content-Type"])
		assert.Equal(t, "SOAP", rec.Body.String())
		assert.Equal(t, "PAYLOAD", string(service.Payload))
	}
}
