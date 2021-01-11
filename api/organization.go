package api

import (
	"errors"
)

// Organization represents an Organization under Reliably
type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
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
