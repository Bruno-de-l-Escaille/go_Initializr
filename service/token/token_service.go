package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"go_Initializr/pkg/initializer"
)

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   int64  `json:"expires_at"`
}

const redisTokenKey = "gopeople:access_token"

func GetAccessToken(ctx context.Context) (string, error) {
	// Check Redis first
	token, err := initializer.RedisClient.Get(ctx, redisTokenKey).Result()
	if err == nil && token != "" {
		return token, nil
	}

	// If not found, fetch new token
	newToken, err := fetchNewToken()
	if err != nil {
		return "", err
	}

	// Store token with expiration
	expiration := time.Until(time.Unix(newToken.ExpiresAt, 0))
	initializer.RedisClient.Set(ctx, redisTokenKey, newToken.AccessToken, expiration)

	return newToken.AccessToken, nil
}

func fetchNewToken() (*OAuthResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", os.Getenv("GOPEOPLE_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GOPEOPLE_CLIENT_SECRET"))
	data.Set("scope", `[ "users:read", "users:write" ]`)
	goPeopleApiUrl := fmt.Sprintf(os.Getenv("GOPEOPLE_API_URL")+"/oauth/token") 	

	req, err := http.NewRequest("POST",goPeopleApiUrl , bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch token from GoPeople API")
	}

	var result OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
