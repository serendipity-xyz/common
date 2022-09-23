package strava

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/serendipity-xyz/core/request"
	"github.com/serendipity-xyz/core/storage"
	"github.com/serendipity-xyz/core/types"
)

var (
	httpClient = &http.Client{Timeout: 15 * time.Second}
)

const (
	stravaAPIBaseURL       = "https://www.strava.com/api/v3"
	stravaWebAuthURI       = "https://www.strava.com/oauth/authorize"
	stravaIOSAuthURI       = "strava://oauth/mobile/authorize"
	stravaAndroidAuthURI   = "https://www.strava.com/oauth/mobile/authorize"
	maxUnauthorizedRetries = 1
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int    `json:"expires_at"`
}

type CallContextalizer interface {
	DatabaseManager() storage.Manager
	L() types.Logger
}

type TokenManager interface {
	SetUserTokens(cc CallContextalizer, userID string, tokens Tokens) error
}

// Client represents a strava client that can list and retrieve a user's activities
type Client struct {
	httpClient   request.HTTPClient
	userID       string
	clientID     string
	clientSecret string
	redirectURI  string
	accessToken  string
	refreshToken string
	expiresAt    int
	tokenManager TokenManager
}

type ClientParams struct {
	RedirectURI  string
	ClientID     string
	ClientSecret string
}

// NewClient returns a new Strava client
func NewClient(userID string, tokens Tokens, tokenManager TokenManager, params *ClientParams) *Client {
	return &Client{
		userID:       userID,
		clientID:     params.ClientID,
		clientSecret: params.ClientSecret,
		redirectURI:  params.RedirectURI,
		accessToken:  tokens.AccessToken,
		refreshToken: tokens.RefreshToken,
		expiresAt:    tokens.ExpiresAt,
		tokenManager: tokenManager,
		httpClient:   httpClient,
	}
}

func (sc *Client) SetClient(client request.HTTPClient) {
	sc.httpClient = client
}

// AuthorizationURL returns the redirect URL for strava to authenticate a user
func (sc *Client) AuthorizationURL(scope string) string {
	return fmt.Sprintf("%v?client_id=%v&response_type=code&redirect_uri=%v&approval_prompt=auto&scope=%v", stravaWebAuthURI, sc.clientID, sc.redirectURI, scope)
}

type Athlete struct {
	ID            int         `json:"id"`
	Username      string      `json:"username"`
	ResourceState int         `json:"resource_state"`
	Firstname     string      `json:"firstname"`
	Lastname      string      `json:"lastname"`
	City          string      `json:"city"`
	State         string      `json:"state"`
	Country       interface{} `json:"country"`
	Sex           string      `json:"sex"`
}

