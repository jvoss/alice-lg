package frrproxy

import (
	"encoding/json"
	"time"

	"github.com/alice-lg/alice-lg/pkg/api"
	"github.com/alice-lg/alice-lg/pkg/decoders"
	"github.com/alice-lg/alice-lg/pkg/pools"
)

func parseStatusResponse(body io.Reader) (*api.StatusResponse, error) {
	status := &api.StatusResponse{}
	dec := json.NewDecoder(body)
	if err := dec.Decode(status); err != nil {
		return nil, err
	}
	return status, nil
}

func parseNeighborsResponse(body io.Reader) (*api.NeighborsResponse, error) {
	// FRR's `show bgp summary json` is a map of peers
	var summary struct {
		Peers map[string]interface{} `json:"peers"`
	}
	if err := json.NewDecoder(body).Decode(&summary); err != nil {
		return nil, err
	}

	neighbors := api.Neighbors{}
	for peerID, peerData := range summary.Peers {
		data := peerData.(map[string]interface{})
		neighbor := &api.Neighbor{
			ID:          peerID,
			Address:     peerID,
			ASN:         decoders.Int(data["as"], 0),
			State:       decoders.String(data["state"], "unknown"),
			Description: decoders.String(data["description"], ""),
			Uptime:      decoders.Int(data["peerUptimeMsec"], 0) / 1000,
			Routes: &api.NeighborRoutes{
				Imported: decoders.Int(data["pfxRcd"], 0),
			},
		}
		neighbors = append(neighbors, neighbor)
	}

	return &api.NeighborsResponse{
		Response:  api.Response{Meta: &api.Meta{}},
		Neighbors: neighbors,
	}, nil
}

func parseNeighborsStatusResponse(body io.Reader) (*api.NeighborsStatusResponse, error) {
	res, err := parseNeighborsResponse(body)
	if err != nil {
		return nil, err
	}
	return &api.NeighborsStatusResponse{
		Response:  res.Response,
		Neighbors: res.Neighbors,
	}, nil
}

func parseRouteData(
	rdata map[string]interface{},
	network string, // additional input from the parent key in "routes"
	config Config,
	keepDetails bool,
) *api.Route {
	gwpool := pools.Gateways4 // Default to IPv4, may vary depending on your logic

	// Build BGP Info
	asPath := decoders.String(rdata["aspath"].(map[string]interface{})["string"], "")
	origin := decoders.String(rdata["origin"], "IGP")
	metric := decoders.Int(rdata["metric"], 0)
	locPrf := decoders.Int(rdata["locPrf"], 0)

	// Extract peer info
	peer := rdata["peer"].(map[string]interface{})
	peerId := decoders.String(peer["peerId"], "unknown")
	neighborID := pools.Neighbors.Acquire(peerId)

	// Build BGPInfo
	bgpInfo := &api.BGPInfo{
		Origin: pools.Origins.Acquire(origin),
		// AsPath:    pools.ASPaths.Acquire([]int{}), // No AS numbers in "Local"
		AsPath:    asPath,
		NextHop:   gwpool.Acquire("unknown"), // Set below from nexthop
		LocalPref: locPrf,
		Med:       metric,
	}

	// Nexthop parsing
	nexthops, ok := rdata["nexthops"].([]interface{})
	gateway := "unknown"
	if ok && len(nexthops) > 0 {
		first := nexthops[0].(map[string]interface{})
		gateway = decoders.String(first["ip"], "unknown")
		bgpInfo.NextHop = gwpool.Acquire(gateway)
	}

	// TTL / Age: Derive from `lastUpdate.epoch`
	age := time.Duration(0)
	if lastUpdate, ok := rdata["lastUpdate"].(map[string]interface{}); ok {
		if epochF, ok := lastUpdate["epoch"].(float64); ok {
			epoch := int64(epochF)
			age = time.Since(time.Unix(epoch, 0).UTC())
		}
	}

	// Precompute details if needed
	var details json.RawMessage
	if keepDetails {
		detailsJSON, err := json.Marshal(rdata)
		if err == nil {
			details = json.RawMessage(detailsJSON)
		}
	}

	route := &api.Route{
		NeighborID: neighborID,
		Network:    network,
		Interface:  pools.Interfaces.Acquire("unknown"),
		Metric:     metric,
		Primary:    decoders.Bool(rdata["valid"], false),
		LearntFrom: gwpool.Acquire(gateway),
		Gateway:    gwpool.Acquire(gateway),
		Age:        age,
		Type:       pools.Types.Acquire([]string{"BGP"}),
		BGP:        bgpInfo,
		Details:    &details,
	}
	return route
}
