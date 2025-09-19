package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Register custom validation for password regex
	validate.RegisterValidation("regexp", validateRegex)
}

func validateRegex(fl validator.FieldLevel) bool {
	pattern := fl.Param()
	value := fl.Field().String()
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func IsDefaultPassword(password string) bool {
	return password == "telkom@2025"
}
