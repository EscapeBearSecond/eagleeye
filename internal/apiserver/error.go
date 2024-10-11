package apiserver

import (
	"fmt"
	"net/http"
	"runtime"
)

type Error struct {
	HTTPCode int
	Status   *status
	Internal error
}

type status struct {
	Code    int
	Message string
}

const (
	CodePlanNotFound        = 1000
	CodePlanResultsNotFound = 1001
)

var errMsg = map[int]string{
	CodePlanNotFound:        "plan not found",
	CodePlanResultsNotFound: "plan results not found",
}

var (
	ErrPlanNotFound        = NewNotFoundError(Status(CodePlanNotFound))
	ErrPlanResultsNotFound = NewNotFoundError(Status(CodePlanResultsNotFound))
)

func Status(code int, message ...string) *status {
	msg := errMsg[code]
	if len(message) != 0 && message[0] != "" {
		msg = message[0]
	}

	return &status{Code: code, Message: msg}
}

func NewError(httpCode int, status *status, internal ...error) *Error {
	err := &Error{}
	if len(internal) != 0 && internal[0] != nil {
		err.Internal = internal[0]
	}
	err.HTTPCode = httpCode
	if status == nil {
		status = Status(0)
	}
	if status.Code == 0 {
		status.Code = httpCode
	}
	if status.Message == "" {
		status.Message = http.StatusText(httpCode)
	}
	err.Status = status
	return err
}

func WithCaller(internal ...error) error {
	var cause error
	if len(internal) != 0 && internal[0] != nil {
		cause = internal[0]
	} else {
		return nil
	}

	pc, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s[%s:%d]", runtime.FuncForPC(pc).Name(), file, line)

	return fmt.Errorf("%s: %w", caller, cause)
}

func (e *Error) clone() *Error {
	newE := &Error{}
	newE.HTTPCode = e.HTTPCode
	newE.Status = Status(e.Status.Code, e.Status.Message)
	newE.Internal = e.Internal
	return newE
}

func (e *Error) WithCause(err error) *Error {
	newE := e.clone()
	newE.Internal = err
	return newE
}

func (e *Error) WithMessage(message string) *Error {
	newE := e.clone()
	newE.Status.Message = message
	return newE
}

func (e *Error) Unwrap() error {
	return e.Internal
}

func (e *Error) Error() string {
	if e.Internal == nil {
		return fmt.Sprintf("%d[%s]", e.Status.Code, e.Status.Message)
	}
	return fmt.Sprintf("%d[%s]: %s", e.Status.Code, e.Status.Message, e.Internal)
}

func NewNotFoundError(status *status, internal ...error) *Error {
	return NewError(http.StatusNotFound, status, internal...)
}

func NewNotFoundErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusNotFound, Status(0, message), internal...)
}

func NewBadRequestError(status *status, internal ...error) *Error {
	return NewError(http.StatusBadRequest, status, internal...)
}

func NewBadRequestErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusBadRequest, Status(0, message), internal...)
}

func NewInternalServerError(status *status, internal ...error) *Error {
	return NewError(http.StatusInternalServerError, status, internal...)
}

func NewInternalServerErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusInternalServerError, Status(0, message), internal...)
}

func NewUnauthorizedError(status *status, internal ...error) *Error {
	return NewError(http.StatusUnauthorized, status, internal...)
}

func NewUnauthorizedErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusUnauthorized, Status(0, message), internal...)
}

func NewForbiddenError(status *status, internal ...error) *Error {
	return NewError(http.StatusForbidden, status, internal...)
}

func NewForbiddenErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusForbidden, Status(0, message), internal...)
}

func NewNotImplementedError(status *status, internal ...error) *Error {
	return NewError(http.StatusNotImplemented, status, internal...)
}

func NewNotImplementedErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusNotImplemented, Status(0, message), internal...)
}

func NewNotAcceptableError(status *status, internal ...error) *Error {
	return NewError(http.StatusNotAcceptable, status, internal...)
}

func NewNotAcceptableErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusNotAcceptable, Status(0, message), internal...)
}

func NewConflictError(status *status, internal ...error) *Error {
	return NewError(http.StatusConflict, status, internal...)
}

func NewConflictErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusConflict, Status(0, message), internal...)
}

func NewGoneError(status *status, internal ...error) *Error {
	return NewError(http.StatusGone, status, internal...)
}

func NewGoneErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusGone, Status(0, message), internal...)
}

func NewUnprocessableEntityError(status *status, internal ...error) *Error {
	return NewError(http.StatusUnprocessableEntity, status, internal...)
}

func NewUnprocessableEntityErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusUnprocessableEntity, Status(0, message), internal...)
}

func NewSeviceUnavailableError(status *status, internal ...error) *Error {
	return NewError(http.StatusServiceUnavailable, status, internal...)
}

func NewSeviceUnavailableErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusServiceUnavailable, Status(0, message), internal...)
}

func NewTooManyRequestsError(status *status, internal ...error) *Error {
	return NewError(http.StatusTooManyRequests, status, internal...)
}

func NewTooManyRequestsErrorM(message string, internal ...error) *Error {
	return NewError(http.StatusTooManyRequests, Status(0, message), internal...)
}
