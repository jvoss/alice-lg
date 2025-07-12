package frrproxy

import (
	"github.com/alice-lg/alice-lg/pkg/caches"
	"github.com/alice-lg/alice-lg/pkg/sources"
)

// FrrProxy is a variant of an alice source and
// implements different strategies for fetching
// route information from FRRouting.
type FrrProxy interface {
	sources.Source
}

// GenericFrrProxy is a source for Alice.
type GenericFrrProxy struct {
	config Config
	client *Client

	// Caches: Neighbors
	neighborsCache *caches.NeighborsCache

	// Caches: Routes
	routesRequiredCache *caches.RoutesCache
	// routesReceivedCache    *caches.RoutesCache
	// routesFilteredCache    *caches.RoutesCache
	routesNotExportedCache *caches.RoutesCache

	// Mutices:
	routesFetchMutex *LockMap
}

// NewFrrProxy creates a new FrrProxy instance.
// This might be either a SingleTableFrrProxy or MultiTableFrrProxy.
func NewFrrProxy(config Config) FrrProxy {
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
	routesNotExportedCache := caches.NewRoutesCache(
		routesCacheDisabled, routesCacheMaxSize)

	var frrProxy FrrProxy

	if config.Type == "single_table" {
		singleTableFrrProxy := new(SingleTableFrrProxy)

		singleTableFrrProxy.config = config
		singleTableFrrProxy.client = client

		singleTableFrrProxy.neighborsCache = neighborsCache

		singleTableFrrProxy.routesRequiredCache = routesRequiredCache
		singleTableFrrProxy.routesNotExportedCache = routesNotExportedCache

		singleTableFrrProxy.routesFetchMutex = NewLockMap()

		frrProxy = singleTableFrrProxy
	}
	// else if config.Type == "multi_table" {
	// 	multiTableFrrProxy := new(MultiTableFrrProxy)

	// 	multiTableFrrProxy.config = config
	// 	multiTableFrrProxy.client = client

	// 	multiTableFrrProxy.neighborsCache = neighborsCache

	// 	multiTableFrrProxy.routesRequiredCache = routesRequiredCache
	// 	multiTableFrrProxy.routesNotExportedCache = routesNotExportedCache

	// 	multiTableFrrProxy.routesFetchMutex = NewLockMap()

	// 	frrProxy = multiTableFrrProxy
	// }

	return frrProxy
}

// ExpireCaches clears all local caches
func (b *GenericFrrProxy) ExpireCaches() int {
	count := b.routesRequiredCache.Expire()
	count += b.routesNotExportedCache.Expire()
	return count
}
