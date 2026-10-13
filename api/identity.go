package api

import (
	"fmt"

	clickyapi "github.com/flanksource/clicky/api"
)

// Pretty keeps the human-readable identity fields in the same order as JSON.
// An empty organization is explicit rather than silently dropping the field.
func (i Identity) Pretty() clickyapi.Text {
	return clickyapi.Text{Content: fmt.Sprintf(`
ID: %s
Issuer: %s
Name: %s
Email: %s
Organization ID: %s
Organization Slug: %s`, i.ID, i.Issuer, i.Name, i.Email, i.OrganizationID, i.OrganizationSlug)}
}
