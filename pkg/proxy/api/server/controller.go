package server

import (
	"fmt"
	commonDTO "github.com/aldas/xroad-mock-proxy/pkg/common/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/api/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/server"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

const (
	serversPath = "/servers"
)

type controller struct {
	srv server.Service
}

// RegisterRoutes registers API routes with server
func RegisterRoutes(srv server.Service, eg *echo.Group) {
	h := controller{srv}

	eg.GET(serversPath, h.getAll)
	eg.POST(serversPath, h.addServer)
	eg.PUT(serversPath+"/:id", h.modifyServer)
}

func (h *controller) getAll(c echo.Context) error {
	servers := h.srv.Servers()

	return c.JSON(http.StatusOK, commonDTO.APIResponse{
		Data:    dto.ProxyServersToDTO(servers),
		Success: true,
	})
}

func (h *controller) addServer(c echo.Context) error {
	s, err := extractServer(c)
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}

	s, err = h.srv.Save(s)
	if err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	return c.JSON(http.StatusCreated, commonDTO.APIResponse{
		Data:    dto.ProxyServerToDTO(s),
		Success: true,
	})
}

func (h *controller) modifyServer(c echo.Context) error {
	serverID, err := extractID(c, "id")
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}

	s, err := extractServer(c)
	if err != nil {
		return err
	}
	s.ID = serverID

	s, err = h.srv.Save(s)
	if err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	return c.JSON(http.StatusCreated, commonDTO.APIResponse{
		Data:    dto.ProxyServerToDTO(s),
		Success: true,
	})
}

func extractServer(c echo.Context) (domain.ProxyServer, error) {
	serverDTO := dto.ServerDTO{}
	if err := c.Bind(&serverDTO); err != nil {
		return domain.ProxyServer{}, errors.Wrap(err, "failed to bind payload")
	}

	s, err := dto.ToProxyServer(serverDTO)
	if err != nil {
		return domain.ProxyServer{}, errors.Wrap(err, "failed to convert DTO to domain object")
	}
	return s, nil
}

func extractID(c echo.Context, param string) (int64, error) {
	raw := c.Param(param)
	if raw == "" {
		return 0, echo.ErrNotFound
	}

	ID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("%v must be int value", param))
	}

	return ID, err
}
