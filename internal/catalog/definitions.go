package catalog

var definitions = func() []ItemDefinition {
	var all []ItemDefinition
	all = append(all, itemFamilies_armors...)
	all = append(all, itemFamilies_artefacts...)
	all = append(all, itemFamilies_bags...)
	all = append(all, itemFamilies_capes...)
	all = append(all, itemFamilies_consumables...)
	all = append(all, itemFamilies_crafting...)
	all = append(all, itemFamilies_farming...)
	all = append(all, itemFamilies_furniture...)
	all = append(all, itemFamilies_gathering...)
	all = append(all, itemFamilies_head...)
	all = append(all, itemFamilies_mounts...)
	all = append(all, itemFamilies_offhands...)
	all = append(all, itemFamilies_other...)
	all = append(all, itemFamilies_shoes...)
	all = append(all, itemFamilies_uncategorized...)
	all = append(all, itemFamilies_vanity...)
	all = append(all, itemFamilies_weapons...)
	return all
}()
