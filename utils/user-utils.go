package utils

import "github.com/go-passwd/validator"

func PasswordValidator(password string) bool {
	validatePassword := validator.New(validator.MinLength(6, nil), validator.ContainsAtLeast("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 1, nil), validator.ContainsAtLeast("0123456789", 1, nil))

	err := validatePassword.Validate(password)
	if err != nil {
		return false
	}

	return true
}
