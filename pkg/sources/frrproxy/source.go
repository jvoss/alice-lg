package frrproxy

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/alice-lg/alice-lg/pkg/api"
	"github.com/alice-lg/alice-lg/pkg/caches"
	"github.com/alice-lg/alice-lg/pkg/pools"
)

// Ensure source implements the interface
// var _FrrProxySource sources.Source = &FrrProxy{}

// FrrProxy is a source for Alice
type FrrProxy struct {
	config Config
	client *Client

	// Caches: Neighbors
	neighborsCache *caches.NeighborsCache

	// Caches: Routes
	routesRequiredCache    *caches.RoutesCache
	routesReceivedCache    *caches.RoutesCache
	routesFilteredCache    *caches.RoutesCache
	routesNotExportedCache *caches.RoutesCache

	// Mutices:
	routesFetchMutex *LockMap
}

// NewFrrProxy creates a new FrrProxy instance.
func NewFrrProxy(config Config) *FrrProxy {
	client := NewClient(config)

	// Cache settings:
	// TODO: Maybe read from config file
	neighborsCacheDisable := false

	routesCacheDisabled := false
	routesCacheMaxSize := 128

	// Initialize caches
	neighborsCache := caches.NewNeighborsCache(neighborsCacheDisable)
	routesRequiredCache := caches.NewRoutesCache(
		routesCacheDisabled, routesCacheMaxSize)
	routesReceivedCache := caches.NewRoutesCache(
		routesCacheDisabled, routesCacheMaxSize)
	routesFilteredCache := caches.NewRoutesCache(
		routesCacheDisabled, routesCacheMaxSize)
	routesNotExportedCache := caches.NewRoutesCache(
		routesCacheDisabled, routesCacheMaxSize)

	return &FrrProxy{
		config: config,
		client: client,

		neighborsCache: neighborsCache,

		routesRequiredCache:    routesRequiredCache,
		routesReceivedCache:    routesReceivedCache,
		routesFilteredCache:    routesFilteredCache,
		routesNotExportedCache: routesNotExportedCache,
	}
}

// ExpireCaches clears all local caches
func (frr *FrrProxy) ExpireCaches() int {
	count := frr.routesRequiredCache.Expire()
	count += frr.routesNotExportedCache.Expire()
	return count
}

// Neighbors retrieves a list of neighbors
func (src *FrrProxy) Neighbors(
	ctx context.Context,
) (*api.NeighborsResponse, error) {
	vrf := src.config.Vrf

	response := api.NeighborsResponse{}
	response.Neighbors = make(api.Neighbors, 0)

	var neighborsResponse = make(map[string]api.Neighbor, 0)

	// Fetch neighbors from the configured "main_table" for the configured AFI
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+vrf+" "+src.client.afi+" neighbors")
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
				ID:             ip,
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
func (src *FrrProxy) NeighborsStatus(
	ctx context.Context,
) (*api.NeighborsStatusResponse, error) {
	log.Printf("I'm here.................TWO..............")
	// panic("unimplemented")

	log.Printf("NeighborsStatus")

	vrf := src.config.Vrf

	response := api.NeighborsStatusResponse{}
	response.Neighbors = make(api.NeighborsStatus, 0)

	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+vrf+" "+src.client.afi+" summary")
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

	for ip := range bgpSummary.Peers {
		ns := api.NeighborStatus{}
		ns.ID = PeerHash(src.config.ID, ip)
		ns.State = "up" // TODO
		ns.Since = 5 * time.Second

		response.Neighbors = append(response.Neighbors, &ns)
	}

	return &response, nil
}

// NeighborsSummary implements FrrProxy.
func (src *FrrProxy) NeighborsSummary(context.Context) (*api.NeighborsResponse, error) {
	return &api.NeighborsResponse{}, nil
}

// Routes implements FrrProxy.
func (src *FrrProxy) Routes(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return &api.RoutesResponse{}, nil
}

// RoutesFiltered implements FrrProxy.
func (src *FrrProxy) RoutesFiltered(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	vrf := src.config.Vrf

	routes := api.Routes{}

	// Fetch routes from the configured VRF for the configured AFI
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+vrf+" "+src.client.afi+" neighbor "+neighborID+" filtered-routes")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data BgpNeighborRoutes

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}

	for prefix, routeData := range data.Routes {
		// log.Printf("Prefix: %s", prefix)

		for _, data := range routeData {
			route := api.Route{}
			route.Network = prefix
			route.Interface = pools.Interfaces.Acquire("unknown")
			route.BGP = &api.BGPInfo{}
			route.Type = pools.Types.Acquire([]string{"BGP"})

			route.NeighborID = pools.Neighbors.Acquire(neighborID)

			// Age
			// epoch := data.LastUpdate.Epoch
			// then := time.Unix(int64(epoch), 0)
			// route.Age = time.Since(then)

			// if data.Bestpath.Overall {
			// 	route.Primary = true
			// } else {
			// 	route.Primary = false
			// }

			origin := data.Origin

			route.BGP.Origin = &origin
			route.BGP.AsPath = data.AsPath()

			for _, nexthop := range data.Nexthops {
				if nexthop.IP != "::" {
					nh := nexthop
					route.Gateway = &nh.IP
					route.BGP.NextHop = &nh.IP
				}
			}

			// route.BGP.Communities = parseBgpCommunityList(data.Community.List)
			// route.BGP.LargeCommunities = parseBgpCommunityList(data.LargeCommunity.List)
			// route.BGP.ExtCommunities = parseExtBgpCommunities(data.ExtCommunity.String)
			// route.BGP.LocalPref = data.LocPrf
			// route.BGP.Med = data.Metric

			// route.Metric = data.Metric

			if route.NeighborID != nil {
				routes = append(routes, &route)
			}
		}
	}

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: &api.Meta{},
		},
		Imported: nil,
		Filtered: routes,
	}

	return response, nil
}

