package api

import (
	"time"
)

type UserInfo struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullname"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login"`
}

func CurrentUser(client *Client, hostname string) (*UserInfo, error) {
	var user UserInfo
	err := client.REST(hostname, "GET", "userinfo", nil, &user)
	return &user, err
}

func CurrentUsername(client *Client, hostname string) (string, error) {

	user, err := CurrentUser(client, hostname)
	if err != nil {
		return "", err
	}

	return user.Username, nil
}

func CurrentUserID(client *Client, hostname string) (string, error) {
	user, err := CurrentUser(client, hostname)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}
