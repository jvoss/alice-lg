package frrproxy

// Parsing for `show bgp vrf <table> [ipv4|ipv6] neighbors` json

type BgpNeighbor struct {
	RemoteAs          uint32            `json:"remoteAs"`
	BgpState          string            `json:"bgpState"`
	BgpTimerUpMsec    uint64            `json:"bgpTimerUpMsec"`
	NbrDesc           string            `json:"nbrDesc"`
	PrefixStats       PrefixStats       `json:"prefixStats"`
	AddressFamilyInfo AddressFamilyInfo `json:"addressFamilyInfo"`
	LastResetDueTo    string            `json:"lastResetDueTo"`
}

type PrefixStats struct {
	InboundFiltered int `json:"inboundFiltered"`
}

type AddressFamilyInfo struct {
	Ipv4Unicast AddressFamilyCounters `json:"ipv4Unicast"`
	Ipv6Unicast AddressFamilyCounters `json:"ipv6Unicast"`
}

type AddressFamilyCounters struct {
	AcceptedPrefixCounter int `json:"acceptedPrefixCounter"`
	SentPrefixCounter     int `json:"sentPrefixCounter"`
}

func (n BgpNeighbor) State() string {
	var stateLabel string

	switch n.BgpState {
	case "Established":
		stateLabel = "up"
	default:
		stateLabel = "down"
	}

	return stateLabel
}

func (n BgpNeighbor) RoutesAccepted() int {
	return n.AddressFamilyInfo.Ipv4Unicast.AcceptedPrefixCounter +
		n.AddressFamilyInfo.Ipv6Unicast.AcceptedPrefixCounter
}

func (n BgpNeighbor) RoutesExported() int {
	return n.AddressFamilyInfo.Ipv4Unicast.SentPrefixCounter +
		n.AddressFamilyInfo.Ipv6Unicast.SentPrefixCounter
}

func (n BgpNeighbor) RoutesFiltered() int {
	return n.PrefixStats.InboundFiltered
}
