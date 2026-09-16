package storage

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/blackscorp/albion-helper/internal/catalog"
)

func TestOpenCreatesSchemaAndPersistsPrices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prices.db")
	want := []catalog.Price{{
		ItemID: "T4_WOOD", Market: catalog.FortSterling, Quality: 2,
		Buy: 410, Sell: 525, UpdatedAt: 1_757_900_000,
	}}

	repository, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := repository.UpsertPrices(context.Background(), want); err != nil {
		t.Fatalf("UpsertPrices() error = %v", err)
	}
	if err := repository.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	repository, err = Open(path)
	if err != nil {
		t.Fatalf("reopen error = %v", err)
	}
	defer repository.Close()
	got, err := repository.Prices(context.Background())
	if err != nil {
		t.Fatalf("Prices() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Prices() = %#v, want %#v", got, want)
	}
}

func TestUpsertPricesReplacesExistingQuote(t *testing.T) {
	repository, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer repository.Close()

	price := catalog.Price{ItemID: "T4_CAPE", Market: catalog.Lymhurst, Buy: 100, Sell: 120, UpdatedAt: 10}
	if err := repository.UpsertPrices(context.Background(), []catalog.Price{price}); err != nil {
		t.Fatal(err)
	}
	price.Buy, price.Sell, price.UpdatedAt = 125, 150, 20
	if err := repository.UpsertPrices(context.Background(), []catalog.Price{price}); err != nil {
		t.Fatal(err)
	}

	got, err := repository.Prices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], price) {
		t.Fatalf("Prices() = %#v, want one updated price %#v", got, price)
	}
}

func TestUpsertPricesIsAtomicOnInvalidInput(t *testing.T) {
	repository, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer repository.Close()

	prices := []catalog.Price{
		{ItemID: "T4_WOOD", Market: catalog.Martlock, Buy: 1, Sell: 2, UpdatedAt: 1},
		{Market: catalog.Martlock, Buy: 3, Sell: 4, UpdatedAt: 2},
	}
	if err := repository.UpsertPrices(context.Background(), prices); err == nil {
		t.Fatal("UpsertPrices() succeeded for invalid input")
	}
	got, err := repository.Prices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("Prices() = %#v after failed transaction, want empty", got)
	}
}

func TestDefaultDatabasePathUsesAlbionHelperDirectory(t *testing.T) {
	path, err := DefaultDatabasePath()
	if err != nil {
		t.Fatalf("DefaultDatabasePath() error = %v", err)
	}
	if filepath.Base(path) != "albion-helper.db" || filepath.Base(filepath.Dir(path)) != "Albion Helper" {
		t.Fatalf("DefaultDatabasePath() = %q, want Albion Helper/albion-helper.db", path)
	}
}
