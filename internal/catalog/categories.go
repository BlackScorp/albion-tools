package catalog

// Category nodes and IDs follow the shopcategory hierarchy in items.xml.
var (
	WeaponsCategory   = &Category{ID: "weapons", Name: "Weapons"}
	SwordTypeCategory = &Category{ID: "sword", Name: "Sword", Parent: WeaponsCategory}
	SwordsCategory    = &Category{ID: "sword_sword", Name: "Swords", Parent: SwordTypeCategory}
	CapesCategory     = &Category{ID: "capes", Name: "Capes"}
	CapeTypeCategory  = &Category{ID: "accessoires_capes_capes", Name: "Standard Capes", Parent: CapesCategory}
	CraftingCategory  = &Category{ID: "crafting", Name: "Crafting"}
	ResourcesCategory = &Category{ID: "resources", Name: "Resources", Parent: CraftingCategory}
	WoodCategory      = &Category{ID: "wood", Name: "Wood", Parent: ResourcesCategory}
	MountsCategory    = &Category{ID: "mounts", Name: "Mounts"}
	BaseMountCategory = &Category{ID: "basemounts", Name: "Base Mounts", Parent: MountsCategory}
	HorsesCategory    = &Category{ID: "horse", Name: "Horses", Parent: BaseMountCategory}
)
