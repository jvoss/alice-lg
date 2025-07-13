package frrproxy

// Parsing for `show bgp vrf <table> [ipv4|ipv6] json`

type BgpRouteData struct {
	Routes map[string][]BgpRoute `json:"routes"`
}

type BgpRoute struct {
	Bestpath bool `json:"bestpath"`
	Metric   int  `json:"metric"`

	// interface
	// gateway
	// metric
	// bgpinfo
	// age
	// type [BGP, unicast, univ]
	// primary [bool]
	// learntFrom string

	// Details // original json raw message
}

type BgpRouteNextHop struct {
	Ip string `json:"ip"`
}
