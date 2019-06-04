package dto

import (
	"github.com/pkg/errors"
	"regexp"
)

// APIResponse is common response structure for all responses
type APIResponse struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

// RegExpToSlice converts slice of regexps to slice of strings
func RegExpToSlice(regexps []*regexp.Regexp) []string {
	result := make([]string, len(regexps))
	for i, reg := range regexps {
		result[i] = reg.String()
	}
	return result
}

// SliceToRegexp converts slice of strings to slice of regexps
func SliceToRegexp(regexps []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, len(regexps))
	for i, reg := range regexps {
		r, err := regexp.Compile(reg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile regexp")
		}
		result[i] = r
	}
	return result, nil
}
