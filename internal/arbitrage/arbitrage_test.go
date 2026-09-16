package arbitrage

import (
	"testing"
	"time"

	"github.com/blackscorp/albion-helper/internal/catalog"
)

func TestCalculateProfitableDirectedPairsAndConservativeAge(t *testing.T) {
	now := time.Unix(2_000, 0)
	prices := []catalog.Price{
		{ItemID: "sword", Market: catalog.Thetford, Quality: 2, Buy: 100, Sell: 130, UpdatedAt: 1_900},
		{ItemID: "sword", Market: catalog.Martlock, Quality: 2, Buy: 150, Sell: 180, UpdatedAt: 1_800},
		{ItemID: "sword", Market: catalog.Martlock, Quality: 3, Buy: 50, Sell: 500, UpdatedAt: 1_900},
	}

	got, err := Calculate(prices, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d opportunities, want 1: %+v", len(got), got)
	}
	forward := got[0]
	if forward.BuyMarket != catalog.Thetford || forward.SellMarket != catalog.Martlock {
		t.Fatalf("unexpected opportunity direction: %+v", forward)
	}
	if forward.ItemID != "sword" || forward.Quality != 2 || forward.BuyPrice != 100 || forward.SellPrice != 180 {
		t.Fatalf("unexpected opportunity quotes: %+v", forward)
	}
	if forward.Profit != 80 || forward.ROI != 80 || forward.Range != 1 || forward.DataAge != 200*time.Second {
		t.Fatalf("unexpected calculated values: %+v", forward)
	}
}

func TestCalculateExactValuesAndExclusions(t *testing.T) {
	now := time.Unix(2_000, 0)
	prices := []catalog.Price{
		{ItemID: "good", Market: catalog.Thetford, Quality: 1, Buy: 100, UpdatedAt: 1_900},
		{ItemID: "good", Market: catalog.FortSterling, Quality: 1, Sell: 125, UpdatedAt: 1_800},
		{ItemID: "zero", Market: catalog.Thetford, Buy: 0, UpdatedAt: 1_900},
		{ItemID: "zero", Market: catalog.FortSterling, Sell: 100, UpdatedAt: 1_900},
		{ItemID: "loss", Market: catalog.Thetford, Buy: 120, UpdatedAt: 1_900},
		{ItemID: "loss", Market: catalog.FortSterling, Sell: 100, UpdatedAt: 1_900},
		{ItemID: "same", Market: catalog.Caerleon, Buy: 10, Sell: 30, UpdatedAt: 1_900},
		{ItemID: "same", Market: catalog.BlackMarket, Buy: 20, Sell: 5, UpdatedAt: 1_900},
	}

	got, err := Calculate(prices, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d opportunities, want one Royal and one special-market route: %+v", len(got), got)
	}
	royal := got[0]
	if royal.ItemID != "good" || royal.BuyPrice != 100 || royal.SellPrice != 125 || royal.Profit != 25 || royal.ROI != 25 || royal.Range != 1 || royal.DataAge != 200*time.Second {
		t.Errorf("unexpected calculated values: %+v", royal)
	}
	special := got[1]
	if special.ItemID != "same" || special.BuyMarket != catalog.BlackMarket || special.SellMarket != catalog.Caerleon || special.Range != 1 {
		t.Errorf("special market pair not represented: %+v", special)
	}
}

func TestCalculateRejectsUnknownMarket(t *testing.T) {
	_, err := Calculate([]catalog.Price{{ItemID: "item", Market: "unknown"}}, time.Now())
	if err == nil {
		t.Fatal("expected unknown market error")
	}
}
