// Package versionparse is the single source of truth for Gamaj version tags.
//
// Every subsystem that reads or validates a version — the panel update
// checker, the node service and the shell installers — must accept exactly the
// same tag shapes:
//
//   - stable releases: is.0.0.1, v1.2.3, 1.2.3
//   - dev builds:      dev-abcdef0
//
// Keeping the rules here prevents the panel, the node and the installers from
// drifting apart again.
package versionparse

import (
	"fmt"
	"regexp"
	"strings"
)

// ReleasePattern matches a stable Gamaj release tag such as is.0.0.1.
var ReleasePattern = regexp.MustCompile(`^(?:v|is)?\.?\d+(?:\.\d+){1,3}(?:[-+._A-Za-z0-9]*)?$`)

// DevPattern matches a dev build tag such as dev-abcdef0.
var DevPattern = regexp.MustCompile(`^dev-[0-9a-fA-F]{7,40}$`)

// IsRelease reports whether the value is a stable release tag.
func IsRelease(value string) bool {
	return ReleasePattern.MatchString(strings.TrimSpace(value))
}

// IsDev reports whether the value belongs to the dev channel.
func IsDev(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "dev" || DevPattern.MatchString(normalized)
}

// IsUpdateTarget reports whether the value may be handed to the CLI as an
// update target. Channels (latest, dev) are always valid; anything else must
// be a concrete release or dev tag.
func IsUpdateTarget(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "latest", "dev":
		return true
	}
	return DevPattern.MatchString(normalized) || IsRelease(value)
}

// Parts parses a version into its numeric components. The v and is prefixes
// are ignored, and any pre-release or build suffix is dropped. It returns nil
// when the value is not a numeric version.
func Parts(value string) []int {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.TrimPrefix(normalized, "v")
	normalized = strings.TrimPrefix(normalized, "is.")
	normalized = strings.SplitN(normalized, "-", 2)[0]
	normalized = strings.SplitN(normalized, "+", 2)[0]
	segments := strings.Split(normalized, ".")
	if len(segments) < 2 || len(segments) > 4 {
		return nil
	}
	parts := make([]int, len(segments))
	for index, segment := range segments {
		var number int
		if _, err := fmt.Sscanf(segment, "%d", &number); err != nil {
			return nil
		}

		parts[index] = number
	}
	return parts
}

// AtLeast reports whether version is greater than or equal to floor.
func AtLeast(version string, floor string) bool {
	left, right := Parts(version), Parts(floor)
	if left == nil || right == nil {
		return false
	}
	for index := 0; index < len(left) || index < len(right); index++ {
		var leftPart, rightPart int
		if index < len(left) {
			leftPart = left[index]
		}
		if index < len(right) {
			rightPart = right[index]
		}
		if leftPart != rightPart {
			return leftPart > rightPart
		}
	}
	return true
}
