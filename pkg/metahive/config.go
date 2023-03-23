package metahive

// Config data for metahive.
type Config struct {
	// Disable this NRI plugin and containerd NRI functionality altogether.
	Disable bool `toml:"disable" json:"disable"`
	// Broadcast defines whether broadcast discovery is enabled
	Broadcast bool `toml:"broadcast" json:"broadcast"`
	// RootAddress defines the root xferer to join
	RootAddress string `toml:"root_address" json:"rootAddress"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Disable:     true,
		Broadcast:   false,
		RootAddress: "0.0.0.0",
	}
}
