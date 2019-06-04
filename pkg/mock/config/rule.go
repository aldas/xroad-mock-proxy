package config

// RuleConfigs is collection type for RuleConf structure
type RuleConfigs []RuleConf

// RuleConf describes configuration for rules how to mock requests
type RuleConf struct {
	Service        string   `mapstructure:"service"`
	Priority       int64    `mapstructure:"priority"`
	MatcherRegex   []string `mapstructure:"matcher_regexes"`
	IdentityRegex  string   `mapstructure:"identity_regex"`
	TemplateFile   string   `mapstructure:"template_file"`
	Timeout        string   `mapstructure:"timeout_duration"`
	ResponseStatus int      `mapstructure:"response_status"`
	IsReadOnly     *bool    `mapstructure:"read_only"`
}
