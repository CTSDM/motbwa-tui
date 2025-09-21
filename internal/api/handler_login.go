package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type loginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type responseVals struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (s *State) HandlerLogin(ctx context.Context, username, password string) error {
	credentialsMarshal, err := json.Marshal(loginCredentials{Username: username, Password: password})
	if err != nil {
		return fmt.Errorf("could not marshal the input credentials: %w", err)
	}
	credentialsBuffer := bytes.NewBuffer(credentialsMarshal)

	resLogin, err := makeLoginRequest(ctx, s.Server.BaseURL+s.Server.Login, credentialsBuffer)
	if err != nil {
		return fmt.Errorf("failed while calling the login endpoint: %w", err)
	}

	// assign login credentials to the state variable
	s.User.UserID = resLogin.ID
	s.User.Username = resLogin.Username
	s.User.Token = resLogin.Token
	s.User.RefreshToken = resLogin.RefreshToken

	return nil
}

func makeLoginRequest(ctx context.Context, url string, credentials io.Reader) (responseVals, error) {
	// create the request
	req, err := http.NewRequestWithContext(ctx, "POST", url, credentials)
	if err != nil {
		return responseVals{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	// make the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return responseVals{}, err
	}

	// check code status
	if res.StatusCode > 201 {
		return responseVals{}, fmt.Errorf("server responded with status on endpoint %s: %v", url, res.Status)
	}

	// parse the response
	var resVals responseVals
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&resVals); err != nil {
		return responseVals{}, err
	}
	return resVals, nil
}
