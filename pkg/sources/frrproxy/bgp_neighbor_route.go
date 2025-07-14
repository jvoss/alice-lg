package frrproxy

import (
	"strconv"
	"strings"
)

type BgpNeighborRoutes struct {
	Routes map[string][]BgpNeighborRoute `json:"routes"`
}

type BgpNeighborRoute struct {
	Origin   string    `json:"origin"`
	Path     string    `json:"path"`
	Nexthops []Nexthop `json:"nexthops"`
}

func (a BgpNeighborRoute) AsPath() []int {
	var list []int

	parts := strings.Split(a.Path, " ")

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
