package frrproxy

type BgpSummary struct {
	Peers map[string]BgpSummaryPeer `json:"peers"`
}

type BgpSummaryPeer struct {
	RemoteAs       uint32 `json:"remoteAs"`
	State          string `json:"state"`
	PeerUptimeMsec uint64 `json:"peerUptimeMsec"`
}
