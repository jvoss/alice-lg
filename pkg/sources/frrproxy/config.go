package frrproxy

import "time"

// Config contains all configuration attributes
// for a frr-proxy based source.
type Config struct {
	ID   string
	Name string

	API       string        `ini:"api"`
	APIKey    string        `ini:"apiKey"`
	Timeout   time.Duration `ini:"timeout"`
	TLS       bool          `ini:"tls"`
	TLSVerify bool          `ini:"tlsVerify"`
	// Timezone        string `ini:"timezone"`
	// ServerTime      string `ini:"servertime"`
	// ServerTimeShort string `ini:"servertime_short"`
	// ServerTimeExt   string `ini:"servertime_ext"`
	// ShowLastReboot  bool   `ini:"show_last_reboot"`

	Type      string `ini:"type"`
	MainTable string `ini:"main_table"`
	Afi       string `ini:"afi"`
	// PeerTablePrefix         string `ini:"peer_table_prefix"`
	// PipeProtocolPrefix      string `ini:"pipe_protocol_prefix"`
	// AltPipeProtocolPrefix   string `ini:"alt_pipe_protocol_prefix"`
	// AltPipeProtocolSuffix   string `ini:"alt_pipe_protocol_suffix"`
	// NeighborsRefreshTimeout int `ini:"neighbors_refresh_timeout"`

	StreamParserThrottle int
}
