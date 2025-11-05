package validator

import (
	"net"
	"net/mail"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

func validateServiceName(fl validator.FieldLevel) bool {
	serviceName := fl.Field().String()

	return len(serviceName) != 0
}

func validateClientName(fl validator.FieldLevel) bool {
	const maxClientNameLength = 100

	clientName := fl.Field().String()

	return len(clientName) < maxClientNameLength
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}

	domain := strings.Split(email, "@")[1]
	mxRecords, err := net.LookupMX(domain)

	return err == nil && len(mxRecords) > 0
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	const usaCountry = "US"

	phoneNumber := fl.Field().String()

	_, err := phonenumbers.Parse(phoneNumber, usaCountry)

	return err == nil
}

func validateAddress(fl validator.FieldLevel) bool {
	const maxAddressLength = 255

	address := fl.Field().String()

	return len(address) <= maxAddressLength
}
