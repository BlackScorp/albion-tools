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
	item, itemID   string
	opportunity    arbitrage.Opportunity
	hasOpportunity bool
}

var columns = []string{"Item", "Kaufstadt", "Verkaufsstadt", "Range", "Kaufpreis", "Verkaufspreis", "Bruttogewinn", "ROI (brutto)", "Datenalter"}

const maxItemsPerPage = 50

type syncFunc func(string, []string, []catalog.Market) ([]catalog.Price, error)

func NewWindow(a fyne.App) fyne.Window { return NewWindowWithData(a, nil, nil) }

// NewWindowWithData creates the market opportunity view from cached prices.
func NewWindowWithData(a fyne.App, cached []catalog.Price, syncPrices syncFunc) fyne.Window {
	w := a.NewWindow("Albion Helper")
	w.Resize(fyne.NewSize(1200, 720))

	server := widget.NewSelect([]string{"Europe", "Americas", "Asia"}, nil)
	server.SetSelected("Europe")
	search := widget.NewEntry()
	search.SetPlaceHolder("Name, ID oder Kategorie filtern …")
	selectedCategory := "Alle Kategorien"
	categoryButton := widget.NewButton(selectedCategory, nil)
	tier := widget.NewSelect([]string{"Alle Tiers", "1", "2", "3", "4", "5", "6", "7", "8"}, nil)
	tier.SetSelected("Alle Tiers")
	enchantment := widget.NewSelect([]string{"Alle Verzauberungen", "0", "1", "2", "3", "4"}, nil)
	enchantment.SetSelected("Alle Verzauberungen")
	rangeFilter := widget.NewSelect([]string{"Alle Ranges", "1", "2", "3", "4", "5"}, nil)
	rangeFilter.SetSelected("Alle Ranges")
	minProfit := widget.NewEntry()
	minProfit.SetPlaceHolder("Mindestgewinn")
	minROI := widget.NewEntry()
	minROI.SetPlaceHolder("Mindest-ROI %")
	cityChecks := make(map[catalog.Market]*widget.Check, len(catalog.Markets))
	cityControls := container.NewHBox()
	defaults := make(map[string]bool)
	for _, name := range defaultCitySelection() {
		defaults[name] = true
	}
	for _, market := range catalog.Markets {
		check := widget.NewCheck(string(market), nil)
		check.SetChecked(defaults[string(market)])
		cityChecks[market] = check
		cityControls.Add(check)
	}
	syncButton := widget.NewButton("Preise aktualisieren", nil)
	status := widget.NewLabel("Offline · Gebühren und Transportkosten nicht berücksichtigt")

	items := catalog.Items()
	itemByID := make(map[string]catalog.Item, len(items))
	for _, item := range items {
		itemByID[item.ID] = item
	}
	allPrices := append([]catalog.Price(nil), cached...)
	allRows := catalogRows(items, bestPriceRows(filterPricesByMarkets(allPrices, checkedCityMarkets(cityChecks)), itemByID, time.Now()))
	observationCount := len(allPrices)
	if observationCount > 0 {
		status.SetText(fmt.Sprintf("Offline · %d gespeicherte Marktbeobachtungen · %d Items im Katalog", observationCount, len(allRows)))
	}
	rows := append([]marketRow(nil), allRows...)
	filteredRows := append([]marketRow(nil), allRows...)
	currentPage, pageSize := 0, maxItemsPerPage
	ascending, sortedColumn := true, -1
	previousPage := widget.NewButton("‹ Zurück", nil)
	nextPage := widget.NewButton("Weiter ›", nil)
	pageLabel := widget.NewLabel("")
	table := widget.NewTable(
		func() (int, int) { return len(rows) + 1, len(columns) },
		func() fyne.CanvasObject { return widget.NewLabel("#######") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			label.Truncation = fyne.TextTruncateEllipsis
			label.Wrapping = fyne.TextWrapOff
			if id.Row == 0 {
				label.SetText(columns[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			r := rows[id.Row-1]
			if !r.hasOpportunity {
				values := []string{r.item, "–", "–", "–", "–", "–", "–", "–", "–"}
				label.SetText(values[id.Col])
				return
			}
			op := r.opportunity
			rangeLabel := fmt.Sprint(op.Range)
			if op.Range < 0 {
				rangeLabel = "–"
			}
			values := []string{r.item, string(op.BuyMarket), string(op.SellMarket), rangeLabel, formatSilver(int(op.BuyPrice)), formatSilver(int(op.SellPrice)), formatSilver(int(op.Profit)), formatROI(int(op.Profit), int(op.BuyPrice)), formatAge(op.DataAge)}
			label.SetText(values[id.Col])
		},
	)
	for col, width := range []float32{220, 130, 140, 65, 110, 120, 125, 105, 100} {
		table.SetColumnWidth(col, width)
	}
	var refreshPage func()
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row != 0 {
			return
		}
		if id.Col == 3 {
			table.UnselectAll()
			return
		}
		sortedColumn, ascending = nextSortState(id.Col, sortedColumn, ascending)
		sortRows(filteredRows, id.Col, ascending)
		table.UnselectAll()
		refreshPage()
		table.Refresh()
	}
	refreshPage = func() {
		pageCount := (len(filteredRows) + pageSize - 1) / pageSize
		if pageCount == 0 {
			currentPage = 0
			rows = nil
			pageLabel.SetText("Keine Items")
			previousPage.Disable()
			nextPage.Disable()
			table.Refresh()
			return
		}
		if currentPage >= pageCount {
			currentPage = pageCount - 1
		}
		start := currentPage * pageSize
		end := start + pageSize
		if end > len(filteredRows) {
			end = len(filteredRows)
		}
		rows = pageRows(filteredRows, currentPage, pageSize)
		pageLabel.SetText(fmt.Sprintf("Seite %d/%d · Items %d–%d von %d · maximal %d je Sync", currentPage+1, pageCount, start+1, end, len(filteredRows), pageSize))
		if currentPage == 0 {
			previousPage.Disable()
		} else {
			previousPage.Enable()
		}
		if currentPage+1 == pageCount {
			nextPage.Disable()
		} else {
			nextPage.Enable()
		}
		table.Refresh()
	}
	previousPage.OnTapped = func() {
		if currentPage > 0 {
			currentPage--
			refreshPage()
		}
	}
	nextPage.OnTapped = func() {
		if (currentPage+1)*pageSize < len(filteredRows) {
			currentPage++
			refreshPage()
		}
	}
	applyFilters := func() {
		criteria, err := readCriteria(search.Text, selectedCategory, tier.Selected, enchantment.Selected, rangeFilter.Selected, minProfit.Text, minROI.Text)
		if err != nil {
			status.SetText("Filterfehler: " + err.Error())
			return
		}
		filteredRows = filterRows(allRows, itemByID, criteria)
		currentPage = 0
		refreshPage()
		if observationCount == 0 {
			status.SetText(fmt.Sprintf("Keine gespeicherten Marktbeobachtungen · Offline · %d Items passen", len(filteredRows)))
		} else if len(filteredRows) == 0 {
			status.SetText(fmt.Sprintf("%d gespeicherte Marktbeobachtungen · Keine Items passen", observationCount))
		} else {
			status.SetText(fmt.Sprintf("%d gespeicherte Marktbeobachtungen · %d Items passen · Gebühren und Transportkosten nicht berücksichtigt", observationCount, len(filteredRows)))
		}
		table.Refresh()
	}
	categoryButton.OnTapped = func() {
		menu := categoryMenu(func(category string) {
			selectedCategory = category
			categoryButton.SetText(category)
			applyFilters()
		})
		widget.ShowPopUpMenuAtRelativePosition(menu, w.Canvas(), fyne.NewPos(0, categoryButton.Size().Height), categoryButton)
	}
	search.OnChanged = func(string) { applyFilters() }
	tier.OnChanged = func(string) { applyFilters() }
	enchantment.OnChanged = func(string) { applyFilters() }
	rangeFilter.OnChanged = func(string) { applyFilters() }
	minProfit.OnChanged = func(string) { applyFilters() }
	minROI.OnChanged = func(string) { applyFilters() }

	var syncMu syncpkg.Mutex
	if syncPrices != nil {
		syncButton.OnTapped = func() {
			if !syncMu.TryLock() {
				return
			}
			criteria, err := readCriteria(search.Text, selectedCategory, tier.Selected, enchantment.Selected, rangeFilter.Selected, minProfit.Text, minROI.Text)
			if err != nil {
				syncMu.Unlock()
				status.SetText("Filterfehler: " + err.Error())
				return
			}
			if !hasItemScope(criteria) {
				syncMu.Unlock()
				status.SetText("Bitte zuerst eine Kategorie, ein Item, Tier oder eine Verzauberung filtern")
				return
			}
			selectedMarkets := checkedCityMarkets(cityChecks)
			if len(selectedMarkets) == 0 {
				syncMu.Unlock()
				status.SetText("Bitte mindestens eine Stadt auswählen")
				return
			}
			ids := filteredItemIDs(items, criteria)
			if len(ids) == 0 {
				syncMu.Unlock()
				status.SetText("Keine Items passen zu den Filtern")
				return
			}
			selectedServer := server.Selected
			syncButton.Disable()
			status.SetText(fmt.Sprintf("Synchronisierung läuft · %d Items in %d Städten …", len(ids), len(selectedMarkets)))
			go func() {
				prices, err := syncPrices(selectedServer, ids, selectedMarkets)
				fyne.Do(func() {
					defer syncMu.Unlock()
					syncButton.Enable()
					if err != nil {
						status.SetText("Synchronisierung fehlgeschlagen: " + err.Error())
						return
					}
					allPrices = mergePrices(allPrices, prices)
					observationCount = len(allPrices)
					allRows = catalogRows(items, bestPriceRows(filterPricesByMarkets(allPrices, selectedMarkets), itemByID, time.Now()))
					applyFilters()
					status.SetText(fmt.Sprintf("Synchronisierung %s abgeschlossen · %d gefilterte Items aktualisiert · insgesamt %d gespeicherte Marktbeobachtungen · Gebühren und Transportkosten nicht berücksichtigt", time.Now().Format("15:04:05"), len(ids), observationCount))
				})
			}()
		}
	}

	for _, check := range cityChecks {
		check.OnChanged = func(bool) {
			selected := checkedCityMarkets(cityChecks)
			allRows = catalogRows(items, bestPriceRows(filterPricesByMarkets(allPrices, selected), itemByID, time.Now()))
			applyFilters()
		}
	}
	toolbar := container.NewHBox(widget.NewLabel("Server:"), server)
	filters := container.NewHBox(widget.NewLabel("Kategorie"), categoryButton, tier, enchantment, rangeFilter)
	thresholds := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, widget.NewLabel("Mindestgewinn:"), nil, minProfit),
		container.NewBorder(nil, nil, widget.NewLabel("Mindest-ROI %:"), nil, minROI),
	)
	cityRow := container.NewHBox(widget.NewLabel("Städte:"), cityControls)
	filterPanel := container.NewVBox(toolbar, search, filters, thresholds, cityRow, syncButton, status)
	pagination := container.NewHBox(previousPage, pageLabel, nextPage)
	refreshPage()
	w.SetContent(container.NewBorder(filterPanel, pagination, nil, nil, table))
	return w
}

