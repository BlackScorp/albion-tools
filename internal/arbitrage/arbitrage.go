// Package arbitrage calculates direct market opportunities from price quotes.
package arbitrage

import (
	"fmt"
	"sort"
	"time"

	"github.com/blackscorp/albion-helper/internal/catalog"
)

// Opportunity describes buying at the source's lowest sell order and selling
// into the destination's highest buy order. DataAge is based on the older of
// those two observations.
type Opportunity struct {
	ItemID     string
	Quality    catalog.Quality
	BuyMarket  catalog.Market
	SellMarket catalog.Market
	BuyPrice   int64
	SellPrice  int64
	Profit     int64
	ROI        float64
	Range      int
	DataAge    time.Duration
}

// Calculate returns every profitable directed market pair, grouped by item
// and quality. A timestamp of zero is treated as unknown and gives a zero age.
func Calculate(prices []catalog.Price, now time.Time) ([]Opportunity, error) {
	type key struct {
		item    string
		quality catalog.Quality
	}
	groups := make(map[key][]catalog.Price)
	for _, price := range prices {
		if price.ItemID == "" {
			return nil, fmt.Errorf("price has empty item ID")
		}
		if _, err := catalog.RingDistance(price.Market, price.Market); err != nil {
			return nil, err
		}
		groups[key{price.ItemID, price.Quality}] = append(groups[key{price.ItemID, price.Quality}], price)
	}

	var results []Opportunity
	for pair, quotes := range groups {
		for _, source := range quotes {
			if source.Buy <= 0 {
				continue
			}
			for _, destination := range quotes {
				if source.Market == destination.Market || destination.Sell <= source.Buy {
					continue
				}
				distance, err := catalog.RingDistance(source.Market, destination.Market)
				if err != nil {
					return nil, err
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
				results = append(results, Opportunity{
					ItemID: pair.item, Quality: pair.quality,
					BuyMarket: source.Market, SellMarket: destination.Market,
					BuyPrice: source.Buy, SellPrice: destination.Sell,
					Profit: profit, ROI: float64(profit) * 100 / float64(source.Buy),
					Range: distance, DataAge: age,
				})
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.ItemID != b.ItemID {
			return a.ItemID < b.ItemID
		}
		if a.Quality != b.Quality {
			return a.Quality < b.Quality
		}
		if a.BuyMarket != b.BuyMarket {
			return a.BuyMarket < b.BuyMarket
		}
		return a.SellMarket < b.SellMarket
	})
	return results, nil
}
