// Package platform reads every organization's data for the platform dashboard, across
// tenants. Nothing here writes; callers authorize platform admins before using it.
package platform

import "gorm.io/gorm"

// Store reads the organizations of one Clerk instance and what they own.
type Store struct {
	DB *gorm.DB
	// Issuer is the Clerk instance whose organizations are read; the local
	// organization and other instances' are left out.
	Issuer string
}
