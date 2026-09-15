package catalog

import "testing"

func TestRingDistanceRoyalCityRing(t *testing.T) {
	tests := []struct {
		from, to Market
		want     int
	}{
		{Thetford, Thetford, 0},
		{Thetford, FortSterling, 1},
		{Thetford, Martlock, 1},
		{Thetford, Lymhurst, 2},
		{FortSterling, Bridgewatch, 2},
	}
	for _, tt := range tests {
		got, err := RingDistance(tt.from, tt.to)
		if err != nil || got != tt.want {
			t.Errorf("RingDistance(%q, %q) = %d, %v; want %d", tt.from, tt.to, got, err, tt.want)
		}
	}
}

func TestRingDistanceSpecialMarkets(t *testing.T) {
	for _, pair := range [][2]Market{{Caerleon, BlackMarket}, {Caerleon, Thetford}, {BlackMarket, Martlock}} {
		got, err := RingDistance(pair[0], pair[1])
		if err != nil || got != 1 {
			t.Errorf("RingDistance(%q, %q) = %d, %v; want 1", pair[0], pair[1], got, err)
		}
	}
}

func TestStartCatalogHasRepresentativeMetadata(t *testing.T) {
	got := Items()
	if len(got) < 5 {
		t.Fatalf("catalog has %d items, want representative entries", len(got))
	}
	seen := make(map[string]bool)
	for _, item := range got {
		if item.ID == "" || item.Name == "" || item.Category == "" || item.Tier < 1 || item.Tier > 8 || item.Enchantment < 0 || item.Enchantment > 4 {
			t.Errorf("invalid catalog item: %+v", item)
		}
		if seen[item.ID] {
			t.Errorf("duplicate item ID %q", item.ID)
		}
		seen[item.ID] = true
	}
	got[0].Name = "mutated"
	if Items()[0].Name == "mutated" {
		t.Fatal("Items exposed mutable catalog storage")
	}
}