// RoutesNotExported implements FrrProxy.
func (src *FrrProxy) RoutesNotExported(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return &api.RoutesResponse{}, nil
}

// RoutesReceived implements FrrProxy.
func (src *FrrProxy) RoutesReceived(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	vrf := src.config.Vrf

	routes := api.Routes{}

	// Fetch routes from the configured VRF for the configured AFI
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+vrf+" "+src.client.afi+" neighbor "+neighborID+" routes")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data BgpNeighborRoutes

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}

	for prefix, routeData := range data.Routes {
		// log.Printf("Prefix: %s", prefix)

		for _, data := range routeData {
			route := api.Route{}
			route.Network = prefix
			route.Interface = pools.Interfaces.Acquire("unknown")
			route.BGP = &api.BGPInfo{}
			route.Type = pools.Types.Acquire([]string{"BGP"})

			route.NeighborID = pools.Neighbors.Acquire(neighborID)

			// Age
			// epoch := data.LastUpdate.Epoch
			// then := time.Unix(int64(epoch), 0)
			// route.Age = time.Since(then)

			// if data.Bestpath.Overall {
			// 	route.Primary = true
			// } else {
			// 	route.Primary = false
			// }

			origin := data.Origin

			route.BGP.Origin = &origin
			route.BGP.AsPath = data.AsPath()

			for _, nexthop := range data.Nexthops {
				if nexthop.IP != "::" {
					nh := nexthop
					route.Gateway = &nh.IP
					route.BGP.NextHop = &nh.IP
				}
			}

			// route.BGP.Communities = parseBgpCommunityList(data.Community.List)
			// route.BGP.LargeCommunities = parseBgpCommunityList(data.LargeCommunity.List)
			// route.BGP.ExtCommunities = parseExtBgpCommunities(data.ExtCommunity.String)
			// route.BGP.LocalPref = data.LocPrf
			// route.BGP.Med = data.Metric

			// route.Metric = data.Metric

			if route.NeighborID != nil {
				routes = append(routes, &route)
			}
		}
	}

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: &api.Meta{},
		},
		Imported: routes,
		Filtered: api.Routes{}, // Caching of filter routes unsupported
	}

	return response, nil
}

// Status implements FrrProxy.
func (src *FrrProxy) Status(context.Context) (*api.StatusResponse, error) {
	// panic("unimplemented")
	response := api.StatusResponse{}
	response.Meta = &api.Meta{}
	// response.Status.ServerTime = time.Now()   // TODO
	// response.Status.LastReboot = time.Now()   // TODO
	// response.Status.LastReconfig = time.Now() // TODO
	response.Status.Message = "status-message"
	response.Status.RouterID = "1.2.3.4"
	response.Status.Version = "version-string-here"
	response.Status.Backend = "FRR"

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
func (src *FrrProxy) AllRoutes(
	ctx context.Context,
) (*api.RoutesResponse, error) {
	vrf := src.config.Vrf

	importedRoutes := api.Routes{}

	// Fetch routes from the configured VRF for the configured AFI
	// Imported
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+vrf+" "+src.client.afi+" detail-routes")
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
			if data.ImportedFrom != "" {
				// Do not process imported routes
				continue
			}

			route := api.Route{}
			route.Network = prefix
			route.Interface = pools.Interfaces.Acquire("unknown")
			route.BGP = &api.BGPInfo{}
			route.Type = pools.Types.Acquire([]string{"BGP"})

			route.NeighborID = pools.Neighbors.Acquire(data.Peer.PeerID)

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

			route.BGP.Communities = parseBgpCommunityList(data.Community.List)
			route.BGP.LargeCommunities = parseBgpCommunityList(data.LargeCommunity.List)
			route.BGP.ExtCommunities = parseExtBgpCommunities(data.ExtCommunity.String)
			route.BGP.LocalPref = data.LocPrf
			route.BGP.Med = data.Metric

			route.Metric = data.Metric

			if route.NeighborID != nil {
				importedRoutes = append(importedRoutes, &route)
			}
		}
	}

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: &api.Meta{},
		},
		Imported: importedRoutes,
		Filtered: api.Routes{}, // Caching of filter routes unsupported
	}

	return response, nil
}
