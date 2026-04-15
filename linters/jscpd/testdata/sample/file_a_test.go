package sample

import "strings"

func ProcessItemsFixture(items []string) []string {
	var result []string
	for _, item := range items {
		if item == "" {
			continue
		}
		cleaned := strings.TrimSpace(item)
		if len(cleaned) > 0 {
			result = append(result, cleaned)
		}
	}
	return result
}
