package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var dummyUser map[string]interface{} = map[string]interface{}{
	"email":      "me@email.com",
	"fullname":   "I'm a developer",
	"username":   "me",
	"id":         "1234",
	"last_login": "2021-03-03T09:30:00.105730+00:00",
}

func TestCurrentUser(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost").
		Get("/api/v1/userinfo").
		Reply(200).
		JSON(dummyUser)

	httpClient := NewHTTPClient()
	gock.InterceptClient(httpClient)

	apiClient := NewClientFromHTTP(httpClient)
	user, err := CurrentUser(apiClient, "localhost")

	assert.Equal(t, nil, err, "Unexpected error")
	assert.NotEqual(t, nil, user, "Invalid user")
}

func TestCurrentUsername(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost").
		Get("/api/v1/userinfo").
		Reply(200).
		JSON(dummyUser)

	httpClient := NewHTTPClient()
	gock.InterceptClient(httpClient)

	apiClient := NewClientFromHTTP(httpClient)
	username, err := CurrentUsername(apiClient, "localhost")

	assert.Equal(t, nil, err, "Unexpected error")
	assert.Equal(t, "me", username, "Username is not as expected")
}

func TestCurrentUserID(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost").
		Get("/api/v1/userinfo").
		Reply(200).
		JSON(dummyUser)

	httpClient := NewHTTPClient()
	gock.InterceptClient(httpClient)

	apiClient := NewClientFromHTTP(httpClient)
	userID, err := CurrentUserID(apiClient, "localhost")

	assert.Equal(t, nil, err, "Unexpected error")
	assert.Equal(t, "1234", userID, "User ID is not as expected")
}
