package proxy

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/server"
	"github.com/rs/zerolog"
	"net/http"
)

type transportSwitcher struct {
	logger        *zerolog.Logger
	Transport     http.RoundTripper
	serverService server.Service
}

// RoundTrip is to use different Transports depending on request url. This is needed in when X-road server uses TLS and
// needs Cert authentication but mock server is using TLS but is just ordinary HTTPS server
// in that case we handle request with matching transport
func (s transportSwitcher) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := s.Transport

	ID := req.Header.Get(requestIDHeader)

	host := req.URL.Host
	hostServer, ok := s.serverService.HostToProxyServer(host)
	if ok && hostServer.Transport != nil {
		transport = hostServer.Transport

		s.logger.Info().
			Str("ID", ID).
			Str("url.host", host).
			Msg("transportSwitcher.RoundTrip")
	}

	if transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}

	return transport.RoundTrip(req)
}
