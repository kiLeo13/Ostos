package validation

import (
	"reflect"
	"regexp"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// Validator is the single application-wide validation engine.
type Validator struct {
	engine *validator.Validate
}

func New() (*Validator, error) {
	engine := validator.New(validator.WithRequiredStructEnabled())
	engine.RegisterTagNameFunc(jsonFieldName)
	if err := engine.RegisterValidation("username", validateUsername); err != nil {
		return nil, err
	}
	return &Validator{engine: engine}, nil
}

func (instance *Validator) Struct(value any) map[string]string {
	if instance == nil || instance.engine == nil {
		return map[string]string{"body": "Validation is unavailable"}
	}

	err := instance.engine.Struct(value)
	if err == nil {
		return nil
	}

	fields := make(map[string]string)
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		fields["body"] = "The request body is invalid"
		return fields
	}
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		if _, exists := fields[field]; exists {
			continue
		}
		fields[field] = messageFor(field, fieldError)
	}
	return fields
}

func (instance *Validator) Engine() *validator.Validate { return instance.engine }

func validateUsername(field validator.FieldLevel) bool {
	value := field.Field().String()
	return len(value) >= 2 && len(value) <= 32 && usernamePattern.MatchString(value)
}

func jsonFieldName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}

func messageFor(field string, fieldError validator.FieldError) string {
	if field == "username" && (fieldError.Tag() == "username" || fieldError.Tag() == "min" || fieldError.Tag() == "max") {
		return "Use 2-32 lowercase letters, digits, hyphens, or underscores"
	}
	if field == "email" && fieldError.Tag() == "email" {
		return "Enter a valid email address"
	}

	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field is too short"
	case "max":
		return "This field is too long"
	case "oneof":
		return "Choose a supported value"
	case "gt":
		return "Enter a value greater than zero"
	default:
		return "This field is invalid"
	}
}
