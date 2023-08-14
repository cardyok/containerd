package log

// LoggerConfig defines the logger struct to be used for lumberjack logger
type LoggerConfig struct {
	// LogPath of containerd main process
	LogPath string `toml:"log_path"`
	// LogReplica defines how many compressed replica of log should be keeped
	LogReplica int `toml:"log_replica"`
	// LogSize defines max size of log before rotated
	LogSize int `toml:"log_size"`
	// NoCompress defines if rotated file should not be compressed
	NoCompress bool `toml:"noCompress"`
}
