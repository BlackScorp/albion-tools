// Package catalog contains the Go item definitions and shared market types.
package catalog

import (
	"fmt"
	"sync"
)

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

// Category is one node in the parent-linked item category tree.
type Category struct {
	ID     string
	Name   string
	Parent *Category
}

// Path returns the category and all its parents, from root to leaf.
func (c *Category) Path() string {
	if c == nil {
		return ""
	}
	if c.Parent == nil {
		return c.Name
	}
	return c.Parent.Path() + " / " + c.Name
}

// IsWithin reports whether c is the requested category or one of its children.
func (c *Category) IsWithin(parent *Category) bool {
	for current := c; current != nil; current = current.Parent {
		if current == parent {
			return true
		}
	}
	return false
}

// IncludesPath reports whether the category is at or below the given path.
func (c *Category) IncludesPath(path string) bool {
	for current := c; current != nil; current = current.Parent {
		if current.Path() == path {
			return true
		}
	}
	return false
}

// Item describes one concrete tier/enchantment variant in the catalog.
type Item struct {
	ID          string
	Name        string
	Category    *Category
	Tier        int
	Enchantment int
	Recipe      *CraftingRecipe
}

// CraftingRecipe preserves the recipe metadata supplied by items.xml.
type CraftingRecipe struct {
	SourceItemID string
	Attributes   map[string]string
	Resources    []CraftingResource
}

// CraftingResource is one source item and its recipe-specific metadata.
type CraftingResource struct {
	ItemID     string
	Attributes map[string]string
}

// ItemVariant describes one concrete, valid ID in a family.
type ItemVariant struct {
	ID           string
	Name         string
	Tier         int
	Enchantment  int
	CategoryPath string
	Recipe       *CraftingRecipe
}

// ItemDefinition defines valid concrete variants for one family.
type ItemDefinition struct {
	Name           string
	BaseID         string
	MinTier        int
	MaxTier        int
	MinEnchantment int
	MaxEnchantment int
	Variants       []ItemVariant
}

// GetID returns the source-listed Albion ID for a supported variant.
func (d ItemDefinition) GetID(tier, enchantment int) (string, error) {
	for _, variant := range d.Variants {
		if variant.Tier == tier && variant.Enchantment == enchantment {
			return variant.ID, nil
		}
	}
	return "", fmt.Errorf("tier %d enchantment %d is not defined for %s", tier, enchantment, d.Name)
}

func (d ItemDefinition) variants() []Item {
	items := make([]Item, 0, len(d.Variants))
	categories := catalogCategories()
	for _, variant := range d.Variants {
		items = append(items, Item{
			ID: variant.ID, Name: variant.Name, Category: categories[variant.CategoryPath],
			Tier: variant.Tier, Enchantment: variant.Enchantment, Recipe: variant.Recipe,
		})
	}
	return items
}

type categorySpec struct{ Path, ID, Name, ParentPath string }

var categoriesOnce sync.Once
var categoriesByPath map[string]*Category

func catalogCategories() map[string]*Category {
	categoriesOnce.Do(func() {
		categoriesByPath = make(map[string]*Category, len(categorySpecs))
		for _, spec := range categorySpecs {
			category := &Category{ID: spec.ID, Name: spec.Name, Parent: categoriesByPath[spec.ParentPath]}
			categoriesByPath[spec.Path] = category
		}
	})
	return categoriesByPath
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

// Items expands the bundled Go definitions into concrete API item IDs.
func Items() []Item {
	items := make([]Item, 0)
	for _, definition := range definitions {
		items = append(items, definition.variants()...)
	}
	return items
}

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
