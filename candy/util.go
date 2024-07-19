package candy

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

func IsAlphanumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", s)
	return match
}

func Sha256sum(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}
