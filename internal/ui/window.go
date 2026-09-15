package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type marketRow struct {
	item, buyCity, sellCity string
	buy, sell               int
}

var sampleRows = []marketRow{
	{item: "Adept's Broadsword", buyCity: "Martlock", sellCity: "Bridgewatch", buy: 1240, sell: 1590},
	{item: "Adept's Cape", buyCity: "Lymhurst", sellCity: "Thetford", buy: 8700, sell: 9450},
	{item: "Elder's Wood", buyCity: "Fort Sterling", sellCity: "Caerleon", buy: 410, sell: 525},
	{item: "Expert's Riding Horse", buyCity: "Thetford", sellCity: "Martlock", buy: 18200, sell: 19750},
}

var columns = []string{"Item", "Kaufstadt", "Verkaufsstadt", "Kaufpreis", "Verkaufspreis", "Bruttogewinn", "ROI (brutto)"}

// NewWindow constructs the initial application window and its sample results.
func NewWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("Albion Helper")
	w.Resize(fyne.NewSize(1100, 650))

	server := widget.NewSelect([]string{"Europe", "Americas", "Asia"}, nil)
	server.SetSelected("Europe")
	filter := widget.NewEntry()
	filter.SetPlaceHolder("Item filtern …")
	sync := widget.NewButton("Preise aktualisieren", nil)
	status := widget.NewLabel("Beispieldaten · Gebühren und Transportkosten nicht berücksichtigt")

	rows := append([]marketRow(nil), sampleRows...)
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
		for _, row := range sampleRows {
			if query == "" || containsFold(row.item, query) {
				rows = append(rows, row)
			}
		}
		table.Refresh()
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
