package rule

import (
	"fmt"
	commonDTO "github.com/aldas/xroad-mock-proxy/pkg/common/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/api/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

const (
	rulesPath = "/rules"
)

type controller struct {
	srv rule.Service
}

// RegisterRoutes registers API routes with server
func RegisterRoutes(srv rule.Service, eg *echo.Group) {
	h := controller{srv}

	eg.GET(rulesPath, h.getAll)
	eg.POST(rulesPath, h.addRule)
	eg.PUT(rulesPath+"/:id", h.modifyRule)
	eg.GET(rulesPath+"/:id", h.getRule)
	eg.DELETE(rulesPath+"/:id", h.removeRule)
}

func (h *controller) getAll(c echo.Context) error {
	rules := h.srv.GetAll()

	return c.JSON(http.StatusOK, commonDTO.APIResponse{
		Data:    dto.RulesToDTO(rules),
		Success: true,
	})
}

func (h *controller) getRule(c echo.Context) error {
	ruleID, err := extractID(c, "id")
	if err != nil {
		return err
	}

	r, ok := h.srv.GetRule(ruleID)
	if !ok {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, commonDTO.APIResponse{
		Data:    dto.RuleToDTO(r),
		Success: true,
	})
}

func (h *controller) addRule(c echo.Context) error {
	r, err := extractRule(c)
	if err != nil {
		return err
	}

	r, err = h.srv.Save(r)
	if err != nil {
		return errors.Wrap(err, "failed to persist rule")
	}

	return c.JSON(http.StatusCreated, commonDTO.APIResponse{
		Data:    dto.RuleToDTO(r),
		Success: true,
	})
}

func (h *controller) modifyRule(c echo.Context) error {
	ruleID, err := extractID(c, "id")
	if err != nil {
		return err
	}

	r, err := extractRule(c)
	if err != nil {
		return err
	}
	r.ID = ruleID

	r, err = h.srv.Save(r)
	if err != nil {
		return errors.Wrap(err, "failed to persist rule")
	}

	return c.JSON(http.StatusOK, commonDTO.APIResponse{
		Data:    dto.RuleToDTO(r),
		Success: true,
	})
}

func (h *controller) removeRule(c echo.Context) error {
	ruleID, err := extractID(c, "id")
	if err != nil {
		return err
	}

	if ok := h.srv.Remove(ruleID); !ok {
		return errors.Wrap(err, "failed to remove rule")
	}

	return c.JSON(http.StatusOK, commonDTO.APIResponse{
		Data:    nil,
		Success: true,
	})
}

func extractRule(c echo.Context) (domain.Rule, error) {
	ruleDTO := dto.RuleDTO{}
	if err := c.Bind(&ruleDTO); err != nil {
		return domain.Rule{}, errors.Wrap(err, "failed to bind payload")
	}

	r, err := dto.ToRule(ruleDTO)
	if err != nil {
		return domain.Rule{}, errors.Wrap(err, "failed to convert DTO to domain object")
	}
	return r, nil
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
