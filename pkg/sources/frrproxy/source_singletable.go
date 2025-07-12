package frrproxy

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/alice-lg/alice-lg/pkg/api"
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

	// Fetch neighbors from the configured "main_table" for each AFI
	for _, ipVersion := range []string{"ipv4", "ipv6"} {
		res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+ipVersion+" neighbors")
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
					ID:             PeerHash(info.RemoteAs, ip),
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
	}

	for _, n := range neighborsResponse {
		log.Printf("Neighbor IP: %s | Remote AS: %d | Uptime: %s | Accepted: %d | Exported: %d | Filtered: %d", n.Address, n.ASN, n.Uptime, n.RoutesAccepted, n.RoutesExported, n.RoutesFiltered)
		response.Neighbors = append(response.Neighbors, &n)
	}

	return &response, nil
}

// NeighborsStatus implements FrrProxy.
func (src *SingleTableFrrProxy) NeighborsStatus(context.Context) (*api.NeighborsStatusResponse, error) {
	panic("unimplemented")
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
	panic("unimplemented")
}

// AllRoutes implements FrrProxy.
func (src *SingleTableFrrProxy) AllRoutes(context.Context) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// AllRoutes retrieves a route dump
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
