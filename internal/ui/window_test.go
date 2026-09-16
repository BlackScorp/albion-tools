package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/blackscorp/albion-helper/internal/arbitrage"
	"github.com/blackscorp/albion-helper/internal/catalog"
)

func TestNewWindowBuildsWalkingSkeleton(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := NewWindow(a)
	if w.Title() != "Albion Helper" {
		t.Fatalf("window title = %q, want Albion Helper", w.Title())
	}
	if w.Content() == nil {
		t.Fatal("window content is nil")
	}
}

func TestReadCriteriaAndFilterRows(t *testing.T) {
	c, err := readCriteria("sword", "Weapons", "4", "0", "100", "25.5")
	if err != nil {
		t.Fatal(err)
	}
	items := map[string]catalog.Item{
		"base":      {ID: "base", Name: "Broadsword", Category: "Weapons / Swords", Tier: 4, Enchantment: 0},
		"enchanted": {ID: "enchanted", Name: "Broadsword", Category: "Weapons / Swords", Tier: 4, Enchantment: 1},
		"cape":      {ID: "cape", Name: "Cape", Category: "Accessories / Capes", Tier: 4},
	}
	good := marketRow{item: "Broadsword · T4.0 · Q1", itemID: "base", opportunity: arbitrage.Opportunity{Profit: 100, ROI: 25.5}}
	wrongEnchantment := marketRow{itemID: "enchanted", opportunity: arbitrage.Opportunity{Profit: 200, ROI: 50}}
	wrongCategory := marketRow{itemID: "cape", opportunity: arbitrage.Opportunity{Profit: 200, ROI: 50}}
	filtered := filterRows([]marketRow{good, wrongEnchantment, wrongCategory}, items, c)
	if len(filtered) != 1 || filtered[0].itemID != "base" {
		t.Fatalf("filtered rows = %#v, want base only", filtered)
	}
}

func TestReadCriteriaRejectsInvalidThresholds(t *testing.T) {
	for _, args := range [][2]string{{"-1", ""}, {"abc", ""}, {"", "-0.1"}, {"", "NaN"}} {
		if _, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", args[0], args[1]); err == nil {
			t.Errorf("readCriteria(%q, %q) succeeded", args[0], args[1])
		}
	}
}

func TestSortRowsHandlesNumericColumnsBothWays(t *testing.T) {
	rows := []marketRow{
		{itemID: "low", opportunity: arbitrage.Opportunity{Profit: 9, DataAge: time.Minute}},
		{itemID: "high", opportunity: arbitrage.Opportunity{Profit: 100, DataAge: 2 * time.Minute}},
	}
	sortRows(rows, 6, false)
	if rows[0].itemID != "high" {
		t.Fatalf("descending profit order = %q first", rows[0].itemID)
	}
	sortRows(rows, 8, true)
	if rows[0].itemID != "low" {
		t.Fatalf("ascending age order = %q first", rows[0].itemID)
	}
}

func TestOpportunityRowsUsesArbitrageAndDropsUnknownItems(t *testing.T) {
	items := map[string]catalog.Item{"sword": {ID: "sword", Name: "Sword", Category: "Weapons / Swords", Tier: 4}}
	prices := []catalog.Price{
		{ItemID: "sword", Market: catalog.Thetford, Quality: 1, Buy: 100, UpdatedAt: 1000},
		{ItemID: "sword", Market: catalog.Martlock, Quality: 1, Sell: 140, UpdatedAt: 1000},
		{ItemID: "unknown", Market: catalog.Martlock, Quality: 1, Buy: 100, UpdatedAt: 1000},
	}
	rows := opportunityRows(prices, items, time.Unix(1060, 0))
	if len(rows) != 1 || rows[0].opportunity.Profit != 40 {
		t.Fatalf("opportunity rows = %#v", rows)
	}
}
