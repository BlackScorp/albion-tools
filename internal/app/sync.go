// Package app coordinates price synchronization between the API and local storage.
package app

import (
	"context"
	"fmt"

	"github.com/blackscorp/albion-helper/internal/catalog"
	"github.com/blackscorp/albion-helper/internal/marketapi"
)

type PriceFetcher interface {
	FetchPrices(context.Context, []string) ([]catalog.Price, error)
}

type PriceRepository interface {
	UpsertPrices(context.Context, []catalog.Price) error
	Prices(context.Context) ([]catalog.Price, error)
}

type SyncService struct {
	Fetcher PriceFetcher
	Store   PriceRepository
}

// Sync stores all API observations and returns only observations from this request.
func (s SyncService) Sync(ctx context.Context, server marketapi.Server, itemIDs []string) ([]catalog.Price, error) {
	return s.SyncAt(ctx, server, itemIDs, nil)
}

// SyncAt limits observations to selected markets when the configured fetcher supports it.
func (s SyncService) SyncAt(ctx context.Context, server marketapi.Server, itemIDs []string, markets []catalog.Market) ([]catalog.Price, error) {
	fetcher := s.Fetcher
	if fetcher == nil {
		var err error
		fetcher, err = marketapi.NewClient(server)
		if err != nil {
			return nil, err
		}
	}
	if s.Store == nil {
		return nil, fmt.Errorf("price repository is not configured")
	}
	var prices []catalog.Price
	var err error
	if locationFetcher, ok := fetcher.(interface {
		FetchPricesAt(context.Context, []string, []catalog.Market) ([]catalog.Price, error)
	}); ok {
		prices, err = locationFetcher.FetchPricesAt(ctx, itemIDs, markets)
	} else {
		prices, err = fetcher.FetchPrices(ctx, itemIDs)
	}
	if err != nil {
		return nil, err
	}
	if err := s.Store.UpsertPrices(ctx, prices); err != nil {
		return nil, fmt.Errorf("save synchronized prices: %w", err)
	}
	return prices, nil
}
