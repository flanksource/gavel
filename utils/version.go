package utils

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	gitSHARegex = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
)

type ImageReference struct {
	Registry string
	Name     string
	Tag      string
	Digest   string
}

func IsSemver(value string) bool {
	if value == "" {
		return false
	}
	_, err := semver.NewVersion(value)
	return err == nil
}

func IsSHA256(value string) bool {
	return len(value) == 64 && sha256Regex.MatchString(value)
}

func IsGitSHA(value string) bool {
	return gitSHARegex.MatchString(value)
}

func ParseImageReference(image string) ImageReference {
	ref := ImageReference{}

	parts := strings.Split(image, "@")
	imageWithTag := parts[0]

	if len(parts) > 1 {
		ref.Digest = parts[1]
	}

	tagParts := strings.Split(imageWithTag, ":")

	if len(tagParts) > 1 {
		lastPart := tagParts[len(tagParts)-1]
		if !strings.Contains(lastPart, "/") {
			ref.Tag = lastPart
			ref.Name = strings.Join(tagParts[:len(tagParts)-1], ":")
		} else {
			ref.Name = imageWithTag
		}
	} else {
		ref.Name = imageWithTag
	}

	if strings.Contains(ref.Name, "/") {
		nameParts := strings.SplitN(ref.Name, "/", 2)
		if strings.Contains(nameParts[0], ".") || strings.Contains(nameParts[0], ":") {
			ref.Registry = nameParts[0]
			ref.Name = nameParts[1]
		}
	}

	return ref
}

func MatchesPattern(fieldPath, pattern string) bool {
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")

		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			prefix = strings.TrimSuffix(prefix, ".")
			suffix = strings.TrimPrefix(suffix, ".")

			hasPrefix := prefix == "" || strings.HasPrefix(fieldPath, prefix+".")
			hasSuffix := suffix == "" || strings.HasSuffix(fieldPath, "."+suffix)

			if suffix != "" && strings.Contains(suffix, "*") {
				if prefix == "" {
					pathParts := strings.Split(fieldPath, ".")
					suffixParts := strings.Split(suffix, ".")

					for i := 0; i <= len(pathParts)-len(suffixParts); i++ {
						candidate := strings.Join(pathParts[i:i+len(suffixParts)], ".")
						candidateSlash := strings.ReplaceAll(candidate, ".", "/")
						suffixSlash := strings.ReplaceAll(suffix, ".", "/")

						matched, err := filepath.Match(suffixSlash, candidateSlash)
						if err == nil && matched {
							return true
						}
					}
					return false
				}

				suffixPart := fieldPath
				if prefix != "" {
					suffixPart = strings.TrimPrefix(fieldPath, prefix+".")
				}

				pathSlash := strings.ReplaceAll(suffixPart, ".", "/")
				suffixSlash := strings.ReplaceAll(suffix, ".", "/")

				matched, err := filepath.Match(suffixSlash, pathSlash)
				if err != nil {
					return false
				}
				return matched
			}

			return hasPrefix && hasSuffix
		}
	}

	pathSlash := strings.ReplaceAll(fieldPath, ".", "/")
	patternSlash := strings.ReplaceAll(pattern, ".", "/")

	matched, err := filepath.Match(patternSlash, pathSlash)
	if err != nil {
		return false
	}

	return matched
}
