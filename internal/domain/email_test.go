package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

var validEmails = []string{
	"bruce.wayne@wayneenterprises.com",
	"clark.kent@dailyplanet.com",
	"peter.parker@dailybugle.org",
	"tony.stark+tech@starkindustries.com",
	"natasha.romanoff@shield.gov",
	"steve_rogers@avengers.io",
	"arthur.curry@atlantis.co.uk",
	"wade.wilson@example.co",
	"diana.prince@themiscira.org",
	"lee.smith123@example-domain.com",
	"user+tag@sub.mail.example.com",
	"firstname-lastname@company.name",
	"a@b.co",
}

func TestNewEmail_ShouldReturnEmail_WhenInputIsValid(t *testing.T) {
	for _, input := range validEmails {
		t.Run(input, func(t *testing.T) {
			email, err := domain.NewEmail(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, email)
			assert.Equal(t, input, email.String())
		})
	}
}

func TestNewEmail_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	email, err := domain.NewEmail(input)

	assert.Empty(t, email)
	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrEmailEmpty.Error())
}

func TestNewEmail_ShouldReturnInvalidFormatError_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc  string
		input string
	}{
		{desc: "email with spaces", input: "spider man @ oscorp.com"},
		{desc: "email without @", input: "doctor-strange.magic.io"},
		{desc: "email with @ at the end", input: "barry-allen@starlab@"},
		{desc: "email with @ at the beginning", input: "@robin@batmail.com"},
		{desc: "domain without dot", input: "batman@batmail"},
		{desc: "domain with dot at the end", input: "superman@super.net."},
		{desc: "domain with dot at the beginning", input: "odin@.asgard.net"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			email, err := domain.NewEmail(tC.input)
			assert.Empty(t, email)
			assert.Error(t, err)
			assert.EqualError(t, err, domain.ErrEmailInvalidFormat.Error())
		})
	}
}

func TestValidateEmail_ShouldReturnExpectedErrors(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []error
	}{
		{
			desc:     "email empty",
			input:    "",
			expected: []error{domain.ErrEmailEmpty},
		},
		{
			desc:     "email with invalid format",
			input:    "invalid",
			expected: []error{domain.ErrEmailInvalidFormat},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := domain.ValidateEmail(tC.input)
			assert.NotEmpty(t, errs)
			assert.Equal(t, tC.expected, errs)
		})
	}
}

func TestValidateEmail_ShouldReturnEmpty_WhenInputIsValid(t *testing.T) {
	for _, input := range validEmails {
		t.Run(input, func(t *testing.T) {
			errs := domain.ValidateEmail(input)
			assert.Empty(t, errs)
		})
	}
}
