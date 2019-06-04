package mock

import (
	"bytes"
	"github.com/labstack/echo"
	"io/ioutil"
)

const (
	// XroadDefaulURL is default path where x-road is usually served from security server
	XroadDefaulURL string = "/cgi-bin/consumer_proxy"
)

type controller struct {
	srv Service
}

// RegisterRoutes registers mock routes with server
func RegisterRoutes(srv Service, eg *echo.Group) {
	h := controller{srv}

	eg.POST(XroadDefaulURL, h.mock)
}

// mock returns mocked SOAP response
func (h *controller) mock(c echo.Context) error {
	resp, statusCode := h.srv.mock(extractBody(c))

	return c.Blob(statusCode, "text/xml;charset=UTF-8", resp)
}

func extractBody(c echo.Context) []byte {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}

	// Restore the io.ReadCloser to its original state
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes
}
