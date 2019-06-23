package dto

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/dto"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// RuleDTO is DTO for rule
type RuleDTO struct {
	ID                   int64            `json:"id"`
	Server               string           `json:"server"`
	Service              string           `json:"service"`
	Priority             int64            `json:"priority"`
	MatcherRemoteAddr    []string         `json:"matcher_remote_addr"`
	MatcherRegex         []string         `json:"matcher_regexes"`
	RequestReplacements  []ReplacementDTO `json:"request_replacements"`
	ResponseReplacements []ReplacementDTO `json:"response_replacements"`
	IsReadOnly           bool             `json:"read_only"`
}

// ReplacementDTO is DTO for replacements
type ReplacementDTO struct {
	Regex string `json:"regex"`
	Value string `json:"value"`
}

// RulesToDTO converts slice of rules to DTOs
func RulesToDTO(rules []domain.Rule) []RuleDTO {
	result := make([]RuleDTO, len(rules))
	for i := 0; i < len(rules); i++ {
		result[i] = RuleToDTO(rules[i])
	}
	return result
}

// RuleToDTO converts rule to DTO
func RuleToDTO(r domain.Rule) RuleDTO {
	return RuleDTO{
		ID:                   r.ID,
		Server:               r.Server,
		Service:              r.Service,
		Priority:             r.Priority,
		MatcherRemoteAddr:    r.MatcherRemoteAddr,
		MatcherRegex:         dto.RegExpToSlice(r.MatcherRegex),
		RequestReplacements:  replacementsToDTO(r.RequestReplacements),
		ResponseReplacements: replacementsToDTO(r.ResponseReplacements),
		IsReadOnly:           r.IsReadOnly,
	}
}

// ToRule converts DTO object to rule domain object
func ToRule(r RuleDTO) (domain.Rule, error) {
	if r.Server == "" {
		return domain.Rule{}, errors.New("server can not be empty")
	}

	if r.Service == "" && len(r.MatcherRemoteAddr) == 0 && len(r.MatcherRegex) == 0 {
		return domain.Rule{}, errors.New("at least one matcher must be set")
	}

	matchers, err := dto.SliceToRegexp(r.MatcherRegex)
	if err != nil {
		return domain.Rule{}, err
	}

	requestReplacements, err := toReplacements(r.RequestReplacements)
	if err != nil {
		return domain.Rule{}, err
	}

	responseReplacements, err := toReplacements(r.ResponseReplacements)
	if err != nil {
		return domain.Rule{}, err
	}

	return domain.Rule{
		ID:                   r.ID,
		Server:               strings.ToLower(r.Server),
		Service:              r.Service,
		Priority:             r.Priority,
		MatcherRemoteAddr:    r.MatcherRemoteAddr,
		MatcherRegex:         matchers,
		RequestReplacements:  requestReplacements,
		ResponseReplacements: responseReplacements,
		IsReadOnly:           r.IsReadOnly,
	}, nil
}

func replacementsToDTO(replacements domain.Replacements) []ReplacementDTO {
	result := make([]ReplacementDTO, len(replacements))
	for i := 0; i < len(replacements); i++ {
		result[i] = replacementToDTO(replacements[i])
	}
	return result
}

func replacementToDTO(r domain.Replacement) ReplacementDTO {
	return ReplacementDTO{
		Regex: r.Regex.String(),
		Value: r.Value,
	}
}

func toReplacements(replacements []ReplacementDTO) ([]domain.Replacement, error) {
	result := make([]domain.Replacement, len(replacements))
	for i := 0; i < len(replacements); i++ {
		r, err := toReplacement(replacements[i])
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func toReplacement(r ReplacementDTO) (domain.Replacement, error) {
	regex, err := regexp.Compile(r.Regex)
	if err != nil {
		return domain.Replacement{}, errors.Wrap(err, "failed to compile replacement regexp")
	}

	return domain.Replacement{
		Regex: regex,
		Value: r.Value,
	}, nil
}
