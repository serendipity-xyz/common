package request_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/serendipity-xyz/core/request"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

type mockClient struct {
	attempts  *int
	responses []*http.Response
	errors    []error
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	currAttempt := *c.attempts
	var resp *http.Response
	var err error
	if currAttempt < len(c.responses) {
		resp = c.responses[currAttempt]
	}
	if currAttempt < len(c.errors) {
		err = c.errors[currAttempt]
	}
	if req.Method == http.MethodPost {
		resp.Body = req.Body
	}
	*c.attempts++
	return resp, err
}

func TestSuccessfulGet(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"hello": "test", "world": 3}`))),
			},
		},
		errors: []error{},
	}
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL")
	require.Nil(t, err, "no error on get expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 1, callCount, "call count")
	require.Equal(t, map[string]interface{}{
		"hello": "test",
		"world": float64(3),
	}, res, "expected output to be equal")
}

func TestFailedGet(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error": "myError"}`))),
			},
		},
		errors: []error{},
	}
	var res map[string]interface{}
	var reason map[string]interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL")
	require.NotNil(t, err, "expected an error")
	require.IsType(t, request.BadStatusError{}, err, "expected error type")
	require.True(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 1, callCount, "call count")
	require.Equal(t, map[string]interface{}{
		"error": "myError",
	}, reason, "expected output to be equal")
}

func TestRetryGet(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"hello": "test", "world": 3}`))),
			},
		},
		errors: []error{},
	}
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL")
	require.Nil(t, err, "no error on get expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 2, callCount, "call count")
	require.Equal(t, map[string]interface{}{
		"hello": "test",
		"world": float64(3),
	}, res, "expected output to be equal")
}

func TestMaxRetriesExhausted(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
		},
		errors: []error{},
	}
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason)
	_, err := r.Get("mockURL")
	require.NotNil(t, err, "expected err")
	require.Equal(t, "max retries exhausted", err.Error(), "err check")
}

func TestSuccessfulPost(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
		},
		errors: []error{},
	}
	var res interface{}
	var reason interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason).SetBody("youhoo")
	resp, err := r.Post("mockURL")
	require.Nil(t, err, "no error on post expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 1, callCount, "call count")
	require.Equal(t, "youhoo", res, "expected output to be equal")
}

func TestSuccessfulPostRetry(t *testing.T) {
	callCount := 0
	httpClient := mockClient{
		attempts: &callCount,
		responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
		},
		errors: []error{},
	}
	var res interface{}
	var reason interface{}
	r := request.DefaultR(&httpClient).SetResult(&res).SetReason(&reason).SetBody("youhoo")
	resp, err := r.Post("mockURL")
	require.Nil(t, err, "no error on post expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 2, callCount, "call count")
	require.Equal(t, "youhoo", res, "expected output to be equal")
}