type criteria struct {
	query, category   string
	tier, enchantment int
	rangeValue        int
	minProfit         int64
	minROI            float64
}

func readCriteria(query, category, tier, enchantment, rangeFilter, profit, roi string) (criteria, error) {
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
	if rangeFilter != "" && rangeFilter != "Alle Ranges" {
		parsed, err := strconv.Atoi(rangeFilter)
		if err != nil || parsed < 1 || parsed > 5 {
			return c, fmt.Errorf("Range muss zwischen 1 und 5 liegen")
		}
		c.rangeValue = parsed
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
	if c.query != "" && !containsFold(item.Name+" "+item.FullName+" "+item.ID+" "+item.Category.Path(), c.query) {
		return false
	}
	if c.category != "" && c.category != "Alle Kategorien" && !item.Category.IncludesPath(c.category) {
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

func hasItemScope(c criteria) bool {
	return c.query != "" || (c.category != "" && c.category != "Alle Kategorien") || c.tier > 0 || c.enchantment >= 0
}

func filteredItemIDs(items []catalog.Item, c criteria) []string {
	ids := make([]string, 0)
	for _, item := range items {
		if itemMatches(item, c) {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func selectedCityMarkets(selected []string) []catalog.Market {
	chosen := make(map[string]bool, len(selected))
	for _, name := range selected {
		chosen[name] = true
	}
	markets := make([]catalog.Market, 0, len(selected))
	for _, market := range catalog.Markets {
		if chosen[string(market)] {
			markets = append(markets, market)
		}
	}
	return markets
}

func defaultCitySelection() []string {
	return []string{string(catalog.Thetford), string(catalog.FortSterling), string(catalog.Lymhurst), string(catalog.Bridgewatch), string(catalog.Martlock)}
}

func checkedCityMarkets(checks map[catalog.Market]*widget.Check) []catalog.Market {
	markets := make([]catalog.Market, 0, len(checks))
	for _, market := range catalog.Markets {
		if check := checks[market]; check != nil && check.Checked {
			markets = append(markets, market)
		}
	}
	return markets
}

func filterPricesByMarkets(prices []catalog.Price, markets []catalog.Market) []catalog.Price {
	selected := make(map[catalog.Market]bool, len(markets))
	for _, market := range markets {
		selected[market] = true
	}
	filtered := make([]catalog.Price, 0, len(prices))
	for _, price := range prices {
		if selected[price.Market] {
			filtered = append(filtered, price)
		}
	}
	return filtered
}

func itemTitle(item catalog.Item) string {
	if item.Tier <= 0 {
		return item.Name
	}
	return fmt.Sprintf("%s T%d.%d", item.Name, item.Tier, item.Enchantment)
}

func filterRows(rows []marketRow, items map[string]catalog.Item, c criteria) []marketRow {
	filtered := make([]marketRow, 0, len(rows))
	for _, row := range rows {
		item, ok := items[row.itemID]
		if !ok || !itemMatches(item, c) {
			continue
		}
		if c.rangeValue > 0 && (!row.hasOpportunity || row.opportunity.Range != c.rangeValue) {
			continue
		}
		if (c.minProfit > 0 || c.minROI > 0) && !row.hasOpportunity {
			continue
		}
		if row.hasOpportunity && (row.opportunity.Profit < c.minProfit || row.opportunity.ROI < c.minROI) {
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
		name := itemTitle(item)
		rows = append(rows, marketRow{item: name, itemID: item.ID, opportunity: op, hasOpportunity: true})
	}
	return rows
}

func bestPriceRows(prices []catalog.Price, items map[string]catalog.Item, now time.Time) []marketRow {
	best := make(map[string]marketRow)
	type priceKey struct {
		item    string
		quality catalog.Quality
	}
	groups := make(map[priceKey][]catalog.Price)
	for _, price := range prices {
		key := priceKey{item: price.ItemID, quality: price.Quality}
		groups[key] = append(groups[key], price)
	}
	for key, quotes := range groups {
		for _, source := range quotes {
			if source.Buy <= 0 {
				continue
			}
			for _, destination := range quotes {
				if source.Market == destination.Market || destination.Sell <= 0 {
					continue
				}
				item, exists := items[key.item]
				if !exists {
					continue
				}
				distance, err := catalog.RingDistance(source.Market, destination.Market)
				if err != nil {
					continue
				}
				oldest := source.UpdatedAt
				if oldest == 0 || destination.UpdatedAt == 0 {
					oldest = 0
				} else if destination.UpdatedAt < oldest {
					oldest = destination.UpdatedAt
				}
				var age time.Duration
				if oldest != 0 {
					age = now.Sub(time.Unix(oldest, 0))
					if age < 0 {
						age = 0
					}
				}
				profit := destination.Sell - source.Buy
				row := marketRow{
					item:   itemTitle(item),
					itemID: item.ID, hasOpportunity: true,
					opportunity: arbitrage.Opportunity{
						ItemID: item.ID, Quality: key.quality, BuyMarket: source.Market, SellMarket: destination.Market,
						BuyPrice: source.Buy, SellPrice: destination.Sell, Profit: profit,
						ROI: float64(profit) * 100 / float64(source.Buy), Range: distance, DataAge: age,
					},
				}
				if previous, exists := best[item.ID]; exists && previous.opportunity.Profit >= profit {
					continue
				}
				best[item.ID] = row
			}
		}
	}
	rows := make([]marketRow, 0, len(best))
	for _, row := range best {
		rows = append(rows, row)
	}
	return rows
}

func catalogRows(items []catalog.Item, opportunities []marketRow) []marketRow {
	rows := make([]marketRow, 0, len(items))
	byID := make(map[string]int, len(items))
	baseNames := make(map[string]string, len(items))
	for _, item := range items {
		name := itemTitle(item)
		byID[item.ID] = len(rows)
		baseNames[item.ID] = name
		rows = append(rows, marketRow{item: name, itemID: item.ID})
	}
	for _, opportunity := range opportunities {
		index, exists := byID[opportunity.itemID]
		if !exists || (rows[index].hasOpportunity && rows[index].opportunity.Profit >= opportunity.opportunity.Profit) {
			continue
		}
		rows[index].opportunity = opportunity.opportunity
		rows[index].hasOpportunity = true
		rows[index].item = baseNames[opportunity.itemID]
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].item == rows[j].item {
			return rows[i].itemID < rows[j].itemID
		}
		return rows[i].item < rows[j].item
	})
	return rows
}

func pageRows(rows []marketRow, page, pageSize int) []marketRow {
	if page < 0 || pageSize <= 0 || page*pageSize >= len(rows) {
		return nil
	}
	start := page * pageSize
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return append([]marketRow(nil), rows[start:end]...)
}

func mergePrices(existing, updated []catalog.Price) []catalog.Price {
	merged := make(map[string]catalog.Price, len(existing)+len(updated))
	key := func(price catalog.Price) string {
		return fmt.Sprintf("%s\x00%s\x00%d", price.ItemID, price.Market, price.Quality)
	}
	for _, price := range existing {
		merged[key(price)] = price
	}
	for _, price := range updated {
		merged[key(price)] = price
	}
	result := make([]catalog.Price, 0, len(merged))
	for _, price := range merged {
		result = append(result, price)
	}
	return result
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
		if rows[i].hasOpportunity != rows[j].hasOpportunity {
			return rows[i].hasOpportunity
		}
		if ascending {
			return lessValue(rows[i], rows[j])
		}
		return lessValue(rows[j], rows[i])
	})
}

func nextSortState(column, sortedColumn int, ascending bool) (int, bool) {
	if column == sortedColumn {
		return sortedColumn, !ascending
	}
	return column, true
}

type categoryNode struct {
	children map[string]*categoryNode
}

func categoryMenu(selectCategory func(string)) *fyne.Menu {
	root := &categoryNode{children: map[string]*categoryNode{}}
	for _, item := range catalog.Items() {
		parts := strings.Split(item.Category.Path(), " / ")
		node := root
		for _, part := range parts {
			if node.children[part] == nil {
				node.children[part] = &categoryNode{children: map[string]*categoryNode{}}
			}
			node = node.children[part]
		}
	}
	menu := fyne.NewMenu("")
	menu.Items = append(menu.Items, fyne.NewMenuItem("Alle Kategorien", func() { selectCategory("Alle Kategorien") }))
	for _, name := range sortedCategoryNames(root.children) {
		menu.Items = append(menu.Items, categoryMenuItem(name, name, root.children[name], selectCategory))
	}
	return menu
}

func categoryMenuItem(name, path string, node *categoryNode, selectCategory func(string)) *fyne.MenuItem {
	if len(node.children) == 0 {
		return fyne.NewMenuItem(name, func() { selectCategory(path) })
	}
	children := fyne.NewMenu("")
	children.Items = append(children.Items, fyne.NewMenuItem(name+" (alle)", func() { selectCategory(path) }))
	for _, child := range sortedCategoryNames(node.children) {
		childPath := path + " / " + child
		children.Items = append(children.Items, categoryMenuItem(child, childPath, node.children[child], selectCategory))
	}
	return &fyne.MenuItem{Label: name, ChildMenu: children}
}

func sortedCategoryNames(children map[string]*categoryNode) []string {
	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
