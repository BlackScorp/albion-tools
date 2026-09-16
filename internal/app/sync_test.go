package app

import (
	"context"
	"errors"
	"testing"

	"github.com/blackscorp/albion-helper/internal/catalog"
	"github.com/blackscorp/albion-helper/internal/marketapi"
)

type fakeFetcher struct {
	prices []catalog.Price
	err    error
}

func (f fakeFetcher) FetchPrices(context.Context, []string) ([]catalog.Price, error) {
	return f.prices, f.err
}

type fakeStore struct {
	prices []catalog.Price
	writes int
	err    error
}

func (s *fakeStore) UpsertPrices(_ context.Context, prices []catalog.Price) error {
	s.writes++
	if s.err != nil {
		return s.err
	}
	s.prices = append(s.prices, prices...)
	return nil
}

func TestSyncReturnsNewObservationsButKeepsOlderStoredPrices(t *testing.T) {
	old := catalog.Price{ItemID: "T4_WOOD", Market: catalog.Thetford, Buy: 10}
	newPrice := catalog.Price{ItemID: "T4_CAPE", Market: catalog.Bridgewatch, Buy: 20}
	store := &fakeStore{prices: []catalog.Price{old}}
	got, err := (SyncService{Fetcher: fakeFetcher{prices: []catalog.Price{newPrice}}, Store: store}).Sync(context.Background(), marketapi.Europe, []string{"T4_CAPE"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != newPrice {
		t.Fatalf("Sync() returned %v, want only this sync's observation %v", got, newPrice)
	}
	if len(store.prices) != 2 || store.prices[0] != old || store.prices[1] != newPrice {
		t.Fatalf("stored prices = %v, want old and new observations", store.prices)
	}
}
func (s *fakeStore) Prices(context.Context) ([]catalog.Price, error) {
	return append([]catalog.Price(nil), s.prices...), nil
}

func TestSyncFetchesThenStoresAndReturnsPrices(t *testing.T) {
	want := []catalog.Price{{ItemID: "T4_MAIN_SWORD", Market: catalog.Martlock, Buy: 10, Sell: 12}}
	store := &fakeStore{}
	got, err := (SyncService{Fetcher: fakeFetcher{prices: want}, Store: store}).Sync(context.Background(), marketapi.Europe, []string{"T4_MAIN_SWORD"})
	if err != nil {
		t.Fatal(err)
	}
	if store.writes != 1 || len(got) != 1 || got[0] != want[0] {
		t.Fatalf("stored=%d got=%v", store.writes, got)
	}
}

func TestSyncFetchFailureDoesNotWrite(t *testing.T) {
	store := &fakeStore{prices: []catalog.Price{{ItemID: "old"}}}
	_, err := (SyncService{Fetcher: fakeFetcher{err: errors.New("offline")}, Store: store}).Sync(context.Background(), marketapi.Europe, []string{"T4_MAIN_SWORD"})
	if err == nil || store.writes != 0 || len(store.prices) != 1 || store.prices[0].ItemID != "old" {
		t.Fatalf("err=%v store=%+v", err, store)
	}
}
