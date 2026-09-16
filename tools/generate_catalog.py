#!/usr/bin/env python3
"""Generate the bundled Go catalog from the local Albion metadata dumps."""

import collections
import hashlib
import json
import re
from pathlib import Path
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "internal" / "catalog"


def go_string(value):
    return json.dumps(value, ensure_ascii=False)


def slug(value):
    value = re.sub(r"[^a-zA-Z0-9]+", "_", value).strip("_").lower()
    return value or "empty"


def readable(value):
    words = re.sub(r"([a-z])([A-Z])", r"\1 \2", value).replace("_", " ").split()
    return " ".join(word[:1].upper() + word[1:].lower() for word in words)


def short_name(value):
    value = re.sub(r"^(?:Ungewöhnliches|Seltenes|Hervorragendes|Meisterhaftes|Uncommon|Rare|Exceptional|Masterpiece)\s+", "", value, flags=re.I)
    value = re.sub(r"\s+(?:des|der|of|the)\s+[^ ]+(?:\s+[^ ]+)?$", "", value, flags=re.I)
    return value.strip()


def family_name(family_id, variants):
    generic_resources = {"WOOD": "Holz", "FIBER": "Faser", "PLANKS": "Bretter"}
    if family_id in generic_resources:
        return generic_resources[family_id]
    names = [short_name(entry[4]) for entry in variants]
    return next((name for name in names if name), "Item")


def category_name(path_parts):
    ident = path_parts[-1]
    parent_id = path_parts[-2] if len(path_parts) > 1 else ""
    if ident.startswith("accessoires_capes_"):
        suffix = ident.removeprefix("accessoires_capes_")
        return "Standard" if suffix == "capes" else readable(suffix)
    if ident in {"cloth_armor", "leather_armor", "plate_armor", "cloth_shoes", "leather_shoes", "plate_shoes", "cloth_helmet", "leather_helmet", "plate_helmet", "mace"}:
        return ""
    if ident.endswith(("_fey", "_hell", "_royal", "_keeper", "_morgana", "_avalon", "_crystal", "_undead", "_heretic", "_demon")):
        return readable(ident.rsplit("_", 1)[-1])
    label = readable(ident)
    ancestor_words = {word.lower() for part in path_parts[:-1] for word in readable(part).split()}
    label_words = [word for word in label.split() if word.lower() not in ancestor_words]
    label = " ".join(label_words)
    label = re.sub(r"\b(Set)([1-3])\b", r"\1 \2", label)
    if ident.endswith("_main_mace"):
        return "One Handed"
    if ident.endswith("_2h_mace"):
        return "Two Handed"
    if parent_id and ident.startswith(parent_id + "_") and ident.endswith("_" + parent_id):
        return ""
    return label


def item_ids():
    result = []
    for line in (ROOT / "items.txt").read_text(encoding="utf-8-sig").splitlines():
        match = re.match(r"\s*\d+:\s+(\S+)", line)
        if match:
            result.append(match.group(1))
    return result


def category_data(root, xml_items):
    specs = {}

    def add(path_parts, name=None):
        if not path_parts:
            return
        path = "/".join(path_parts)
        if path not in specs:
            specs[path] = (path_parts[-1], name if name is not None else category_name(path_parts), "/".join(path_parts[:-1]))

    tree = root.find("shopcategories")
    if tree is not None:
        for first in tree:
            p1 = [first.get("id", "")]
            add(p1)
            for second in first:
                p2 = p1 + [second.get("id", "")]
                add(p2)
                for third in second:
                    p3 = p2 + [third.get("id", "")]
                    add(p3)
                    for fourth in third:
                        add(p3 + [fourth.get("id", "")])

    for item in xml_items.values():
        parts = [item.get(key) for key in ("shopcategory", "shopsubcategory1", "shopsubcategory2", "shopsubcategory3")]
        parts = [part for part in parts if part]
        for length in range(1, len(parts) + 1):
            add(parts[:length])
    add(["uncategorized"], "Uncategorized")
    return specs


def category_path(item, specs):
    if item is None:
        return "uncategorized"
    parts = [item.get(key) for key in ("shopcategory", "shopsubcategory1", "shopsubcategory2", "shopsubcategory3")]
    parts = [part for part in parts if part]
    path = "/".join(parts)
    return path if path in specs else "uncategorized"


