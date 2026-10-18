package apierrors

import (
	"errors"
	"net/http"
)

const (
	CodeMalformedJSON        = "MALFORMED_JSON"
	CodeInvalidFormBody      = "INVALID_FORM_BODY"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeRequestTooLarge      = "REQUEST_TOO_LARGE"
	CodeNotFound             = "NOT_FOUND"
	CodeInternal             = "INTERNAL_ERROR"
	CodeUnavailable          = "SERVICE_UNAVAILABLE"
)

// Error is safe for transport. Cause remains server-side and is never serialized.
type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Cause   error
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	return err.Code + ": " + err.Message
}

func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type Envelope struct {
	Errors Detail `json:"errors"`
}

type Detail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func MalformedJSON(cause error) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Code:    CodeMalformedJSON,
		Message: "The request body must contain valid JSON",
		Cause:   cause,
	}
}

func InvalidFormBody(fields map[string]string) *Error {
	return &Error{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeInvalidFormBody,
		Message: "One or more fields are invalid",
		Fields:  fields,
	}
}

func UnsupportedMediaType() *Error {
	return New(http.StatusUnsupportedMediaType, CodeUnsupportedMediaType, "The request Content-Type must be application/json")
}

func RequestTooLarge(cause error) *Error {
	return &Error{
		Status:  http.StatusRequestEntityTooLarge,
		Code:    CodeRequestTooLarge,
		Message: "The request body is too large",
		Cause:   cause,
	}
}

func Public(err error) *Error {
	var apiError *Error
	if errors.As(err, &apiError) {
		return apiError
	}
	return New(http.StatusInternalServerError, CodeInternal, "An internal error occurred")
}

func ToEnvelope(err error) Envelope {
	public := Public(err)
	return Envelope{Errors: Detail{Code: public.Code, Message: public.Message, Fields: public.Fields}}
}
