// Package cache provides Gavel's local analysis cache abstractions.
//
// Database-backed violation and linter-stat caches use the PostgreSQL
// connection prepared by internal/database. Their schema is declared in the
// embedded Atlas HCL bundle there. The separate Cache type remains a small
// filesystem JSON cache for last-run timestamps and does not own a database.
package cache
