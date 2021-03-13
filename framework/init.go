package framework

import "github.com/go-playground/validator/v10"

var Validator *validator.Validate // validates variables and structs to enforce rules

/*
 * Initialize globals.
 */
func init() {
	Validator = validator.New()
}
