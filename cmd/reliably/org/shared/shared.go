package shared

import (
	"strings"

	"github.com/reliablyhq/cli/api"
)

// FilterOrgs returns the list of organizations that succeeded the test function
func FilterOrgs(orgs []api.Organization, test func(o api.Organization) bool) (ret []api.Organization) {
	for _, o := range orgs {
		if test(o) {
			ret = append(ret, o)
		}
	}
	return
}

// IsOwner returns true if the given user ID is the owner of the organization
// If the organization has no specified owner, the creator will be considered
// as the owner.
func IsOwner(userID string, org api.Organization) bool {
	return userID == org.Owner || (org.Owner == "" && userID == org.CreatedBy)
}

// FilterOrgByName returns the reference to the organization for which
// the name is matching the name given as argument
func FilterOrgByName(orgs *[]api.Organization, name string) *api.Organization {
	var org *api.Organization
	for _, o := range *orgs {
		if strings.EqualFold(name, o.Name) {
			org = &o
			break
		}
	}
	return org
}
