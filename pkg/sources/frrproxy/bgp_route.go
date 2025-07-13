package frrproxy

import (
	"strconv"
	"strings"
)

// Parsing for `show bgp vrf <table> [ipv4|ipv6] json`

type BgpRouteData struct {
	Routes map[string][]BgpRoute `json:"routes"`
}

type BgpRoute struct {
	ImportedFrom string     `json:"importedFrom"`
	AsPath       AsPath     `json:"asPath"`
	Origin       string     `json:"origin"`
	Bestpath     BestPath   `json:"bestpath"`
	Metric       int        `json:"metric"`
	LocPrf       int        `json:"locPrf"`
	LastUpdate   LastUpdate `json:"lastUpdate"`
	Nexthops     []Nexthop  `json:"nexthops"`
	Peer         Peer       `json:"peer"`

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

type AsPath struct {
	String string `json:"string"`
}

func (a AsPath) List() []int {
	var list []int

	parts := strings.Split(a.String, " ")

	for _, p := range parts {
		n, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			// handle error (skip, log, etc.)
			continue
		}
		list = append(list, int(n))
	}

	return list
}

type BestPath struct {
	Overall bool `json:"overall"`
}

type LastUpdate struct {
	Epoch int `json:"epoch"`
}

type Nexthop struct {
	IP string `json:"ip"`
}

type Peer struct {
	PeerID string `json:"peerId"`
}
