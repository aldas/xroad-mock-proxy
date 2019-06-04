package request

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/apperror"
	dto2 "github.com/aldas/xroad-mock-proxy/pkg/proxy/api/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/labstack/echo"
	"net/http"
)

const (
	// RequestPath is route to request API
	RequestPath = "/requests"
)

type controller struct {
	service request.Service
}

// RegisterRoutes registers request routes with server
func RegisterRoutes(srv request.Service, eg *echo.Group) {
	h := controller{srv}

	ur := eg.Group(RequestPath)

	ur.GET("", h.getRequests)
	ur.GET("/:id", h.getRequest)
}

// getRequests returns list of all cached proxy requests
func (h *controller) getRequests(c echo.Context) error {
	reqs := h.service.GetRequests()

	return c.JSON(http.StatusOK, dto2.APIResponse{
		Data:    dto2.RequestsToDTO(reqs),
		Success: true,
	})
}

// getRequest return single request entity
func (h *controller) getRequest(c echo.Context) error {
	reqID := c.Param("id")
	if reqID == "" {
		return echo.ErrNotFound
	}

	req, err := h.service.GetRequest(reqID)
	if err != nil {
		if err == apperror.ErrorNotFound {
			return echo.ErrNotFound
		}
		return err
	}

	return c.JSON(http.StatusOK, dto2.APIResponse{
		Data:    dto2.RequestToFullDTO(req),
		Success: true,
	})
}
