package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Organization represents an Organization under Reliably
type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	Owner     string `json:"owner,omitempty"`
}

// ListOrganizations list all organizations to which
// the current logged in user is a member of
func ListOrganizations(client *Client, hostname string) ([]Organization, error) {
	var orgs []Organization

	err := client.REST(hostname, "GET", "orgs", nil, &orgs)
	return orgs, err
}

// CurrentUserOrganization returns the default organization of the
// current logged in user
func CurrentUserOrganization(client *Client, hostname string) (*Organization, error) {

	orgs, err := ListOrganizations(client, hostname)
	if err != nil {
		return nil, err
	}

	user, err := CurrentUser(client, hostname)
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.Name == user.Username && org.CreatedBy == user.ID {
			return &org, nil
		}

	}

	return nil, errors.New("No organization found for current username")
}

// CurrentUserOrganizationID returns the identifier of the
// default organization for the current logged in user
func CurrentUserOrganizationID(client *Client, hostname string) (string, error) {

	org, err := CurrentUserOrganization(client, hostname)
	if err != nil {
		return "", err
	}

	return org.ID, nil
}

func CreateOrganisation(client *Client, hostname, orgName string) error {
	type model struct {
		Name string `json:"name"`
	}

	m := model{Name: orgName}
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&m); err != nil {
		return err
	}

	return client.REST(hostname, "POST", "orgs", &buffer, nil)
}

func DeleteOrganisation(client *Client, hostname, orgID string) error {
	path := fmt.Sprint("orgs/", orgID)
	return client.REST(hostname, "DELETE", path, nil, nil)
}

func AddUserToOrganisation(client *Client, hostname, orgID, username string) error {
	path := fmt.Sprintf("orgs/%s/users/%s", orgID, username)
	return client.REST(hostname, "PUT", path, nil, nil)
}

func RemoveUserFromOrganisation(client *Client, hostname, orgID, username string) error {
	path := fmt.Sprintf("orgs/%s/users/%s", orgID, username)
	return client.REST(hostname, "DELETE", path, nil, nil)
}
