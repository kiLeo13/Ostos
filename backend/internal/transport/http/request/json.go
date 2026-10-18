package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"ostos/backend/internal/apierrors"
	"ostos/backend/internal/transport/http/validation"
)

// DecodeJSON decodes exactly one JSON value, rejects unknown fields, then collects all validation failures.
func DecodeJSON(request *http.Request, destination any, validator *validation.Validator) error {
	if request == nil || request.Body == nil {
		return apierrors.MalformedJSON(errors.New("request body is required"))
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(request.Header.Get("Content-Type"), ";")[0]))
	if mediaType != "application/json" {
		return apierrors.UnsupportedMediaType()
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		if isBodyTooLarge(err) {
			return apierrors.RequestTooLarge(err)
		}
		return apierrors.MalformedJSON(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return apierrors.MalformedJSON(errors.New("request body must contain exactly one JSON value"))
	}

	if fields := validator.Struct(destination); len(fields) > 0 {
		return apierrors.InvalidFormBody(fields)
	}
	return nil
}

func isBodyTooLarge(err error) bool {
	var maxBytesError *http.MaxBytesError
	return errors.As(err, &maxBytesError)
}
