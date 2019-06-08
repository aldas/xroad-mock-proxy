package domain

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/pkg/errors"
	"regexp"
	"sort"
	"strings"
)

// Rules is collection type for Rule structures
type Rules []Rule

// Rule describes rules where and when to proxy requests
type Rule struct {
	ID                   int64
	Server               string
	Service              string
	Priority             int64
	MatcherRemoteAddr    []string
	MatcherRegex         []*regexp.Regexp
	RequestReplacements  Replacements
	ResponseReplacements Replacements
	IsReadOnly           bool
}

// Replacements is collection type for Replacement structures
type Replacements []Replacement

// Replacement describes rules for replacing things in request/response
type Replacement struct {
	Regex *regexp.Regexp
	Value string
}

// ConvertRules convert configuration to domain object
func ConvertRules(conf config.RuleConfigs) (Rules, error) {
	result := Rules{}
	for i, c := range conf {
		r, err := convertRule(c)
		if err != nil {
			return Rules{}, err
		}
		r.ID = int64(i + 1)
		result = append(result, r)
	}
	return result, nil
}

func convertRule(conf config.RuleConf) (Rule, error) {
	matchers := make([]*regexp.Regexp, 0)
	for _, m := range conf.MatcherRegex {
		matcher, err := regexp.Compile(m)
		if err != nil {
			return Rule{}, errors.Wrap(err, "failed to compile matcher regexp")
		}
		matchers = append(matchers, matcher)
	}

	requestReplacements, err := convertReplacements(conf.RequestReplacements)
	if err != nil {
		return Rule{}, err
	}
	responseReplacements, err := convertReplacements(conf.ResponseReplacements)
	if err != nil {
		return Rule{}, err
	}

	isReadOnly := true
	if conf.IsReadOnly != nil {
		isReadOnly = *conf.IsReadOnly
	}

	return Rule{
		Server:               conf.Server,
		Service:              conf.Service,
		Priority:             conf.Priority,
		MatcherRemoteAddr:    conf.MatcherRemoteAddr,
		MatcherRegex:         matchers,
		RequestReplacements:  requestReplacements,
		ResponseReplacements: responseReplacements,
		IsReadOnly:           isReadOnly,
	}, nil
}

func convertReplacements(conf config.ReplacementConfigs) (Replacements, error) {
	result := Replacements{}
	for _, c := range conf {
		repl, err := convertReplacement(c)
		if err != nil {
			return Replacements{}, err
		}
		result = append(result, repl)
	}
	return result, nil
}

func convertReplacement(conf config.ReplacementConf) (Replacement, error) {
	r, err := regexp.Compile(conf.Regex)
	if err != nil {
		return Replacement{}, errors.Wrap(err, "failed to compile replacement regexp")
	}

	return Replacement{
		Regex: r,
		Value: conf.Value,
	}, nil
}

// MatchRemoteAddr returns slice of Rules matching given remote addr
func (r Rules) MatchRemoteAddr(remoteAddr string) Rules {
	result := make(Rules, 0)
	for _, rule := range r {
		if rule.matchRemoteAddr(remoteAddr) {
			result = append(result, rule)
		}
	}
	return result
}

// MatchService returns slice of Rules matching given service name
func (r Rules) MatchService(service string) Rules {
	result := make(Rules, 0)
	for _, rule := range r {
		if rule.Service == service {
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

// FindByID finds rule by ID
func (r Rules) FindByID(ID int64) (Rule, bool) {
	for _, rule := range r {
		if rule.ID == ID {
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

func (r Rule) matchRemoteAddr(remoteAddr string) bool {
	if len(r.MatcherRemoteAddr) == 0 {
		return true
	}
	for _, remote := range r.MatcherRemoteAddr {
		if strings.HasPrefix(remoteAddr, remote) {
			return true
		}
	}
	return false
}

// ApplyRequestReplacements applies rule request replacements on body
func (r Rule) ApplyRequestReplacements(body []byte) []byte {
	return applyReplacements(body, r.RequestReplacements)
}

// ApplyResponseReplacements applies rule response replacements on body
func (r Rule) ApplyResponseReplacements(body []byte) []byte {
	return applyReplacements(body, r.ResponseReplacements)
}

func applyReplacements(body []byte, replacements Replacements) []byte {
	result := make([]byte, len(body))
	copy(result, body)

	for _, r := range replacements {
		result = r.Regex.ReplaceAll(result, []byte(r.Value))
	}

	return result
}

type byPriorityDesc Rules

func (s byPriorityDesc) Len() int      { return len(s) }
func (s byPriorityDesc) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byPriorityDesc) Less(i, j int) bool {
	iPriority := s[i].Priority
	jPriority := s[j].Priority
	if iPriority == jPriority {
		return s[i].ID > s[j].ID
	}
	return iPriority > jPriority
}
