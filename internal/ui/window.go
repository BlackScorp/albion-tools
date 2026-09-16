package ui

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	syncpkg "sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/blackscorp/albion-helper/internal/arbitrage"
	"github.com/blackscorp/albion-helper/internal/catalog"
)

type marketRow struct {
	item, itemID string
	opportunity  arbitrage.Opportunity
}

var columns = []string{"Item", "Kaufstadt", "Verkaufsstadt", "Range", "Kaufpreis", "Verkaufspreis", "Bruttogewinn", "ROI (brutto)", "Datenalter"}

type syncFunc func(string, []string) ([]catalog.Price, error)

func NewWindow(a fyne.App) fyne.Window { return NewWindowWithData(a, nil, nil) }

// NewWindowWithData creates the market opportunity view from cached prices.
func NewWindowWithData(a fyne.App, cached []catalog.Price, syncPrices syncFunc) fyne.Window {
	w := a.NewWindow("Albion Helper")
	w.Resize(fyne.NewSize(1200, 720))

	server := widget.NewSelect([]string{"Europe", "Americas", "Asia"}, nil)
	server.SetSelected("Europe")
	search := widget.NewEntry()
	search.SetPlaceHolder("Name, ID oder Kategorie filtern …")
	category := widget.NewSelect(categoryOptions(), nil)
	category.SetSelected("Alle Kategorien")
	tier := widget.NewSelect([]string{"Alle Tiers", "1", "2", "3", "4", "5", "6", "7", "8"}, nil)
	tier.SetSelected("Alle Tiers")
	enchantment := widget.NewSelect([]string{"Alle Verzauberungen", "0", "1", "2", "3", "4"}, nil)
	enchantment.SetSelected("Alle Verzauberungen")
	minProfit := widget.NewEntry()
	minProfit.SetPlaceHolder("Mindestgewinn")
	minROI := widget.NewEntry()
	minROI.SetPlaceHolder("Mindest-ROI %")
	syncButton := widget.NewButton("Preise aktualisieren", nil)
	status := widget.NewLabel("Offline · Gebühren und Transportkosten nicht berücksichtigt")
	if len(cached) > 0 {
		status.SetText("Gespeicherte Preise geladen · Gebühren und Transportkosten nicht berücksichtigt")
	}

	items := catalog.Items()
	itemByID := make(map[string]catalog.Item, len(items))
	for _, item := range items {
		itemByID[item.ID] = item
	}
	allRows := opportunityRows(cached, itemByID, time.Now())
	rows := append([]marketRow(nil), allRows...)
	ascending, sortedColumn := true, -1
	table := widget.NewTable(
		func() (int, int) { return len(rows) + 1, len(columns) },
		func() fyne.CanvasObject { return widget.NewLabel("#######") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				label.SetText(columns[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			r := rows[id.Row-1]
			op := r.opportunity
			values := []string{r.item, string(op.BuyMarket), string(op.SellMarket), fmt.Sprint(op.Range), formatSilver(int(op.BuyPrice)), formatSilver(int(op.SellPrice)), formatSilver(int(op.Profit)), formatROI(int(op.Profit), int(op.BuyPrice)), formatAge(op.DataAge)}
			label.SetText(values[id.Col])
		},
	)
	for col, width := range []float32{220, 130, 140, 65, 110, 120, 125, 105, 100} {
		table.SetColumnWidth(col, width)
	}
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row != 0 {
			return
		}
		if sortedColumn == id.Col {
			ascending = !ascending
		} else {
			sortedColumn, ascending = id.Col, true
		}
		sortRows(rows, id.Col, ascending)
		table.Refresh()
	}
	applyFilters := func() {
		criteria, err := readCriteria(search.Text, category.Selected, tier.Selected, enchantment.Selected, minProfit.Text, minROI.Text)
		if err != nil {
			status.SetText("Filterfehler: " + err.Error())
			return
		}
		rows = filterRows(allRows, itemByID, criteria)
		if len(allRows) == 0 {
			status.SetText("Keine gespeicherten Preise · Offline")
		} else if len(rows) == 0 {
			status.SetText("Keine passenden profitablen Handelschancen")
		} else {
			status.SetText(fmt.Sprintf("%d Handelschancen · Gebühren und Transportkosten nicht berücksichtigt", len(rows)))
		}
		table.Refresh()
	}
	search.OnChanged = func(string) { applyFilters() }
	category.OnChanged = func(string) { applyFilters() }
	tier.OnChanged = func(string) { applyFilters() }
	enchantment.OnChanged = func(string) { applyFilters() }
	minProfit.OnChanged = func(string) { applyFilters() }
	minROI.OnChanged = func(string) { applyFilters() }

	var syncMu syncpkg.Mutex
	if syncPrices != nil {
		syncButton.OnTapped = func() {
			if !syncMu.TryLock() {
				return
			}
			criteria, err := readCriteria(search.Text, category.Selected, tier.Selected, enchantment.Selected, minProfit.Text, minROI.Text)
			if err != nil {
				syncMu.Unlock()
				status.SetText("Filterfehler: " + err.Error())
				return
			}
			ids := make([]string, 0)
			for _, item := range items {
				if itemMatches(item, criteria) {
					ids = append(ids, item.ID)
				}
			}
			selectedServer := server.Selected
			syncButton.Disable()
			status.SetText("Synchronisierung läuft …")
			go func() {
				prices, err := syncPrices(selectedServer, ids)
				fyne.Do(func() {
					defer syncMu.Unlock()
					syncButton.Enable()
					if err != nil {
						status.SetText("Synchronisierung fehlgeschlagen: " + err.Error())
						return
					}
					allRows = opportunityRows(prices, itemByID, time.Now())
					applyFilters()
					status.SetText(fmt.Sprintf("Synchronisierung %s abgeschlossen · %d Handelschancen · Gebühren und Transportkosten nicht berücksichtigt", time.Now().Format("15:04:05"), len(allRows)))
				})
			}()
		}
	}

	filters := container.NewHBox(widget.NewLabel("Kategorie"), category, tier, enchantment, minProfit, minROI)
	toolbar := container.NewBorder(nil, nil, widget.NewLabel("Server:"), syncButton, server)
	w.SetContent(container.NewBorder(container.NewVBox(toolbar, search, filters, status), nil, nil, nil, table))
	return w
}

type criteria struct {
	query, category   string
	tier, enchantment int
	minProfit         int64
	minROI            float64
}

func readCriteria(query, category, tier, enchantment, profit, roi string) (criteria, error) {
	c := criteria{query: strings.TrimSpace(query), category: category, enchantment: -1}
	if tier != "" && tier != "Alle Tiers" {
		parsed, err := strconv.Atoi(tier)
		if err != nil {
			return c, fmt.Errorf("Tier muss eine Zahl sein")
		}
		c.tier = parsed
	}
	if enchantment != "" && enchantment != "Alle Verzauberungen" {
		parsed, err := strconv.Atoi(enchantment)
		if err != nil {
			return c, fmt.Errorf("Verzauberung muss eine Zahl sein")
		}
		c.enchantment = parsed
	}
	if strings.TrimSpace(profit) != "" {
		parsed, err := strconv.ParseInt(strings.TrimSpace(profit), 10, 64)
		if err != nil || parsed < 0 {
			return c, fmt.Errorf("Mindestgewinn muss eine nichtnegative ganze Zahl sein")
		}
		c.minProfit = parsed
	}
	if strings.TrimSpace(roi) != "" {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(roi), 64)
		if err != nil || parsed < 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return c, fmt.Errorf("Mindest-ROI muss eine nichtnegative Zahl sein")
		}
		c.minROI = parsed
	}
	return c, nil
}

