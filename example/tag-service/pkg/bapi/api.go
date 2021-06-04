package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	APP_KEY    = "eddycjy"
	APP_SECRET = "go-programming-tour-book"
)

type AccessToken struct {
	Token string `json:"token"`
}

func (a *API) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"%s?app_key=%s&app_secret=%s",
		"auth",
		APP_KEY,
		APP_SECRET,
	)
	body, err := a.httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var accessToken AccessToken
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}
	return accessToken.Token, nil
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", a.URL, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type API struct {
	URL string
}

func NewAPI(url string) *API {
	return &API{
		URL: url,
	}
}

func (a API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}
	return body, nil
}
