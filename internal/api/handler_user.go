package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

func (s *State) HandlerCreateUser(ctx context.Context, username, password string) error {
	if err := validateCredentials(username, password); err != nil {
		return err
	}
	credentialsReader, err := createReaderFromStruct(loginCredentials{Username: username, Password: password})
	if err != nil {
		return err
	}

	// create and make the request
	url := s.Server.BaseURL + s.Server.CreateUser
	req, err := http.NewRequestWithContext(ctx, "POST", url, credentialsReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		return fmt.Errorf("could not create the user on endpoint %s: %v", url, res.StatusCode)
	}

	return nil
}

func createReaderFromStruct(val any) (*bytes.Buffer, error) {
	valMarshal, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(valMarshal), nil
}

func validateCredentials(username, password string) error {
	if len(username) < 10 {
		return ValidationError{"The username must be at least 10 characters long."}
	}
	if len(username) < 10 || len(password) < 10 {
		return ValidationError{"The password must be at least 10 characters long."}
	}
	return nil
}
