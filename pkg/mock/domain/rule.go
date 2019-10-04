package domain

import (
	"github.com/aldas/xroad-mock-proxy/pkg/mock/config"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

var appFs = afero.NewOsFs()

// Rules is collection type for Rule structures
type Rules []Rule

// Rule describes rules how to mock requests
type Rule struct {
	ID             int64
	Service        string
	Priority       int64
	MatcherRegex   []*regexp.Regexp
	IdentityRegex  *regexp.Regexp
	Template       template.Template
	TemplateBytes  []byte
	Timeout        time.Duration
	ResponseStatus int
	IsReadOnly     bool
}

// ConvertRules converts config to rules domain object
func ConvertRules(rules config.RuleConfigs) (Rules, error) {
	result := make(Rules, len(rules))
	for i, r := range rules {
		rule, err := convertRule(r)
		if err != nil {
			return nil, err
		}
		rule.ID = int64(i + 1)

		result[i] = rule
	}

	return result, nil
}

func convertRule(r config.RuleConf) (Rule, error) {
	matchers := make([]*regexp.Regexp, 0)
	for _, m := range r.MatcherRegex {
		matcher, err := regexp.Compile(m)
		if err != nil {
			return Rule{}, errors.Wrap(err, "failed to compile matcher regexp")
		}
		matchers = append(matchers, matcher)
	}

	var identityRegex *regexp.Regexp
	if r.IdentityRegex != "" {
		tmpRegex, err := regexp.Compile(r.IdentityRegex)
		if err != nil {
			return Rule{}, errors.Wrap(err, "failed to compile identity regexp")
		}
		identityRegex = tmpRegex
	}

	tmpl, tmplBytes, err := compileTemplate(r.TemplateFile)
	if err != nil {
		return Rule{}, err
	}

	var timeout time.Duration
	if r.Timeout != "" {
		timeout, err = time.ParseDuration(r.Timeout)
		if err != nil {
			return Rule{}, errors.Wrap(err, "failed to parse rule timeout to duration")
		}
	}

	if r.ResponseStatus == 0 {
		r.ResponseStatus = http.StatusOK
	}

	isReadonly := true
	if r.IsReadOnly != nil {
		isReadonly = *r.IsReadOnly
	}

	return Rule{
		Service:        strings.ToLower(r.Service),
		Priority:       r.Priority,
		MatcherRegex:   matchers,
		IdentityRegex:  identityRegex,
		Template:       *tmpl,
		TemplateBytes:  tmplBytes,
		Timeout:        timeout,
		ResponseStatus: r.ResponseStatus,
		IsReadOnly:     isReadonly,
	}, nil
}

func compileTemplate(templateFile string) (*template.Template, []byte, error) {
	body, err := afero.ReadFile(appFs, templateFile)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read template file")
	}

	tmpl, err := template.New("template").Parse(string(body))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse template")
	}

	return tmpl, body, err
}

// MatchService returns slice of Rules matching given service name
func (r Rules) MatchService(service string) Rules {
	service = strings.ToLower(service)

	result := make(Rules, 0)
	for _, rule := range r {
		if strings.ToLower(rule.Service) == service {
			result = append(result, rule)
		}
	}
	return result
}

// MatchRegex returns first rule matching its regex
func (r Rules) MatchRegex(requestBody []byte) (Rule, bool) {
	sort.Sort(byPriorityDesc(r))
	for _, rule := range r {
		if rule.match(requestBody) {
			return rule, true
		}
	}
	return Rule{}, false
}

func (r Rule) match(requestBody []byte) bool {
	if len(r.MatcherRegex) == 0 {
		return true
	}

	for _, regex := range r.MatcherRegex {
		if regex.Match(requestBody) {
			return true
		}
	}
	return false
}

// MatchIdentity matches identity (if there is) from request body
func (r Rule) MatchIdentity(requestBody []byte) (string, bool) {
	if r.IdentityRegex == nil {
		return "", true
	}

	match := r.IdentityRegex.FindSubmatch(requestBody)
	if match == nil || len(match) == 1 {
		return "", false
	}

	return string(match[1]), true
}

type byPriorityDesc Rules

func (s byPriorityDesc) Len() int      { return len(s) }
func (s byPriorityDesc) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byPriorityDesc) Less(i, j int) bool {
	return s[i].Priority > s[j].Priority
}