func itemMatches(item catalog.Item, c criteria) bool {
	if c.query != "" && !containsFold(item.Name+" "+item.ID+" "+item.Category, c.query) {
		return false
	}
	if c.category != "" && c.category != "Alle Kategorien" && item.Category != c.category && !strings.HasPrefix(item.Category, c.category+" / ") {
		return false
	}
	if c.tier > 0 && item.Tier != c.tier {
		return false
	}
	if c.enchantment >= 0 && item.Enchantment != c.enchantment {
		return false
	}
	return true
}

func filterRows(rows []marketRow, items map[string]catalog.Item, c criteria) []marketRow {
	filtered := make([]marketRow, 0, len(rows))
	for _, row := range rows {
		item, ok := items[row.itemID]
		if !ok || !itemMatches(item, c) || row.opportunity.Profit < c.minProfit || row.opportunity.ROI < c.minROI {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func opportunityRows(prices []catalog.Price, items map[string]catalog.Item, now time.Time) []marketRow {
	opportunities, _ := arbitrage.Calculate(prices, now)
	rows := make([]marketRow, 0, len(opportunities))
	for _, op := range opportunities {
		item, ok := items[op.ItemID]
		if !ok {
			continue
		}
		name := fmt.Sprintf("%s · T%d.%d · Q%d", item.Name, item.Tier, item.Enchantment, op.Quality)
		rows = append(rows, marketRow{item: name, itemID: item.ID, opportunity: op})
	}
	return rows
}

func sortRows(rows []marketRow, col int, ascending bool) {
	lessValue := func(a, b marketRow) bool {
		x, y := a.opportunity, b.opportunity
		switch col {
		case 0:
			return a.item < b.item
		case 1:
			return x.BuyMarket < y.BuyMarket
		case 2:
			return x.SellMarket < y.SellMarket
		case 3:
			return x.Range < y.Range
		case 4:
			return x.BuyPrice < y.BuyPrice
		case 5:
			return x.SellPrice < y.SellPrice
		case 6:
			return x.Profit < y.Profit
		case 7:
			return x.ROI < y.ROI
		case 8:
			return x.DataAge < y.DataAge
		default:
			return false
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if ascending {
			return lessValue(rows[i], rows[j])
		}
		return lessValue(rows[j], rows[i])
	})
}

func categoryOptions() []string {
	options := []string{"Alle Kategorien"}
	seen := map[string]bool{}
	for _, item := range catalog.Items() {
		parts := strings.Split(item.Category, " / ")
		for i := range parts {
			name := strings.Join(parts[:i+1], " / ")
			if !seen[name] {
				options = append(options, name)
				seen[name] = true
			}
		}
	}
	sort.Strings(options[1:])
	return options
}

func formatSilver(value int) string { return fmt.Sprintf("%d", value) }
func formatROI(profit, buy int) string {
	return fmt.Sprintf("%.1f%%", float64(profit)*100/float64(buy))
}
func formatAge(age time.Duration) string {
	if age == 0 {
		return "unbekannt"
	}
	if age < time.Minute {
		return "<1 Min."
	}
	return age.Truncate(time.Minute).String()
}
func containsFold(value, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}
