package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type mock struct {
	attempts          int
	responses         []*http.Response
	errors            []error
	validators        []Validator
	defaultStatusCode int
}

type NewMockOpts struct {
	Responses         []*http.Response
	Errors            []error
	Validators        []Validator
	DefaultStatusCode int
}

func NewMock(opts *NewMockOpts) *mock {
	return &mock{
		attempts:          0,
		responses:         opts.Responses,
		errors:            opts.Errors,
		defaultStatusCode: opts.DefaultStatusCode,
		validators:        opts.Validators,
	}
}

type Validator struct {
	Name string // easier for identification on error

	ExpectedURLPath string
	ExpectedMethod  string

	ExpectedStatusCode int

	ExpectedBody string
	BodyContains bool // instead of exact match check it contians this string

	ExpectedCalledWith string
	CalledWithContains bool // instead of exact match check it contians this string
}

func (v *Validator) validate(req *http.Request) error {
	bodyBytes, _ := ioutil.ReadAll(req.Body) // we swallow error so the bodyBytes may be an empty array
	body := string(bodyBytes)
	if v.ExpectedBody != "" {
		if v.BodyContains {
			if !strings.Contains(body, v.ExpectedBody) {
				return validationError{Reason: fmt.Sprintf("body does not fuzzy contain validator (name: %v) expected body", v.Name)}
			} else {
				if body != v.ExpectedBody {
					return validationError{Reason: fmt.Sprintf("body does not full contain validator (name: %v) expected body", v.Name)}
				}
			}
		}
	}
	if v.ExpectedMethod != "" {
		// TODO: finish handling all the validation methods and them implement the unit tests to make
		// use of these validators
	}
	return validationError{Reason: "invalid"}
}

func (m *mock) Do(req *http.Request) (*http.Response, error) {
	currAttempt := m.attempts
	var resp *http.Response
	var err error
	if currAttempt < len(m.responses) {
		resp = m.responses[currAttempt]
	}
	if currAttempt < len(m.errors) {
		err = m.errors[currAttempt]
	}
	if currAttempt < len(m.validators) {
		err = m.validators[currAttempt].validate(req)
	}
	m.attempts++
	return resp, err
}

type validationError struct {
	Reason string
}

func (e validationError) Error() string {
	return "validation did not pass: " + e.Reason
}
