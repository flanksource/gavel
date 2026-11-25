package githubselector

import "regexp"

const QueryPattern = `(^|\s)-?[A-Za-z][A-Za-z0-9_-]*:\S`

var query = regexp.MustCompile(QueryPattern)

func IsQuery(value string) bool { return query.MatchString(value) }
