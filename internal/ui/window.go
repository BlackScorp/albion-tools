package ui

import (
	"fmt"
	"sort"
	"strings"
	syncpkg "sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/blackscorp/albion-helper/internal/catalog"
)

type marketRow struct {
	item, itemID, buyCity, sellCity string
	buy, sell                       int
}

var columns = []string{"Item", "Kaufstadt", "Verkaufsstadt", "Kaufpreis", "Verkaufspreis", "Bruttogewinn", "ROI (brutto)"}

// NewWindow constructs the initial application window and its sample results.
type syncFunc func(string, []string) ([]catalog.Price, error)

func NewWindow(a fyne.App) fyne.Window { return NewWindowWithData(a, nil, nil) }

// NewWindowWithData creates the application window with cached prices and an optional sync action.
func NewWindowWithData(a fyne.App, cached []catalog.Price, syncPrices syncFunc) fyne.Window {
	w := a.NewWindow("Albion Helper")
	w.Resize(fyne.NewSize(1100, 650))

	server := widget.NewSelect([]string{"Europe", "Americas", "Asia"}, nil)
	server.SetSelected("Europe")
	filter := widget.NewEntry()
	filter.SetPlaceHolder("Item filtern …")
	sync := widget.NewButton("Preise aktualisieren", nil)
	status := widget.NewLabel("Offline · Gebühren und Transportkosten nicht berücksichtigt")

	catalogItems := catalog.Items()
	allRows := make([]marketRow, 0, len(catalogItems))
	itemByID := make(map[string]catalog.Item, len(catalogItems))
	for _, item := range catalogItems {
		itemByID[item.ID] = item
	}
	for _, price := range cached {
		item, ok := itemByID[price.ItemID]
		if !ok {
			continue
		}
		name := fmt.Sprintf("%s · T%d.%d · Q%d", item.Name, item.Tier, item.Enchantment, price.Quality)
		allRows = append(allRows, marketRow{item: name, itemID: item.ID, buyCity: string(price.Market), sellCity: string(price.Market), buy: int(price.Buy), sell: int(price.Sell)})
	}
	rows := append([]marketRow(nil), allRows...)
	ascending := true
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
			profit := r.sell - r.buy
			values := []string{r.item, r.buyCity, r.sellCity, formatSilver(r.buy), formatSilver(r.sell), formatSilver(profit), formatROI(profit, r.buy)}
			label.SetText(values[id.Col])
		},
	)
	for col := range columns {
		table.SetColumnWidth(col, []float32{230, 135, 145, 125, 140, 145, 130}[col])
	}
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row != 0 {
			return
		}
		ascending = !ascending
		col := id.Col
		sort.Slice(rows, func(i, j int) bool {
			left, right := rows[i], rows[j]
			less := false
			switch col {
			case 0:
				less = left.item < right.item
			case 1:
				less = left.buyCity < right.buyCity
			case 2:
				less = left.sellCity < right.sellCity
			case 3:
				less = left.buy < right.buy
			case 4:
				less = left.sell < right.sell
			case 5:
				less = left.sell-left.buy < right.sell-right.buy
			case 6:
				less = (left.sell-left.buy)*100/left.buy < (right.sell-right.buy)*100/right.buy
			}
			if left == right {
				return false
			}
			return less == ascending
		})
		table.Refresh()
	}
	filter.OnChanged = func(query string) {
		rows = rows[:0]
		for _, row := range allRows {
			if query == "" || containsFold(row.item, query) {
				rows = append(rows, row)
			}
		}
		table.Refresh()
	}
	var syncMu syncpkg.Mutex
	if syncPrices != nil {
		sync.OnTapped = func() {
			if !syncMu.TryLock() {
				return
			}
			ids := make([]string, 0, len(rows))
			seen := map[string]bool{}
			for _, row := range rows {
				if !seen[row.itemID] {
					seen[row.itemID] = true
					ids = append(ids, row.itemID)
				}
			}
			if len(ids) == 0 {
				for _, item := range catalogItems {
					if filter.Text == "" || containsFold(item.Name, filter.Text) || containsFold(item.ID, filter.Text) {
						ids = append(ids, item.ID)
					}
				}
			}
			selectedServer := server.Selected
			sync.Disable()
			status.SetText("Synchronisierung läuft …")
			go func() {
				prices, err := syncPrices(selectedServer, ids)
				fyne.Do(func() {
					defer syncMu.Unlock()
					sync.Enable()
					if err != nil {
						status.SetText("Synchronisierung fehlgeschlagen: " + err.Error())
						return
					}
					allRows = allRows[:0]
					for _, price := range prices {
						item, ok := itemByID[price.ItemID]
						if !ok {
							continue
						}
						name := fmt.Sprintf("%s · T%d.%d · Q%d", item.Name, item.Tier, item.Enchantment, price.Quality)
						allRows = append(allRows, marketRow{item: name, itemID: item.ID, buyCity: string(price.Market), sellCity: string(price.Market), buy: int(price.Buy), sell: int(price.Sell)})
					}
					rows = rows[:0]
					for _, row := range allRows {
						if filter.Text == "" || containsFold(row.item, filter.Text) {
							rows = append(rows, row)
						}
					}
					status.SetText(fmt.Sprintf("Synchronisierung %s abgeschlossen · %d Preisbeobachtungen", time.Now().Format("15:04:05"), len(prices)))
					table.Refresh()
				})
			}()
		}
	}

	toolbar := container.NewBorder(nil, nil, widget.NewLabel("Server:"), sync,
		container.NewVBox(server, filter))
	w.SetContent(container.NewBorder(container.NewVBox(toolbar, status), nil, nil, nil, table))
	return w
}

func formatSilver(value int) string {
	return fmt.Sprintf("%d", value)
}

func formatROI(profit, buy int) string {
	return fmt.Sprintf("%.1f%%", float64(profit)*100/float64(buy))
}

func containsFold(value, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}
