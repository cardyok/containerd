package imagegcplugin

// Config is to config image gc.
type Config struct {
	// Allow to higher than lowThresholdPercent.
	LowThresholdPercent int `toml:"low_threshold_percent" json:"low_threshold_percent"`
	// Any usage below highThresholdPercent will never triger garbage collect.
	HighThresholdPercent int `toml:"high_threshold_percent" json:"high_threshold_percent"`
	// MinAgeSeconds is minimum age at which an image can be garbage collected.
	MinAgeSeconds uint `toml:"min_age_seconds" json:"min_age_seconds"`
	// Whitelist is to keep images which can't be removed.
	Whitelist []string `toml:"whitelist" json:"whitelist"`
	// WhitelistGoRegex is to keep images matched by whitelist_go_regex.
	// NOTE: Only works for tag, not sha256 digest!
	WhitelistGoRegex string `toml:"whitelist_go_regex" json:"whitelist_go_regex"`
	// GCPeriodSeconds is to say how often garbage collect works.
	GCPeriodSeconds uint `toml:"gc_period_seconds" json:"gc_period_seconds"`
}

func defaultConfig() Config {
	return Config{
		LowThresholdPercent:  75,
		HighThresholdPercent: 100,    // default closed
		MinAgeSeconds:        3 * 60, // 3 mins
		GCPeriodSeconds:      3 * 60, // 3 mins
	}
}
