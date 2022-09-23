package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type mock struct {
	callCount         int
	responses         []*http.Response
	errors            []error
	validators        []Validator
	defaultStatusCode int
}

func (m *mock) CallCount() int { return m.callCount }

func (m *mock) ResetCallCount() { m.callCount = 0 }

type NewMockOpts struct {
	Responses         []*http.Response
	Errors            []error
	Validators        []Validator
	DefaultStatusCode int
}

func NewMock(opts *NewMockOpts) *mock {
	return &mock{
		callCount:         0,
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

	ExpectedCalledWith string
	Fuzzy              bool // instead of exact match check it contians this string
}

func (v *Validator) validate(req *http.Request) error {
	bodyBytes := make([]byte, 0)
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body) // we swallow error so the bodyBytes may be an empty array
	}
	body := string(bodyBytes)
	if v.ExpectedCalledWith != "" {
		if v.Fuzzy {
			if !strings.Contains(body, v.ExpectedCalledWith) {
				return validationError{Reason: fmt.Sprintf("body does not fuzzy contain validator (name: %v) expected body", v.Name)}
			}
		} else {
			if body != v.ExpectedCalledWith {
				return validationError{Reason: fmt.Sprintf("body does not full contain validator (name: %v) expected body", v.Name)}
			}
		}
	}
	if v.ExpectedMethod != "" {
		if v.ExpectedMethod != req.Method {
			return validationError{Reason: fmt.Sprintf("unexpected method (name: %v)", v.Name)}
		}
	}
	if v.ExpectedURLPath != "" {
		path := req.URL.Path
		if v.ExpectedURLPath != path {
			return validationError{Reason: fmt.Sprintf("unexpected URL path (name: %v)", v.Name)}
		}
	}
	return nil
}

func (m *mock) Do(req *http.Request) (*http.Response, error) {
	currAttempt := m.callCount
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
	m.callCount++
	return resp, err
}

type validationError struct {
	Reason string
}

func (e validationError) Error() string {
	return "validation did not pass: " + e.Reason
}
