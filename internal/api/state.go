package api

import "github.com/google/uuid"

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
	CreateUser   string
}
