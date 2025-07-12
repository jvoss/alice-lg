package frrproxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/alice-lg/alice-lg/pkg/api"
)

// SingleTableFrrProxy is an Alice source
type SingleTableFrrProxy struct {
	GenericFrrProxy
}

// Neighbors implements FrrProxy.
func (src *SingleTableFrrProxy) Neighbors(ctx context.Context) (*api.NeighborsResponse, error) {
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp summary json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parseNeighborsResponse(res.Body)
}

// NeighborsStatus implements FrrProxy.
func (src *SingleTableFrrProxy) NeighborsStatus(ctx context.Context) (*api.NeighborsStatusResponse, error) {
	res, err := src.client.RunCommand(ctx, "bgpd", "show bgp summary json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parseNeighborsStatusResponse(res.Body)
}

// NeighborsSummary implements FrrProxy.
func (src *SingleTableFrrProxy) NeighborsSummary(context.Context) (*api.NeighborsResponse, error) {
	panic("unimplemented")
}

// Routes implements FrrProxy.
func (src *SingleTableFrrProxy) Routes(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return src.fetchRoutes(ctx, neighborID, "routes")
}

// RoutesFiltered implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesFiltered(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return src.fetchRoutes(ctx, neighborID, "filtered")
}

// RoutesNotExported implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesNotExported(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return nil, errors.New("not implemented")
}

// RoutesReceived implements FrrProxy.
func (src *SingleTableFrrProxy) RoutesReceived(ctx context.Context, neighborID string) (*api.RoutesResponse, error) {
	return src.fetchRoutes(ctx, neighborID, "received")
}

func (src *SingleTableFrrProxy) fetchRoutes(ctx context.Context, neighborID string, routesType string) (*api.RoutesResponse, error) {
	// Determine FRR command based on routesType
	command := ""
	switch routesType {
	case "routes":
		command = "show bgp neighbor " + neighborID + " routes json"
	case "filtered":
		command = "show bgp neighbor " + neighborID + " filtered-routes json"
	case "received":
		command = "show bgp neighbor " + neighborID + " received-routes json"
	}

	res, err := src.client.RunCommand(ctx, "bgpd", command)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// The response for a single neighbor is just the routes object
	meta, routes, err := parseRoutesResponseStream(map[string]io.Reader{"": res.Body}, src.config)
	if err != nil {
		return nil, err
	}

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: meta,
		},
		Imported: routes,
	}

	return response, nil
}

// Status implements FrrProxy.
func (src *SingleTableFrrProxy) Status(ctx context.Context) (*api.StatusResponse, error) {
	res, err := src.client.RunCommand(ctx, "vtysh", "show version json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parseStatusResponse(res.Body)
}

// AllRoutes retrieves a route dump
func (src *SingleTableFrrProxy) AllRoutes(
	ctx context.Context,
) (*api.RoutesResponse, error) {
	// First fetch all routes from the configured "main_table"
	mainTable := src.GenericFrrProxy.config.MainTable

	// Routes received
	routes := make(map[string]io.Reader)
	for _, ipVersion := range []string{"ipv4", "ipv6"} {
		res, err := src.client.RunCommand(ctx, "bgpd", "show bgp vrf "+mainTable+" "+ipVersion+" json")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		routes[ipVersion] = res.Body
	}

	meta, frrImported, err := parseRoutesResponseStream(routes, src.config)
	if err != nil {
		return nil, err
	}

	response := &api.RoutesResponse{
		Response: api.Response{
			Meta: meta,
		},
		Imported: frrImported,
		Filtered: api.Routes{}, // TODO filtered routes
	}

	return response, nil
}
