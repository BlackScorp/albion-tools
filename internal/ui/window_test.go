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
	c, err := readCriteria("sword", "Weapons", "4", "0", "Alle Ranges", "100", "25.5")
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
		if _, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "Alle Ranges", args[0], args[1]); err == nil {
			t.Errorf("readCriteria(%q, %q) succeeded", args[0], args[1])
		}
	}
	if _, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "6", "", ""); err == nil {
		t.Fatal("readCriteria accepted unsupported range 6")
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

func TestSortHeaderStateTogglesOnRepeatedClick(t *testing.T) {
	column, ascending := -1, true
	column, ascending = nextSortState(7, column, ascending)
	if column != 7 || !ascending {
		t.Fatalf("first ROI click state = %d/%v, want ascending", column, ascending)
	}
	column, ascending = nextSortState(7, column, ascending)
	if column != 7 || ascending {
		t.Fatalf("second ROI click state = %d/%v, want descending", column, ascending)
	}
}

func TestSortingKeepsRowsWithoutPricesLastInBothDirections(t *testing.T) {
	for _, ascending := range []bool{true, false} {
		rows := []marketRow{
			{itemID: "missing"},
			{itemID: "priced", hasOpportunity: true, opportunity: arbitrage.Opportunity{Profit: 10, Range: 2}},
		}
		sortRows(rows, 6, ascending)
		if rows[0].itemID != "priced" || rows[1].itemID != "missing" {
			t.Errorf("ascending=%v sorted IDs = %s, %s; missing-price row should be last", ascending, rows[0].itemID, rows[1].itemID)
		}
	}
}

func TestRangeFilterSelectsMatchingPricedItemsAndProfitFilterDropsMissing(t *testing.T) {
	items := map[string]catalog.Item{
		"range-two":   {ID: "range-two", Name: "Item", Category: catalog.SwordsCategory},
		"range-three": {ID: "range-three", Name: "Item", Category: catalog.SwordsCategory},
		"missing":     {ID: "missing", Name: "Item", Category: catalog.SwordsCategory},
	}
	rows := []marketRow{
		{itemID: "range-two", hasOpportunity: true, opportunity: arbitrage.Opportunity{Range: 2, Profit: 50}},
		{itemID: "range-three", hasOpportunity: true, opportunity: arbitrage.Opportunity{Range: 3, Profit: 100}},
		{itemID: "missing"},
	}
	byRange, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "2", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := filterRows(rows, items, byRange); len(got) != 1 || got[0].itemID != "range-two" {
		t.Fatalf("range 2 filter = %#v", got)
	}
	byProfit, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "Alle Ranges", "1", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := filterRows(rows, items, byProfit); len(got) != 2 || got[0].itemID == "missing" || got[1].itemID == "missing" {
		t.Fatalf("positive minimum profit filter retained missing-price rows: %#v", got)
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

func TestPaginationLimitsDisplayedItemsToFifty(t *testing.T) {
	all := make([]marketRow, 123)
	for index := range all {
		all[index] = marketRow{itemID: fmt.Sprintf("item-%03d", index)}
	}
	first := pageRows(all, 0, maxItemsPerPage)
	last := pageRows(all, 2, maxItemsPerPage)
	if len(first) != 50 || len(last) != 23 || last[0].itemID != "item-100" {
		t.Fatalf("page sizes/last page = %d/%d, first last-page ID %q", len(first), len(last), last[0].itemID)
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
	if !quoted.hasOpportunity || strings.Contains(quoted.item, "Q2") {
		t.Errorf("quoted item row = %+v", quoted)
	}
	if unquoted.hasOpportunity || !strings.Contains(unquoted.item, "Cape") {
		t.Errorf("unquoted item row = %+v", unquoted)
	}
}

func TestSyncScopeUsesAllFilteredItemsAndRequiresAnItemFilter(t *testing.T) {
	items := make([]catalog.Item, 123)
	for index := range items {
		items[index] = catalog.Item{ID: fmt.Sprintf("wood-%03d", index), Name: "Wood", Tier: 4}
	}
	all, err := readCriteria("", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "Alle Ranges", "", "")
	if err != nil || hasItemScope(all) {
		t.Fatalf("unfiltered criteria scope = %v, error %v", hasItemScope(all), err)
	}
	wood, err := readCriteria("wood", "Alle Kategorien", "Alle Tiers", "Alle Verzauberungen", "Alle Ranges", "", "")
	if err != nil || !hasItemScope(wood) {
		t.Fatalf("search criteria scope = %v, error %v", hasItemScope(wood), err)
	}
	if got := filteredItemIDs(items, wood); len(got) != 123 {
		t.Fatalf("IDs queued for sync = %d, want all 123 matching items", len(got))
	}
}

func TestDefaultMarketSelectionAndMarketPriceFilter(t *testing.T) {
	defaults := defaultCitySelection()
	if strings.Contains(strings.Join(defaults, ","), "Caerleon") || strings.Contains(strings.Join(defaults, ","), "Black Market") || strings.Contains(strings.Join(defaults, ","), "Brecilien") || len(defaults) != 5 {
		t.Fatalf("default city selection = %v", defaults)
	}
	selected := selectedCityMarkets([]string{"Thetford", "Martlock", "Caerleon"})
	if len(selected) != 3 || selected[0] != catalog.Thetford || selected[2] != catalog.Caerleon {
		t.Fatalf("selected cities = %v", selected)
	}
	prices := []catalog.Price{{Market: catalog.Thetford}, {Market: catalog.Caerleon}, {Market: catalog.Brecilien}}
	if got := filterPricesByMarkets(prices, selected); len(got) != 2 {
		t.Fatalf("market-filtered prices = %v, want Thetford and Caerleon", got)
	}
}

func TestItemTitleUsesConciseFamilyNameAndTier(t *testing.T) {
	if got := itemTitle(catalog.Item{Name: "Tasche", Tier: 5, Enchantment: 0}); got != "Tasche T5.0" {
		t.Fatalf("item title = %q, want concise name", got)
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
