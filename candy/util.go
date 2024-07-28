package candy

import (
	"regexp"
)

func IsAlphaNumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", s)
	return match
}

func IsValidUsername(username string) bool {
	if username == "@" {
		return true
	}
	if len(username) < 3 {
		return false
	}
	if len(username) > 32 {
		return false
	}
	if !IsAlphaNumeric(username) {
		return false
	}
	return true
}
