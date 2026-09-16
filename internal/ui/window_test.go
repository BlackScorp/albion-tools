package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
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
		"base":      {ID: "base", Name: "Broadsword", Category: catalog.SwordsCategory, Tier: 4, Enchantment: 0},
		"enchanted": {ID: "enchanted", Name: "Broadsword", Category: catalog.SwordsCategory, Tier: 4, Enchantment: 1},
		"cape":      {ID: "cape", Name: "Cape", Category: catalog.CapesCategory, Tier: 4},
	}
	good := marketRow{item: "Broadsword · T4.0 · Q1", itemID: "base", opportunity: arbitrage.Opportunity{Profit: 100, ROI: 25.5}, hasOpportunity: true}
	wrongEnchantment := marketRow{itemID: "enchanted", opportunity: arbitrage.Opportunity{Profit: 200, ROI: 50}, hasOpportunity: true}
	wrongCategory := marketRow{itemID: "cape", opportunity: arbitrage.Opportunity{Profit: 200, ROI: 50}, hasOpportunity: true}
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
	items := map[string]catalog.Item{"sword": {ID: "sword", Name: "Sword", Category: catalog.SwordsCategory, Tier: 4}}
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

func TestPaginationLimitsDisplayedAndSynchronizedItemsToFifty(t *testing.T) {
	all := make([]marketRow, 123)
	for index := range all {
		all[index] = marketRow{itemID: fmt.Sprintf("item-%03d", index)}
	}
	first := pageRows(all, 0, maxItemsPerPage)
	last := pageRows(all, 2, maxItemsPerPage)
	if len(first) != 50 || len(last) != 23 || last[0].itemID != "item-100" {
		t.Fatalf("page sizes/last page = %d/%d, first last-page ID %q", len(first), len(last), last[0].itemID)
	}
	if ids := pageItemIDs(last); len(ids) != 23 {
		t.Fatalf("last-page sync IDs = %d, want 23", len(ids))
	}
	if ids := pageItemIDs(all); len(ids) != maxItemsPerPage {
		t.Fatalf("oversized sync IDs = %d, want cap %d", len(ids), maxItemsPerPage)
	}
}

func TestCatalogRowsShowEveryItemAndBestAvailableQuote(t *testing.T) {
	items := []catalog.Item{
		{ID: "quoted", Name: "Breitschwert", Category: catalog.SwordsCategory, Tier: 4},
		{ID: "unquoted", Name: "Cape", Category: catalog.CapesCategory, Tier: 4},
	}
	opportunity := marketRow{itemID: "quoted", hasOpportunity: true, opportunity: arbitrage.Opportunity{Quality: 2, Profit: 25}}
	rows := catalogRows(items, []marketRow{opportunity})
	if len(rows) != 2 {
		t.Fatalf("catalog row count = %d, want all two items", len(rows))
	}
	var quoted, unquoted marketRow
	for _, row := range rows {
		if row.itemID == "quoted" {
			quoted = row
		}
		if row.itemID == "unquoted" {
			unquoted = row
		}
	}
	if !quoted.hasOpportunity || !strings.Contains(quoted.item, "Q2") {
		t.Errorf("quoted item row = %+v", quoted)
	}
	if unquoted.hasOpportunity || !strings.Contains(unquoted.item, "Cape") {
		t.Errorf("unquoted item row = %+v", unquoted)
	}
}

func TestBestPriceRowsShowsQuotesWithoutAProfitableSpread(t *testing.T) {
	items := map[string]catalog.Item{
		"sword": {ID: "sword", Name: "Breitschwert", Category: catalog.SwordsCategory, Tier: 4},
	}
	prices := []catalog.Price{
		{ItemID: "sword", Market: catalog.Thetford, Quality: 1, Buy: 120, UpdatedAt: 1000},
		{ItemID: "sword", Market: catalog.Martlock, Quality: 1, Sell: 100, UpdatedAt: 1010},
	}
	rows := bestPriceRows(prices, items, time.Unix(1020, 0))
	if len(rows) != 1 || rows[0].opportunity.Profit != -20 || !rows[0].hasOpportunity {
		t.Fatalf("best quote rows = %#v, want visible -20 silver spread", rows)
	}
}

func TestCategoryMenuBuildsExpandableSubcategories(t *testing.T) {
	selected := ""
	menu := categoryMenu(func(category string) { selected = category })
	var weapons *fyne.MenuItem
	for _, item := range menu.Items {
		if item.Label == "Weapons" {
			weapons = item
			break
		}
	}
	if weapons == nil || weapons.ChildMenu == nil {
		t.Fatal("Weapons category has no expandable submenu")
	}
	for _, item := range weapons.ChildMenu.Items {
		if item.Label == "Weapons (alle)" {
			item.Action()
			break
		}
	}
	if selected != "Weapons" {
		t.Fatalf("selecting parent category = %q, want Weapons", selected)
	}
}
