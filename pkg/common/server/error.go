package server

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
)

type customErrHandler struct {
	e *echo.Echo
}

func (ce *customErrHandler) handler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	type resp struct {
		Success bool        `json:"success"`
		Message interface{} `json:"message"`
	}

	if ce.e.Debug {
		msg = err.Error()
		switch err.(type) {
		case *echo.HTTPError:
			e := err.(*echo.HTTPError)
			code = e.Code
			msg = e.Message
		}
	} else {
		switch err.(type) {
		case *echo.HTTPError:
			e := err.(*echo.HTTPError)
			code = e.Code
			msg = e.Message
			if e.Internal != nil {
				msg = fmt.Sprintf("%v, %v", err, e.Internal)
			}
		default:
			msg = http.StatusText(code)
		}
	}

	if _, ok := msg.(string); ok {
		msg = resp{Message: msg}
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == "HEAD" {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			ce.e.Logger.Error(err)
		}
	}
}
