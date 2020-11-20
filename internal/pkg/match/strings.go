// Package match contains functions for matching by regular expressions.
// More detailed about regular expressions you can read by link: https://golang.org/pkg/regexp/
package match

import "regexp"

// Function Email checks if s string is email.
func Email(s string) (bool, error) {
	return regexp.MatchString(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`, s)
}
