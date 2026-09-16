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

func TestRingDistanceBrecilienIsFive(t *testing.T) {
	for _, other := range []Market{Thetford, BlackMarket} {
		for _, pair := range [][2]Market{{Brecilien, other}, {other, Brecilien}} {
			got, err := RingDistance(pair[0], pair[1])
			if err != nil || got != 5 {
				t.Errorf("RingDistance(%q, %q) = %d, %v; want 5", pair[0], pair[1], got, err)
			}
		}
	}
}

func TestStartCatalogHasRepresentativeMetadata(t *testing.T) {
	got := Items()
	if len(got) != 12237 {
		t.Fatalf("catalog has %d items, want all 12,237 source IDs", len(got))
	}
	seen := make(map[string]bool)
	withRecipe := 0
	for _, item := range got {
		if item.ID == "" || item.Name == "" || item.Category == nil || item.Tier < 0 || item.Tier > 8 || item.Enchantment < 0 || item.Enchantment > 4 {
			t.Errorf("invalid catalog item: %+v", item)
		}
		if seen[item.ID] {
			t.Errorf("duplicate item ID %q", item.ID)
		}
		seen[item.ID] = true
		if item.Recipe != nil {
			withRecipe++
		}
	}
	if withRecipe < 9000 {
		t.Fatalf("only %d catalog variants have an XML recipe link", withRecipe)
	}
	got[0].Name = "mutated"
	if Items()[0].Name == "mutated" {
		t.Fatal("Items exposed mutable catalog storage")
	}
}

func TestGeneratedCategoryNodesHaveLinkedParents(t *testing.T) {
	categories := catalogCategories()
	if len(categories) != 1107 {
		t.Fatalf("category nodes = %d, want 1,107 from items.xml", len(categories))
	}
	wood := categories["crafting/resources/wood"]
	if wood == nil || wood.Parent == nil || wood.Parent.Path() != "Crafting / Resources" {
		t.Fatalf("wood category parent chain = %+v", wood)
	}
}

func TestEveryFamilyGetIDResolvesEachSourceVariantExactly(t *testing.T) {
	for _, definition := range definitions {
		seen := make(map[[2]int]string, len(definition.Variants))
		for _, variant := range definition.Variants {
			key := [2]int{variant.Tier, variant.Enchantment}
			if previous, exists := seen[key]; exists {
				t.Fatalf("family %q has ambiguous T%d.%d IDs %q and %q", definition.BaseID, key[0], key[1], previous, variant.ID)
			}
			seen[key] = variant.ID
			got, err := definition.GetID(variant.Tier, variant.Enchantment)
			if err != nil || got != variant.ID {
				t.Fatalf("%s.GetID(%d, %d) = %q, %v, want %q", definition.BaseID, variant.Tier, variant.Enchantment, got, err, variant.ID)
			}
		}
	}
}

func TestItemDefinitionGeneratesSupportedIDs(t *testing.T) {
	broadsword := definitionByBaseID(t, "MAIN_SWORD")
	tests := []struct {
		tier, enchantment int
		want              string
	}{
		{4, 0, "T4_MAIN_SWORD"},
		{4, 4, "T4_MAIN_SWORD@4"},
		{5, 0, "T5_MAIN_SWORD"},
		{8, 4, "T8_MAIN_SWORD@4"},
	}
	for _, tt := range tests {
		got, err := broadsword.GetID(tt.tier, tt.enchantment)
		if err != nil || got != tt.want {
			t.Errorf("GetID(%d, %d) = %q, %v; want %q", tt.tier, tt.enchantment, got, err, tt.want)
		}
	}
	for _, tt := range [][2]int{{9, 0}, {4, 5}} {
		if _, err := broadsword.GetID(tt[0], tt[1]); err == nil {
			t.Errorf("GetID(%d, %d) accepted unsupported variant", tt[0], tt[1])
		}
	}
}

func TestCategoryParentTree(t *testing.T) {
	if got := SwordsCategory.Path(); got != "Weapons / Sword / Swords" {
		t.Fatalf("category path = %q, want Weapons / Swords", got)
	}
	if !SwordsCategory.IsWithin(WeaponsCategory) || SwordsCategory.IsWithin(MountsCategory) {
		t.Fatal("category parent relationships are incorrect")
	}
}

