package strava_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/serendipity-xyz/common/log"
	"github.com/serendipity-xyz/common/mocks"
	"github.com/serendipity-xyz/common/strava"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

type MockUserService struct{}

func (mock *MockUserService) SetUserTokens(cc strava.CallContextalizer, userID string, tokens strava.Tokens) error {
	return nil
}

func TestAuthorizationURL(t *testing.T) {
	sc := strava.NewClient("mockUserID", strava.Tokens{}, &MockUserService{}, &strava.ClientParams{
		ClientID:    "mockClientId",
		RedirectURI: "mockRedirecturi",
	})
	url := sc.AuthorizationURL("mockScope", strava.Web)
	require.Equal(t, "https://www.strava.com/oauth/authorize?client_id=mockClientId&response_type=code&redirect_uri=mockRedirecturi&approval_prompt=auto&scope=mockScope", url, "urls should match [0]")
}

func TestCanGenerateTokens(t *testing.T) {
	stravaClient := strava.NewClient("mockUserId", strava.Tokens{}, &MockUserService{}, &strava.ClientParams{})
	mc := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
					"token_type": "test",
					"expires_at": 103, 
					"expires_in": 3,
					"refresh_token": "mockRefreshToken",
					"access_token": "accessToken",
					"athlete": {
						"id": 23,
						"username": "mockUser",
						"resource_state": 2,
						"firstname": "myFirstname",
						"lastname": "myLastname",
						"city": "myCity",
						"state": "myState",
						"country": "myCountry",
						"sex": "male"
					}
				}
			`))),
			},
		},
	})
	stravaClient.SetClient(mc)
	res, err := stravaClient.GenerateTokens(log.StdOutLogger{}, "mockCode")
	require.Nil(t, err, "no error")
	require.Equal(t, strava.TokenResponse{
		TokenType:    "test",
		ExpiresAt:    103,
		ExpiresIn:    3,
		RefreshToken: "mockRefreshToken",
		AccessToken:  "accessToken",
		Athlete: strava.Athlete{
			ID:            23,
			Username:      "mockUser",
			ResourceState: 2,
			Firstname:     "myFirstname",
			Lastname:      "myLastname",
			City:          "myCity",
			State:         "myState",
			Country:       "myCountry",
			Sex:           "male",
		},
	}, res, "unexpected result")
	require.Equal(t, 1, mc.CallCount(), "only one call")
}

func TestCanListActivities(t *testing.T) {
	t.Skip("todo")
}

func TestCanGetActivity(t *testing.T) {
	t.Skip("todo")
}

func TestCanRetry500s(t *testing.T) {
	t.Skip("todo")
}

func TestCanRetryExpiredTokens(t *testing.T) {
	t.Skip("todo")
}

func TestCanRetryUnauthorizedErrs(t *testing.T) {
	t.Skip("todo")
}

func TestMax500Retrys(t *testing.T) {
	t.Skip("todo")
}

func TestMaxUnathorizedRetrys(t *testing.T) {
	t.Skip("todo")
}