def generate():
    localized = json.loads((ROOT / "items.json").read_text(encoding="utf-8"))
    names = {entry["UniqueName"]: entry.get("LocalizedNames", {}) for entry in localized}
    ids = item_ids()
    if len(ids) != len(set(ids)) or set(ids) != set(names):
        raise SystemExit("items.txt and items.json do not contain the same unique IDs")

    xml_root = ET.parse(ROOT / "items.xml").getroot()
    xml_items = {element.get("uniquename"): element for element in xml_root
                 if element.get("uniquename")}
    specs = category_data(xml_root, xml_items)
    categories = sorted(specs.items(), key=lambda pair: (pair[0].count("/"), pair[0]))
    cat_lines = ["package catalog", "", "// Category tree copied from the shopcategory hierarchy in items.xml.",
                 "var categorySpecs = []categorySpec{"]
    for path, (ident, name, parent) in categories:
        cat_lines.append("\t{Path: %s, ID: %s, Name: %s, ParentPath: %s}," %
                         tuple(go_string(value) for value in (path, ident, name, parent)))
    cat_lines.append("}")
    (OUT / "categories_data.go").write_text("\n".join(cat_lines) + "\n", encoding="utf-8")

    family_members = collections.defaultdict(list)
    for item_id in ids:
        without_enchantment, sep, ench_text = item_id.rpartition("@")
        enchantment = int(ench_text) if sep and ench_text.isdigit() else 0
        base = without_enchantment if sep and ench_text.isdigit() else item_id
        item = xml_items.get(base)
        if item is not None and not (sep and ench_text.isdigit()):
            enchantment = int(item.get("enchantmentlevel", enchantment))
        tier_match = re.fullmatch(r"T([1-8])_(.+)", base)
        if tier_match:
            tier, family_id = int(tier_match.group(1)), tier_match.group(2)
        else:
            tier, family_id = 0, base
        if enchantment > 0:
            family_id = re.sub(r"_LEVEL[1-4]$", "", family_id)
        category = category_path(item, specs)
        names_for_item = names[item_id] or {}
        display_name = names_for_item.get("DE-DE") or names_for_item.get("EN-US") or item_id
        family_members[family_id].append((family_id, item_id, tier, enchantment, display_name, category, item))

    families = collections.defaultdict(list)
    for family_id, entries in family_members.items():
        roots = collections.Counter(entry[5].split("/", 1)[0] for entry in entries if not entry[5].startswith("uncategorized"))
        top_category = sorted(roots, key=lambda root: (-roots[root], root))[0] if roots else "uncategorized"
        families[top_category].append((family_id, entries))

    files_by_category = collections.defaultdict(list)
    for top_category, family_groups in families.items():
        grouped = dict(family_groups)
        file = "items_" + slug(top_category) + ".go"
        lines = ["package catalog", ""]
        recipe_names = {}
        for entries in grouped.values():
            for entry in entries:
                item = entry[6]
                if item is None:
                    continue
                req = item.find("craftingrequirements")
                if req is None:
                    continue
                signature = item.get("uniquename", "") + ET.tostring(req, encoding="unicode")
                digest = hashlib.sha1(signature.encode()).hexdigest()[:12]
                recipe_names.setdefault(digest, ("recipe_" + digest, req, item.get("uniquename", "")))
        for recipe_name, req, source_id in recipe_names.values():
            lines.append("var %s = &CraftingRecipe{SourceItemID: %s, Attributes: map[string]string{" %
                         (recipe_name, go_string(source_id)))
            for key, value in sorted(req.attrib.items()):
                lines.append("\t%s: %s," % (go_string(key), go_string(value)))
            lines.append("}, Resources: []CraftingResource{")
            for resource in req.findall("craftresource"):
                rid = resource.get("uniquename", "")
                lines.append("\t{ItemID: %s, Attributes: map[string]string{" % go_string(rid))
                for key, value in sorted(resource.attrib.items()):
                    if key != "uniquename":
                        lines.append("\t\t%s: %s," % (go_string(key), go_string(value)))
                lines.append("\t}},")
            lines.append("}}\n")

        lines.append("var itemFamilies_" + slug(top_category) + " = []ItemDefinition{")
        for family_id, entries in sorted(grouped.items()):
            variants = sorted(entries, key=lambda entry: (entry[2], entry[3], entry[1]))
            tiers = [entry[2] for entry in variants if entry[2] > 0]
            enchantments = [entry[3] for entry in variants]
            min_tier, max_tier = (min(tiers), max(tiers)) if tiers else (0, 0)
            min_enchant, max_enchant = min(enchantments), max(enchantments)
            display_family_name = family_name(family_id, variants)
            lines.append("\t{Name: %s, BaseID: %s, MinTier: %d, MaxTier: %d, MinEnchantment: %d, MaxEnchantment: %d, Variants: []ItemVariant{" %
                         (go_string(display_family_name), go_string(family_id), min_tier, max_tier, min_enchant, max_enchant))
            for _, item_id, tier, enchantment, name, cat_path, item in variants:
                recipe_name = "nil"
                if item is not None:
                    req = item.find("craftingrequirements")
                    if req is not None:
                        signature = item.get("uniquename", "") + ET.tostring(req, encoding="unicode")
                        digest = hashlib.sha1(signature.encode()).hexdigest()[:12]
                        recipe_name = "recipe_" + digest
                lines.append("\t\t{ID: %s, Name: %s, FullName: %s, Tier: %d, Enchantment: %d, CategoryPath: %s, Recipe: %s}," %
                             (go_string(item_id), go_string(display_family_name), go_string(name), tier, enchantment, go_string(cat_path), recipe_name))
            lines.append("\t}},")
        lines.append("}")
        (OUT / file).write_text("\n".join(lines) + "\n", encoding="utf-8")
        files_by_category[top_category].append("itemFamilies_" + slug(top_category))

    definition_lines = ["package catalog", "", "var definitions = func() []ItemDefinition {", "\tvar all []ItemDefinition"]
    for category in sorted(files_by_category):
        definition_lines.append("\tall = append(all, %s...)" % files_by_category[category][0])
    definition_lines.extend(["\treturn all", "}()"])
    (OUT / "definitions.go").write_text("\n".join(definition_lines) + "\n", encoding="utf-8")

    for filename in OUT.glob("items_*.go"):
        if filename.name != "categories_data.go" and filename.stem.split("_", 1)[1] not in {slug(c) for c in families}:
            filename.unlink()
    print("Generated %d families, %d IDs, %d category nodes" % (len(family_members), len(ids), len(categories)))


if __name__ == "__main__":
    generate()
