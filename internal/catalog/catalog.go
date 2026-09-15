// Package catalog contains the bundled item definitions and shared market types.
package catalog

import "fmt"

type Market string

const (
	Thetford     Market = "Thetford"
	FortSterling Market = "Fort Sterling"
	Lymhurst     Market = "Lymhurst"
	Bridgewatch  Market = "Bridgewatch"
	Martlock     Market = "Martlock"
	Caerleon     Market = "Caerleon"
	BlackMarket  Market = "Black Market"
)

var Markets = []Market{Thetford, FortSterling, Lymhurst, Bridgewatch, Martlock, Caerleon, BlackMarket}

// Item describes one concrete catalog entry. Enchantment is zero for a base item.
type Item struct {
	ID          string
	Name        string
	Category    string
	Tier        int
	Enchantment int
}

// Quality is the market quality attached to a price observation.
type Quality int

// Price stores one market's buy/sell quotes and their observation time.
type Price struct {
	ItemID    string
	Market    Market
	Quality   Quality
	Buy       int64
	Sell      int64
	UpdatedAt int64
}

var items = []Item{
	{ID: "T4_MAIN_SWORD", Name: "Broadsword", Category: "Weapons / Swords", Tier: 4},
	{ID: "T4_CAPE", Name: "Cape", Category: "Accessories / Capes", Tier: 4},
	{ID: "T4_WOOD", Name: "Wood", Category: "Resources / Logs", Tier: 4},
	{ID: "T4_MOUNT_HORSE", Name: "Riding Horse", Category: "Mounts / Horses", Tier: 4},
	{ID: "T4_MAIN_SWORD@1", Name: "Broadsword", Category: "Weapons / Swords", Tier: 4, Enchantment: 1},
	{ID: "T5_MAIN_SWORD", Name: "Broadsword", Category: "Weapons / Swords", Tier: 5},
}

// Items returns a copy so callers cannot mutate the bundled catalog.
func Items() []Item { return append([]Item(nil), items...) }

// RingDistance returns the shortest leg count between Royal Cities. Any
// connection involving Caerleon or the Black Market is represented as one leg;
// those two markets remain distinct even though they share a location.
func RingDistance(from, to Market) (int, error) {
	if !validMarket(from) || !validMarket(to) {
		return 0, fmt.Errorf("unknown market: %q to %q", from, to)
	}
	if from == to {
		return 0, nil
	}
	if isSpecial(from) || isSpecial(to) {
		return 1, nil
	}
	order := []Market{Thetford, FortSterling, Lymhurst, Bridgewatch, Martlock}
	fi, ti := -1, -1
	for i, market := range order {
		if market == from {
			fi = i
		}
		if market == to {
			ti = i
		}
	}
	distance := fi - ti
	if distance < 0 {
		distance = -distance
	}
	if distance > len(order)/2 {
		distance = len(order) - distance
	}
	return distance, nil
}

func validMarket(m Market) bool {
	for _, known := range Markets {
		if known == m {
			return true
		}
	}
	return false
}

func isSpecial(m Market) bool { return m == Caerleon || m == BlackMarket }
