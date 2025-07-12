package frrproxy

import (
	"context"

	"github.com/alice-lg/alice-lg/pkg/api"
)

// SingleTableFrrProxy is an Alice source
type SingleTableFrrProxy struct {
	GenericFrrProxy
}

// Neighbors implements FrrProxy.
func (src *SingleTableFrrProxy) Neighbors(context.Context) (*api.NeighborsResponse, error) {
	panic("unimplemented")
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

// Status implements FrrProxy.
func (src *SingleTableFrrProxy) AllRoutes(context.Context) (*api.RoutesResponse, error) {
	panic("unimplemented")
}

// AllRoutes retrieves a route dump
// func (src *SingleTableFrrProxy) AllRoutes(
// 	ctx context.Context,
// ) (*api.RoutesResponse, error) {
// 	// First fetch all routes from the configured "main_table"
// 	mainTable := src.GenericFrrProxy.config.MainTable

// 	// Routes received
// 	routes := make(map[string]*http.Response)
// 	for _, ipVersion := range []string{"ipv4", "ipv6"} {
// 		res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+ipVersion)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer res.Body.Close()

// 		routes[ipVersion] = res
// 	}

// 	meta, frrImported, err := parseRoutesResponseStream(routes, src.config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	response := &api.RoutesResponse{
// 		Response: api.Response{
// 			Meta: meta,
// 		},
// 		Imported: frrImported,
// 		Filtered: api.Routes{}, // TODO filtered routes
// 	}

// 	return response, nil
// }
