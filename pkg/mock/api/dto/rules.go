package dto

import (
	"encoding/base64"
	"github.com/aldas/xroad-mock-proxy/pkg/common/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/pkg/errors"
	"regexp"
	"text/template"
	"time"
)

// RuleDTO is DTO for rule
type RuleDTO struct {
	ID             int64    `json:"id"`
	Service        string   `json:"service"`
	Priority       int64    `json:"priority"`
	MatcherRegex   []string `json:"matcher_regexes"`
	IdentityRegex  string   `json:"identity_regex"`
	Template       string   `json:"template"`
	Timeout        string   `json:"timeout_duration"`
	ResponseStatus int      `json:"response_status"`
	IsReadOnly     bool     `json:"read_only"`
}

// RulesToDTO converts slice of rules to DTOs
func RulesToDTO(rules []domain.Rule) []RuleDTO {
	result := make([]RuleDTO, len(rules))
	for i := 0; i < len(rules); i++ {
		result[i] = RuleToDTO(rules[i])
	}
	return result
}

// RuleToDTO converts rule to DTO without adding template
func RuleToDTO(r domain.Rule) RuleDTO {
	identityRegexpStr := ""
	if r.IdentityRegex != nil {
		identityRegexpStr = r.IdentityRegex.String()
	}
	return RuleDTO{
		ID:             r.ID,
		Service:        r.Service,
		Priority:       r.Priority,
		MatcherRegex:   dto.RegExpToSlice(r.MatcherRegex),
		IdentityRegex:  identityRegexpStr,
		Timeout:        r.Timeout.String(),
		ResponseStatus: r.ResponseStatus,
		IsReadOnly:     r.IsReadOnly,
	}
}

// RuleToFullDTO converts rule to DTO with all of its fields
func RuleToFullDTO(r domain.Rule) RuleDTO {
	identityRegexpStr := ""
	if r.IdentityRegex != nil {
		identityRegexpStr = r.IdentityRegex.String()
	}

	return RuleDTO{
		ID:             r.ID,
		Service:        r.Service,
		Priority:       r.Priority,
		IdentityRegex:  identityRegexpStr,
		MatcherRegex:   dto.RegExpToSlice(r.MatcherRegex),
		Template:       base64.StdEncoding.EncodeToString(r.TemplateBytes),
		Timeout:        r.Timeout.String(),
		ResponseStatus: r.ResponseStatus,
		IsReadOnly:     r.IsReadOnly,
	}
}

// FullDTOToRule converts DTO to Rule with all of its fields
func FullDTOToRule(r RuleDTO) (domain.Rule, error) {
	if r.Priority <= 0 {
		return domain.Rule{}, errors.New("priority must be larger than 0")
	}

	var timeout time.Duration
	if r.Timeout != "" {
		t, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return domain.Rule{}, errors.Wrap(err, "failed to parse duration")
		}
		timeout = t
	}

	responseStatus := 200
	if r.ResponseStatus != 0 {
		responseStatus = r.ResponseStatus
	}

	matcherRegexps, err := dto.SliceToRegexp(r.MatcherRegex)
	if err != nil {
		return domain.Rule{}, err
	}

	var identityRegex *regexp.Regexp
	if r.IdentityRegex != "" {
		irTmp, err := regexp.Compile(r.IdentityRegex)
		if err != nil {
			return domain.Rule{}, errors.Wrap(err, "failed to compile identityRegex regexp")
		}
		identityRegex = irTmp
	}

	tmpl, tmplBytes, err := compileTemplate(r.Template)
	if err != nil {
		return domain.Rule{}, err
	}

	return domain.Rule{
		ID:             r.ID,
		Service:        r.Service,
		Priority:       r.Priority,
		IdentityRegex:  identityRegex,
		MatcherRegex:   matcherRegexps,
		TemplateBytes:  tmplBytes,
		Template:       *tmpl,
		Timeout:        timeout,
		ResponseStatus: responseStatus,
	}, nil
}

func compileTemplate(base64Template string) (*template.Template, []byte, error) {
	templateBytes, err := base64.StdEncoding.DecodeString(base64Template)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to base64 decode template string")
	}

	tmpl, err := template.New("template").Parse(string(templateBytes))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse template")
	}

	return tmpl, templateBytes, err
}
