// Filename: internal/validator/validator.go

package validator

import (
	"net/url"
	"regexp"
)

var (
	EmailRx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9-](?:[a-zA-Z0-9-]+)*$")

	PhoneRx = regexp.MustCompile(`^\+?\(?[0-9]{3}\)?\s?-\s?[0-9]{3}\s?-s?[0-9]{4}$`)
)

// Create a type that wraps our validation errors map
type Validator struct {
	Errors map[string]string
}

// New() creates a new validator instance
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Method called valid() because we want to have access to the validator type
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// In() checks if an element can be found in a provided list of elements
func In(element string, list ...string) bool {
	for i := range list {
		if element == list[i] {
			return true
		}
	}
	return false
}

// Matches() returns true if a string value matches a specific regexp pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// ValidWebsite() checks if a string value is a valid website URL
func ValidWebsite(website string) bool {
	_, err := url.ParseRequestURI(website)
	return err == nil
}

// AddError() method that adds an error entry to the errors map
func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; exist {
		v.Errors[key] = message
	}
}

// Check() method performas the validation checks and calls the AddError
// methofin turn if an error entry needs to be added
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Unique() method that ensures that the entries in the mode does not repeat
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
