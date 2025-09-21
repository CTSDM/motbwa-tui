package api

import (
	"net/http"

	"github.com/google/uuid"
)

type State struct {
	User   UserInfo
	Server ServerInfo
}

type UserInfo struct {
	UserID       uuid.UUID
	Username     string
	Token        string
	RefreshToken string
}

type ServerInfo struct {
	BaseURL      string
	WebsocketURL string
	Login        string
	Users        string
}

func (s State) AddAuthTokensToHeader(header *http.Header) {
	header.Set("Auth", "Bearer "+s.User.Token)
	header.Set("X-Refresh-Token", "Token "+s.User.RefreshToken)
}
