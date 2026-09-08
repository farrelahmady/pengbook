// Package validator wraps github.com/go-playground/validator/v10 with a
// concise API for validating DTOs in handlers.
package validator

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the singleton validator instance (safe for concurrent use).
var validate = validator.New()

// Struct validates a struct based on its `validate` tags.
// Returns nil when valid, or an error listing the failing fields.
func Struct(s any) error {
	if err := validate.Struct(s); err != nil {
		return format(err)
	}
	return nil
}

// format converts validator.ValidationErrors into a single human-friendly
// message: "Field is invalid: <tag>; ...".
func format(err error) error {
	var verr validator.ValidationErrors
	if !errors.As(err, &verr) {
		return err
	}

	msgs := make([]string, 0, len(verr))
	for _, fe := range verr {
		msgs = append(msgs, fe.Field()+" is invalid: "+fe.Tag())
	}

	return errors.New(strings.Join(msgs, "; "))
}
