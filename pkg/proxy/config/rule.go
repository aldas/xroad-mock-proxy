package config

// RuleConfigs is collection type for RuleConf structure
type RuleConfigs []RuleConf

// RuleConf describes configuration for rules where and when to proxy requests
type RuleConf struct {
	// server name where to direct matched request. Must have matching value in ProxyServerConfigs.Name
	Server string `mapstructure:"server"`
	// full service name to match (subsystemCode.serviceCode.serviceVersion) (for example: rr.RR67_muutus.v1)
	Service string `mapstructure:"service"`
	// priority defines in which order rules are matched. higher the better/sooner is rule matched
	Priority int64 `mapstructure:"priority"`
	// additional list of request remote addresses to decide if request is proxied to this service proxy
	MatcherRemoteAddr []string `mapstructure:"request_matcher_remote_addr"`
	// additional regex'es run on request to decide if request is proxied to this service proxy
	MatcherRegex []string `mapstructure:"request_matcher_regexes"`
	// regex'es to replace contents of request before it is proxied
	RequestReplacements ReplacementConfigs `mapstructure:"request_replacements"`
	// regex'es to replace contents of proxied response
	ResponseReplacements ReplacementConfigs `mapstructure:"response_replacements"`
	// should rule be changeable in API (defaults to true)
	IsReadOnly *bool `mapstructure:"read_only"`
}

// ReplacementConfigs is collection type for ReplacementConf structures
type ReplacementConfigs []ReplacementConf

// ReplacementConf describes rules for replacing things in request/response
type ReplacementConf struct {
	Regex string `mapstructure:"regex"`
	Value string `mapstructure:"value"`
}
