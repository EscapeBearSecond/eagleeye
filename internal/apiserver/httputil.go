package apiserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Code    int
	Message string
	Data    interface{}
}

func OK(c echo.Context, data ...interface{}) error {
	if len(data) != 0 && data[0] != nil {
		return c.JSON(http.StatusOK, &Response{
			Code:    0,
			Message: "success",
			Data:    data[0],
		})
	}
	return c.JSON(http.StatusOK, &Response{
		Code:    0,
		Message: "success",
		Data:    nil,
	})
}

func Handle[T any, U any](f func(context.Context, *T) (*U, error)) func(echo.Context) error {

	return func(c echo.Context) error {
		var request T

		if err := c.Bind(&request); err != nil {
			return NewUnprocessableEntityError(nil, err)
		}

		if err := c.Validate(&request); err != nil {
			return NewBadRequestErrorM(err.Error())
		}

		replay, err := f(c.Request().Context(), &request)
		if err != nil {
			return err
		}

		return OK(c, replay)
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var e *Error
	if !errors.As(err, &e) {
		e = NewInternalServerError(nil)
	}

	if c.Request().Method == http.MethodHead {
		err = c.NoContent(e.HTTPCode)
	} else {
		err = c.JSON(e.HTTPCode, &Response{
			Code:    e.Status.Code,
			Message: e.Status.Message,
		})
	}
	if err != nil {
		Logger.Error("Error response failed", "error", err)
	}
}