type TokenResponse struct {
	TokenType    string  `json:"token_type"`
	ExpiresAt    int     `json:"expires_at"`
	ExpiresIn    int     `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	AccessToken  string  `json:"access_token"`
	Athlete      Athlete `json:"athlete"`
}

// GenerateTokens is used when a user is authenticating via strava. Strava will redirect them to
// our app with a code in the query parameters that allows us to generate new tokens.
func (sc *Client) GenerateTokens(l types.Logger, code string) (TokenResponse, error) {
	url := fmt.Sprintf("%v/oauth/token?client_id=%v&client_secret=%v&code=%v&grant_type=authorization_code", stravaAPIBaseURL, sc.clientID, sc.clientSecret, code)
	var result TokenResponse
	var reason interface{}
	r := request.DefaultR(sc.httpClient).SetResult(&result).SetReason(&reason)
	resp, err := r.Post(url)
	if err != nil {
		l.Error("unable to retrieve strava tokens: %v", err)
		return result, err
	}
	if resp.IsError() {
		e := fmt.Errorf("unable to get Strava auth tokens due to bad status code (%v): %v", resp.StatusCode(), reason)
		l.Error(e.Error())
		return result, e
	}
	return result, nil
}

func (sc *Client) checkExpiry(cc CallContextalizer) bool {
	expiry := time.Unix(int64(sc.expiresAt), 0)
	if expiry.Before(time.Now()) {
		cc.L().Info("detected expired access token [expired_at: %v] [time_since: %v], refreshing...", sc.expiresAt, time.Since(expiry))
		if err := sc.refreshAccessToken(cc); err != nil {
			cc.L().Error("unable to refresh access token: %v", err)
		}
		return true
	}
	return false
}

type Auth struct {
	AthleteID    int    `json:"athleteId" bson:"athleteId"`
	RefreshToken string `json:"refresh_token" bson:"refresh_token"`
	AccessToken  string `json:"access_token" bson:"access_token"`
	ExpiresAt    int    `json:"expires_at" bson:"expires_at"`
}

// @todo figure out how to not rely on rc
func (sc *Client) refreshAccessToken(cc CallContextalizer) error {
	url := fmt.Sprintf("%v/oauth/token?client_id=%v&client_secret=%v&grant_type=refresh_token&refresh_token=%v", stravaAPIBaseURL, sc.clientID, sc.clientSecret, sc.refreshToken)
	var result Auth
	var reason interface{}
	r := request.DefaultR(sc.httpClient).SetResult(&result).SetReason(&reason)
	resp, err := r.Post(url)
	if err != nil {
		cc.L().Error("unable to refresh access token: %v", err)
		return err
	}
	if resp.IsError() {
		e := fmt.Errorf("unable to refresh Strava access token due to bad status code (%v): %v", resp.StatusCode(), reason)
		cc.L().Error(e.Error())
		return e
	}
	tokens := Tokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	}
	if err := sc.tokenManager.SetUserTokens(cc, sc.userID, tokens); err != nil {
		cc.L().Warn("unable to update users access tokens in db: %v", err) // still process the request
	}
	sc.accessToken = tokens.AccessToken
	sc.refreshToken = tokens.RefreshToken
	sc.expiresAt = tokens.ExpiresAt
	return nil
}

type Route string

const (
	WestSideHighwayRoute Route = "WestSideHighway"
)

type Activities []struct {
	Athlete struct {
		ID            int `json:"id"`
		ResourceState int `json:"resource_state"`
	} `json:"athlete"`
	Name            string      `json:"name"`
	Distance        float64     `json:"distance"`
	ElapsedTime     int         `json:"elapsed_time"`
	Type            string      `json:"type"`
	SportType       string      `json:"sport_type"`
	ID              int64       `json:"id"`
	StartDate       time.Time   `json:"start_date"`
	StartDateLocal  time.Time   `json:"start_date_local"`
	Timezone        string      `json:"timezone"`
	UtcOffset       float64     `json:"utc_offset"`
	LocationCity    interface{} `json:"location_city"`
	LocationState   interface{} `json:"location_state"`
	LocationCountry interface{} `json:"location_country"`
	Map             struct {
		ID              string `json:"id"`
		SummaryPolyline string `json:"summary_polyline"`
		ResourceState   int    `json:"resource_state"`
	} `json:"map"`

	Private          bool      `json:"private"`
	Visibility       string    `json:"visibility"`
	StartLatlng      []float64 `json:"start_latlng"`
	EndLatlng        []float64 `json:"end_latlng"`
	SerendipityRoute Route     `json:"serendipity_route"`
}

// ListActivities returns a list of a user's activities
func (sc *Client) ListActivities(cc CallContextalizer) (Activities, error) {
	sc.checkExpiry(cc)
	attempts := 0
	var activites Activities
	var err error
	for attempts <= maxUnauthorizedRetries {
		activites, err = sc.listActivities(cc.L())
		if err != nil {
			if isUnauthorizedErr(err) {
				sc.refreshAccessToken(cc)
				attempts++
				continue
			}
			return activites, err
		}
		return activites, err
	}
	return activites, err
}

func (sc *Client) listActivities(l types.Logger) (Activities, error) {
	url := fmt.Sprintf("%v/athlete/activities?access_token=%v", stravaAPIBaseURL, sc.accessToken)
	var activites Activities
	var reason interface{}
	r := request.DefaultR(sc.httpClient).SetResult(&activites).SetReason(&reason)
	resp, err := r.Get(url)
	if err != nil {
		l.Error("unable to list activities: %v", err)
		return activites, err
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusUnauthorized {
			l.Debug("returning unathorized error to trigger refresh loop")
			return activites, unauthorizedError{}
		}
		e := fmt.Errorf("unable to list strava activities due to bad status code (%v): %v", resp.StatusCode(), reason)
		l.Error(e.Error())
		return activites, e
	}
	return activites, nil
}

type Activity struct {
	ID              int64     `json:"id"`
	ExternalID      string    `json:"external_id"`
	UploadID        int64     `json:"upload_id"`
	Name            string    `json:"name"`
	Distance        float64   `json:"distance"`
	MovingTime      int       `json:"moving_time"`
	ElapsedTime     int       `json:"elapsed_time"`
	Type            string    `json:"type"`
	StartDate       time.Time `json:"start_date"`
	StartDateLocal  time.Time `json:"start_date_local"`
	TimeZone        string    `json:"time_zone"`
	StartLatlng     []float64 `json:"start_latlng"`
	EndLatlng       []float64 `json:"end_latlng"`
	LocationCity    string    `json:"location_city"`
	LocationState   string    `json:"location_state"`
	LocationCountry string    `json:"location_country"`
	Map             struct {
		ID              string `json:"id"`
		Polyline        string `json:"polyline"`
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
	Description string `json:"description"`
}

// GetActivity returns a user's strava activity given an activityId
func (sc *Client) GetActivity(cc CallContextalizer, activityID int64) (*Activity, error) {
	sc.checkExpiry(cc)
	attempts := 0
	var activity *Activity
	var err error
	for attempts <= maxUnauthorizedRetries {
		activity, err = sc.getActivity(cc.L(), activityID)
		if err != nil {
			if isUnauthorizedErr(err) {
				sc.refreshAccessToken(cc)
				attempts++
				continue
			}
			return activity, err
		}
		return activity, err
	}
	return activity, err
}

func (sc *Client) getActivity(l types.Logger, activityID int64) (*Activity, error) {
	url := fmt.Sprintf("%v/activities/%v?access_token=%v", stravaAPIBaseURL, activityID, sc.accessToken)
	var activity *Activity
	var reason interface{}
	r := request.DefaultR(sc.httpClient).SetResult(&activity).SetReason(&reason)
	resp, err := r.Get(url)
	if err != nil {
		l.Error("unable to get strava activity: %v", err)
		return activity, err
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusUnauthorized {
			l.Debug("returning unathorized error to trigger refresh loop")
			return activity, unauthorizedError{}
		}
		e := fmt.Errorf("unable to get strava activity due to bad status code (%v): %v", resp.StatusCode(), reason)
		l.Error(e.Error())
		return activity, e
	}
	return activity, nil
}

type unauthorizedError struct{}

func (ua unauthorizedError) Error() string {
	return "invalid tokens"
}

func (ua unauthorizedError) Code() int {
	return http.StatusUnauthorized
}

func isUnauthorizedErr(err error) bool {
	if _, ok := err.(*unauthorizedError); ok {
		return ok
	}
	if err != nil {
		return strings.Contains(err.Error(), "invalid token")
	}
	return false
}