func TestItemDefinitionsExpandOnlyListedVariantRanges(t *testing.T) {
	broadsword := definitionByBaseID(t, "MAIN_SWORD")
	cape := definitionByBaseID(t, "CAPE")
	wood := definitionByBaseID(t, "WOOD")
	horse := definitionByBaseID(t, "MOUNT_HORSE")
	for name, definition := range map[string]ItemDefinition{"broadsword": broadsword, "cape": cape, "wood": wood, "horse": horse} {
		if len(definition.variants()) == 0 {
			t.Fatalf("%s has no variants", name)
		}
	}
	if got, err := wood.GetID(4, 1); err != nil || got != "T4_WOOD_LEVEL1@1" {
		t.Fatalf("Wood.GetID(4, 1) = %q, %v; want T4_WOOD_LEVEL1@1", got, err)
	}
	if _, err := wood.GetID(3, 1); err == nil {
		t.Fatal("Wood.GetID accepted an enchantment below its supported tier")
	}
}

func TestCatalogItemsCarryTranslatedNamesCategoriesAndCraftingRecipes(t *testing.T) {
	sword := catalogItemByID(t, "T4_MAIN_SWORD")
	if sword.Name != "Breitschwert" || sword.FullName != "Breitschwert des Adepten" {
		t.Errorf("short/full sword name = %q / %q", sword.Name, sword.FullName)
	}
	if got, want := sword.Category.Path(), "Weapons / Sword"; got != want {
		t.Errorf("sword category = %q, want %q", got, want)
	}
	if sword.Recipe == nil || sword.Recipe.SourceItemID != "T4_MAIN_SWORD" {
		t.Fatalf("sword recipe is not linked to its XML item: %+v", sword.Recipe)
	}
	if sword.Recipe.Attributes["craftingfocus"] != "1286" {
		t.Errorf("sword recipe crafting focus = %q", sword.Recipe.Attributes["craftingfocus"])
	}
	resources := map[string]string{}
	for _, resource := range sword.Recipe.Resources {
		resources[resource.ItemID] = resource.Attributes["count"]
	}
	if resources["T4_METALBAR"] != "16" || resources["T4_LEATHER"] != "8" {
		t.Errorf("sword recipe resources = %#v", resources)
	}

	wood := catalogItemByID(t, "T4_WOOD_LEVEL1@1")
	if wood.Name != "Holz" || wood.FullName != "Ungewöhnliches Kiefernholz" {
		t.Errorf("short/full wood name = %q / %q", wood.Name, wood.FullName)
	}
	if wood.Tier != 4 || wood.Enchantment != 1 || wood.Recipe == nil || wood.Recipe.SourceItemID != "T4_WOOD_LEVEL1" {
		t.Errorf("enchanted wood metadata = %+v", wood)
	}
	swordEnchanted := catalogItemByID(t, "T4_MAIN_SWORD@4")
	if swordEnchanted.Enchantment != 4 {
		t.Errorf("T4_MAIN_SWORD@4 enchantment = %d, want 4", swordEnchanted.Enchantment)
	}
}

func TestCatalogCategoriesHideTechnicalRepeats(t *testing.T) {
	checks := map[string]string{
		"T5_CAPE":             "Capes / Standard",
		"T4_MAIN_MACE":        "Weapons / One Handed",
		"T4_SHOES_CLOTH_SET1": "Shoes / Set 1",
	}
	for id, want := range checks {
		item := catalogItemByID(t, id)
		if got := item.Category.Path(); got != want {
			t.Errorf("%s category = %q, want %q", id, got, want)
		}
	}
}

func catalogItemByID(t *testing.T, id string) Item {
	t.Helper()
	for _, item := range Items() {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("catalog item %q not found", id)
	return Item{}
}

func definitionByBaseID(t *testing.T, baseID string) ItemDefinition {
	t.Helper()
	for _, definition := range definitions {
		if definition.BaseID == baseID {
			return definition
		}
	}
	t.Fatalf("no item definition found for %s", baseID)
	return ItemDefinition{}
}
