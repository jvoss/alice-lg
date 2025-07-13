package frrproxy

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/alice-lg/alice-lg/pkg/api"
	"github.com/alice-lg/alice-lg/pkg/pools"
)

// SingleTableFrrProxy is an Alice source
type SingleTableFrrProxy struct {
	GenericFrrProxy
}

// Neighbors retrieves a list of neighbors
func (src *SingleTableFrrProxy) Neighbors(
	ctx context.Context,
) (*api.NeighborsResponse, error) {
	mainTable := src.GenericFrrProxy.config.MainTable

	response := api.NeighborsResponse{}
	response.Neighbors = make(api.Neighbors, 0)

	var neighborsResponse = make(map[string]api.Neighbor, 0)

	// Fetch neighbors from the configured "main_table" for the configured AFI
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+src.client.afi+" neighbors")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var neighbors map[string]BgpNeighbor

	err = json.Unmarshal(bodyBytes, &neighbors)
	if err != nil {
		return nil, err
	}

	for ip, info := range neighbors {
		neighbor, exists := neighborsResponse[ip]

		if !exists {
			neighbor = api.Neighbor{
				ID:             PeerHash(ip),
				Address:        ip,
				ASN:            int(info.RemoteAs),
				State:          info.State(),
				Description:    info.NbrDesc,
				RoutesReceived: info.RoutesAccepted(), // TODO - there is no received total without querying each neighbor
				RoutesFiltered: info.RoutesFiltered(),
				RoutesExported: info.RoutesExported(),
				RoutesAccepted: info.RoutesAccepted(),
				Uptime:         time.Duration(info.BgpTimerUpMsec) * time.Millisecond,
				LastError:      info.LastResetDueTo,
				RouteServerID:  src.config.ID,
				// Details:     <original json>, // TODO
			}
		}

		neighborsResponse[ip] = neighbor
	}

	for _, n := range neighborsResponse {
		neighbor := n
		response.Neighbors = append(response.Neighbors, &neighbor)
	}

	return &response, nil
}

// NeighborsStatus retrievs all status information
// for all peers on the RS.
func (src *SingleTableFrrProxy) NeighborsStatus(
	ctx context.Context,
) (*api.NeighborsStatusResponse, error) {
	log.Printf("I'm here.................TWO..............")
	// panic("unimplemented")

	log.Printf("NeighborsStatus")

	mainTable := src.GenericFrrProxy.config.MainTable

	response := api.NeighborsStatusResponse{}
	response.Neighbors = make(api.NeighborsStatus, 0)

	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+src.client.afi+" summary")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var bgpSummary BgpSummary

	err = json.Unmarshal(bodyBytes, &bgpSummary)
	if err != nil {
		return nil, err
	}

	for ip, _ := range bgpSummary.Peers {
		ns := api.NeighborStatus{}
		ns.ID = PeerHash(ip)
		ns.State = "up" // TODO
		ns.Since = 5 * time.Second

		response.Neighbors = append(response.Neighbors, &ns)
	}

	return &response, nil
}

// NeighborsSummary implements FrrProxy.
func (src *SingleTableFrrProxy) NeighborsSummary(context.Context) (*api.NeighborsResponse, error) {
	panic("unimplemented")
}

// Routes implements FrrProxy.
func (src *SingleTableFrrProxy) Routes(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// RoutesFiltered implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesFiltered(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// RoutesNotExported implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesNotExported(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// RoutesReceived implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesReceived(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// Status implements FrrProxy.
func (src *SingleTableFrrProxy) Status(context.Context) (*api.StatusResponse, error) {
	// panic("unimplemented")
	response := api.StatusResponse{}
	response.Meta = &api.Meta{}
	// response.Status.ServerTime = time.Now()   // TODO
	// response.Status.LastReboot = time.Now()   // TODO
	// response.Status.LastReconfig = time.Now() // TODO
	response.Status.Message = "status-message"
	response.Status.RouterID = "1.2.3.4"
	response.Status.Version = "version-string-here"
	response.Status.Backend = "frr-proxy"

	// response := api.StatusResponse{
	// 	Response: api.Response{
	// 		Meta: &api.Meta{},
	// 	},
	// 	Status: api.Status{},
	// }

	return &response, nil
}

// AllRoutes retrieves a route dump (accepted; not including filtered)
// which is used to learn all prefixes to build
// up a local store for searching.
func (src *SingleTableFrrProxy) AllRoutes(
	ctx context.Context,
) (*api.RoutesResponse, error) {
	mainTable := src.GenericFrrProxy.config.MainTable

	importedRoutes := api.Routes{}

	// Fetch routes from the configured "main_table" for the configured AFI
	// Imported
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+src.client.afi+" detail-routes")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data BgpRouteData

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}

	for prefix, routeData := range data.Routes {
		// log.Printf("Prefix: %s", prefix)

		for _, data := range routeData {
			if data.Peer.PeerID == "::" {
				// Do not process imported routes
				continue
			}

			route := api.Route{}
			route.Network = prefix
			route.Interface = pools.Interfaces.Acquire("unknown")
			route.BGP = &api.BGPInfo{}
			route.Type = pools.Types.Acquire([]string{"BGP"})

			route.NeighborID = pools.Neighbors.Acquire(
				PeerHash(data.Peer.PeerID))

			// Age
			epoch := data.LastUpdate.Epoch
			then := time.Unix(int64(epoch), 0)
			route.Age = time.Since(then)

			if data.Bestpath.Overall {
				route.Primary = true
			} else {
				route.Primary = false
			}

			origin := data.Origin

			route.BGP.Origin = &origin
			route.BGP.AsPath = data.AsPath.List()

			for _, nexthop := range data.Nexthops {
				if nexthop.IP != "::" {
					nh := nexthop
					route.Gateway = &nh.IP
					route.BGP.NextHop = &nh.IP
				}
			}

			route.BGP.Communities = make(api.Communities, 0)       // TODO
			route.BGP.LargeCommunities = make(api.Communities, 0)  // TODO
			route.BGP.ExtCommunities = make(api.ExtCommunities, 0) // TODO
			route.BGP.LocalPref = data.LocPrf
			route.BGP.Med = data.Metric

			route.Metric = data.Metric

			if route.NeighborID != nil {
				importedRoutes = append(importedRoutes, &route)
			}
		}
	}

	// Filtered (need to hit every neighbor's "filtered" routes)
	// TODO

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: &api.Meta{},
		},
		Imported: importedRoutes,
		Filtered: api.Routes{}, // TODO
	}

	return response, nil
}

// func (src *SingleTableFrrProxy) AllRoutes(
// 	ctx context.Context,
// ) (*api.RoutesResponse, error) {
// 	mainTable := src.GenericFrrProxy.config.MainTable

// 	// Routes received
// 	routes := make(map[string]*http.Response)
// 	for _, ipVersion := range []string{"ipv4", "ipv6"} {
// 		// First fetch all routes from the configured "main_table" for each AFI
// 		res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+ipVersion)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer res.Body.Close()

// 		routes[ipVersion] = res
// 	}

// 	// meta, frrImported, err := parseRoutesResponseStream(routes, src.config)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	response := &api.RoutesResponse{
// 		Response: api.Response{
// 			// Meta: meta,
// 			Meta: &api.Meta{},
// 		},
// 		// Imported: frrImported,
// 		Imported: api.Routes{}, // TODO imported routes
// 		Filtered: api.Routes{}, // TODO filtered routes
// 	}

// 	return response, nil
// }
