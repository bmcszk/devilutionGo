# Volume II — Complete Data Inventories
Companion volume to `/tmp/dg-requirements.md` (Volume I = architecture/systems).
This volume is a **row-for-row transcription of every data table** in
`assets/txtdata/` (51 files) plus `mods/hf/txtdata/` (14 Hellfire override files),
together with the C++ algorithms that consume them.
Data-row counts, measured with `tail -n +2 <file> | grep -c .` (header excluded,
blank rows dropped). Note: the physical line count is `data rows + 1` except for
`objects/objdat.tsv` (110 lines / 109 rows — its final line has no trailing newline).
| file | data rows | | file | data rows |
|---|---|---|---|---|
| items/itemdat.tsv | 168 | | sound/effects.tsv | 814 |
| items/item_prefixes.tsv | 83 | | sound/effects-unused.tsv | 15 |
| items/item_suffixes.tsv | 95 | | text/textdat.tsv | 338 |
| items/unique_itemdat.tsv | 90 | | text/textdat-unused.tsv | 139 |
| monsters/monstdat.tsv | 112 | | towners/towners.tsv | 12 |
| monsters/unique_monstdat.tsv | 100 | | towners/quest_dialog.tsv | 10 |
| missiles/misdat.tsv | 68 | | Experience.tsv | 50 |
| missiles/missile_sprites.tsv | 46 | | classes/classdat.tsv | 6 |
| objects/objdat.tsv | 109 | | classes/{6}/animations.tsv | 27 each |
| quests/questdat.tsv | 24 | | classes/{6}/attributes.tsv | 27 each |
| spells/spelldat.tsv | 36 | | classes/{6}/sounds.tsv | 103 each |
|  |  | | classes/{6}/sprites.tsv | 13 each |
|  |  | | classes/{6}/starting_loadout.tsv | 9 each |
### V0.1 TSV format contract (verbatim from `assets/txtdata/Readme.md`)
- Tab-delimited, UTF-8, optional leading BOM, exactly one header row, no tabs inside
  values; `-` or empty cell = absent value.
- Parser: `Source/data/file.hpp:37-110` (`DataFile::loadOrDie`, `parseHeader`, `DataFileRecord`)
  + `Source/data/record_reader.hpp` (`readInt`/`readBool`/`readString`/`read`/`readEnumList`/`readEnumArray`).
- **Every loader skips column 1** (`reader.advance(); // Skip the first column (item ID)`) —
  the `id`/`type` columns are documentation/provenance, **not** parsed. Positions come from
  row order + auto-assigned `mappingId` (itemdat.cpp:585-600, :656-666).
- Table→loader map (verified file:line):
| TSV | loader (file:line) | C++ struct | enum owner |
|---|---|---|---|
| items/itemdat.tsv | tables/itemdat.cpp:616 `LoadItemDat`, per-row `:568-612` | ItemData (tables/itemdat.h:498+) | ItemType :290, ItemSpecialEffect :386, :419, AffixItemType :609 |
| items/unique_itemdat.tsv | tables/itemdat.cpp:674 `LoadUniqueItemDat`, per-row `:636-667` | UniqueItem (itemdat.h:628-655) | `_unique_items` (items.h:47, 91 entries) |
| items/item_prefixes.tsv / _suffixes.tsv | tables/itemdat.cpp:684 `LoadItemAffixesDat` (`:714-715` calls) | PLStruct | IPL power types via `ReadItemPower` :627 |
| monsters/monstdat.tsv | tables/monstdat.cpp:409 | MonsterData (monstdat.h:100-160) | MonsterAIID :22-63, MonsterClass :66, MonsterAvailability :92, treasure flags :86-88 |
| monsters/unique_monstdat.tsv | tables/monstdat.cpp:459 | UniqueMonsterData | UniqueMonsterType |
| missiles/misdat.tsv | tables/misdat.cpp:373 `LoadMisdat` (`:377-395`) | MissileData (misdat.h:166-196) | MissileID (tables/spelldat.h:98-176, 111 entries) |
| missiles/missile_sprites.tsv | tables/misdat.cpp:134 `LoadMissileSpriteData` | MissileFileData | MissileGraphicID via `ParseMissileGraphicID` :67 |
| objects/objdat.tsv | tables/objdat.cpp:259 | ObjectData (objdat.h:100-218) | theme_id :18-37, `_object_id` :42+, ObjectFlag |
| quests/questdat.tsv | tables/questdat.cpp:80 | QuestData (questdat.hpp:117-129) | quest_id :74-99, quest_state :62+ |
| spells/spelldat.tsv | tables/spelldat.cpp:232 `LoadSpellData` | SpellData (spelldat.h:225-252) | SpellID (spelldat.h:30-87), SpellType :20, MagicType |
| text/textdat.tsv | tables/textdat.cpp:106 | _speechText | `_speech_id` (textdat.h:20-448) |
| towners/towners.tsv | tables/townerdat.cpp:100 | TownerData | `_talker_id` (towners.h:26-40) |
| towners/quest_dialog.tsv | tables/townerdat.cpp:135 | towner×quest dialog matrix | quest_id × `_speech_id` |
| sound/effects.tsv | effects.cpp:141-165 `LoadEffectsData` | TSFX (`sgSFX`) | SfxID (sound_effect_enums.h:117), sfx_flag via `ParseSfxFlag` :132 |
| Experience.tsv | tables/playerdat.cpp:91 `ReloadExperienceData` | ExperienceData (:35-70) | — |
| classes/classdat.tsv | tables/playerdat.cpp:345 | HeroClass list | HeroClass (player.h) |
| classes/{c}/attributes.tsv | tables/playerdat.cpp:171 | ClassAttributes | classFlags bitset |
| classes/{c}/starting_loadout.tsv | tables/playerdat.cpp:209 | starting gear | — |
| classes/{c}/sprites.tsv | tables/playerdat.cpp:227 | sprite paths | — |
| classes/{c}/animations.tsv | tables/playerdat.cpp:251 | animation frame table | — |
| classes/{c}/sounds.tsv | tables/playerdat.cpp:289 | HeroSpeech → SfxID | HeroSpeech |
### V0.2 The Hellfire mod-override mechanism (**new architectural finding, Volume I §6 missed it**)
This fork no longer stores Hellfire content as `gbIsHellfire` code branches in the
tables — it stores it as a **built-in mod** in `mods/hf/txtdata/` that shadows the
base tables through the normal mod-override stack:
- `Source/options.cpp:71-80` `DiscoverMods()`: `modNames = {clock, adria_refills_mana,`
  `Floating Numbers - Damage, Floating Numbers - XP}`; `if (HaveHellfire()) modNames.insert("hf")`.
- `Source/engine/assets.cpp:704-760` `LoadModArchives()`: loose override dirs at
  `PrefPath/mods/<name>/` (`:729-733`) then `BasePath/mods/<name>/` (`:736-740`) are pushed into
  `OverridePaths`; packed `mods/<name>.mpq` loaded at **priority ≥ 10000** (`:746-758`), i.e. above
  the retail archives (1000) and devilutionx.mpq (9000), so its `txtdata/*.tsv` win.
- `mods/hf/` contents on disk: `data/`, `lua/`, `manifest.ini`, `nlevels/`, `txtdata/`, `ui_art/`.
- Load order consequence: a Go port must implement **table override + append** semantics,
  not just a `gbIsHellfire` flag. See V12.10 for the exact per-file deltas.
---
## V1. ITEMS — 168 bases + 83 prefixes + 95 suffixes + 90 uniques = 436 rows
### V1.1 itemdat.tsv column → C++ field mapping (`Source/tables/itemdat.cpp:568-612`)
| TSV column | C++ field (`ItemData`) | parsed as | meaning |
|---|---|---|---|
| id | — (skipped) | advance() | provenance only; `_item_indexes`/mappingId assigned by row order (`iMappingId`, itemdat.cpp:593-600) |
| dropRate | dropRate | readInt | weight in droppable-item table (items.cpp:1333 `GetItemIndexForDroppableItem`) |
| class | iClass | ParseItemClass | Weapon / Armor / Misc / Quest / Gold — vendor + affix gate |
| equipType | iLoc | ParseItemEquipType | One-handed, Two-handed, Helm, Body, Ring-Amulet, UnclassifiedItem, Unequippable |
| cursorGraphic | iCurs | ParseItemCursorGraphic | `item_cursor_id` sprite (ICURS_*) |
| itemType | itype | ParseItemType | ItemType enum (tables/itemdat.h:290) — Sword/Axe/Mace/Bow/Shield/LightArmor/Helm/MediumArmor/HeavyArmor/Staff/Ring/Amulet/Gold/Misc/None |
| uniqueBaseItem | iItemId | ParseOrAddUniqueBaseItem | the only base a given unique can roll from |
| name | iName | readString | display name |
| shortName | iSName | readString | fallback when the magic name overflows the panel (items.cpp:1104-1113) |
| minMonsterLevel | iMinMLvl | readInt | level gate for drops/vendors/uniques |
| durability | iDurability | readInt | 0 = not durability tracked; DUR_INDESTRUCTIBLE path (player.cpp:505-522) |
| minDamage / maxDamage | iMinDam / iMaxDam | readInt | weapon damage range |
| minArmor / maxArmor | iMinAC / iMaxAC | readInt | armor class range |
| minStrength / minMagic / minDexterity | iMinStr / iMinMag / iMinDex | readInt | attribute requirements |
| specialEffects | iFlags | readEnumList ParseItemSpecialEffect | ItemSpecialEffect (itemdat.h:386) + ItemSpecialEffectHf (:419) |
| miscId | iMiscId | ParseItemMiscId | IMISC_* (potion/scroll/book/staff/oil/rune/elixir/tome/gold…) |
| spell | iSpell | ParseSpellId | SpellID granted by scroll/book/staff |
| usable | iUsable | readBool | right-click usable |
| value | iValue | readInt | base value feeding CalcItemValue (items.cpp:600-610) |
### V1.2 Every item base — itemdat.tsv rows 2-169 (168 rows, complete)
Legend: `lvl`=minMonsterLevel, `dur`=durability, `dmg`=min-max damage, `ac`=min-max armor,
`req`=Str/Mag/Dex minimums (`-` = 0), `val`=value, `dr`=dropRate, `uniq-base`=uniqueBaseItem.
The `id` column is reproduced verbatim — note that only **50** rows carry an `IDI_*` label
(`LoadItemDatFromFile` skips it; `_item_indexes` order is defined by `tables/itemdat.h:24+`).
Class histogram (`class` col): Misc=68, Weapon=48, Armor=35, Quest=16, Gold=1.
#### items rows 1–42 of 168
| # | id | name | class | itemType | equip | cursor | misc | spell | lvl | dur | dmg | ac | req | val | dr | uniqBase | special |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | IDI_GOLD | Gold | Gold | Gold | Unequippable | GOLD_SMALL | NONE | Null | 1 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 2 | IDI_WARRIOR | Short Sword | Weapon | Sword | One-handed | SHORT_SWORD | NONE | Null | 2 | 24 | 2-6 | - | 18/-/- | 120 | 0 | NONE | - |
| 3 | IDI_WARRSHLD | Buckler | Armor | Shield | One-handed | BUCKLER | NONE | Null | 2 | 16 | - | 3-3 | -/-/- | 30 | 0 | NONE | - |
| 4 | IDI_WARRCLUB | Club | Weapon | Mace | One-handed | CLUB | NONE | Null | 1 | 20 | 1-6 | - | -/-/- | 20 | 0 | SPIKCLUB | - |
| 5 | IDI_ROGUE | Short Bow | Weapon | Bow | Two-handed | SHORT_BOW | NONE | Null | 1 | 30 | 1-4 | - | -/-/- | 100 | 0 | NONE | - |
| 6 | IDI_SORCERER | Short Staff of Mana | Weapon | Staff | Two-handed | SHORT_STAFF | STAFF | Mana | 1 | 25 | 2-4 | - | -/17/- | 210 | 0 | NONE | - |
| 7 | IDI_CLEAVER | Cleaver | Weapon | Axe | Two-handed | CLEAVER | UNIQUE | Null | 10 | 10 | 4-24 | - | -/-/- | 2000 | 0 | CLEAVER | - |
| 8 | IDI_SKCROWN | The Undead Crown | Armor | Helm | Helm | THE_UNDEAD_CROWN | UNIQUE | Null | 0 | 50 | - | 15-15 | -/-/- | 10000 | 0 | SKCROWN | RandomStealLife |
| 9 | IDI_INFRARING | Empyrean Band | Misc | Ring | Ring | EMPYREAN_BAND | UNIQUE | Null | 0 | 0 | - | - | -/-/- | 8000 | 0 | INFRARING | - |
| 10 | IDI_ROCK | Magic Rock | Quest | Misc | Unequippable | MAGIC_ROCK | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 11 | IDI_OPTAMULET | Optic Amulet | Misc | Amulet | Amulet | OPTIC_AMULET | UNIQUE | Null | 0 | 0 | - | - | -/-/- | 5000 | 0 | OPTAMULET | - |
| 12 | IDI_TRING | Ring of Truth | Misc | Ring | Ring | RING_OF_TRUTH | UNIQUE | Null | 0 | 0 | - | - | -/-/- | 1000 | 0 | TRING | - |
| 13 | IDI_BANNER | Tavern Sign | Quest | Misc | Unequippable | TAVERN_SIGN | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 14 | IDI_HARCREST | Harlequin Crest | Armor | Helm | Helm | HARLEQUIN_CREST | UNIQUE | Null | 0 | 15 | - | - | -/-/- | 15 | 0 | HARCREST | - |
| 15 | IDI_STEELVEIL | Veil of Steel | Armor | Helm | Helm | VEIL_OF_STEEL | UNIQUE | Null | 0 | 60 | - | 18-18 | -/-/- | 0 | 0 | STEELVEIL | - |
| 16 | IDI_GLDNELIX | Golden Elixir | Misc | Misc | Unequippable | GOLDEN_ELIXIR | NONE | Null | 15 | 0 | - | - | -/-/- | 0 | 0 | ELIXIR | - |
| 17 | IDI_ANVIL | Anvil of Fury | Quest | Misc | Unequippable | ANVIL_OF_FURY | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 18 | IDI_MUSHROOM | Black Mushroom | Quest | Misc | Unequippable | BLACK_MUSHROOM | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 19 | IDI_BRAIN | Brain | Quest | Misc | Unequippable | BRAIN | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 20 | IDI_FUNGALTM | Fungal Tome | Quest | Misc | Unequippable | FUNGAL_TOME | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 21 | IDI_SPECELIX | Spectral Elixir | Misc | Misc | Unequippable | SPECTRAL_ELIXIR | SPECELIX | Null | 15 | 0 | - | - | -/-/- | 0 | 0 | ELIXIR | - |
| 22 | IDI_BLDSTONE | Blood Stone | Quest | Misc | Unequippable | BLOOD_STONE | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 23 | IDI_MAPOFDOOM | Cathedral Map | Quest | Misc | Unequippable | MAP_OF_THE_STARS | MAPOFDOOM | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | MAPOFDOOM | - |
| 24 | IDI_EAR | Ear | Misc | Misc | Unequippable | - | EAR | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 25 | IDI_HEAL | Potion of Healing | Misc | Misc | Unequippable | POTION_OF_HEALING | HEAL | Null | 0 | 0 | - | - | -/-/- | 50 | 0 | NONE | - |
| 26 | IDI_MANA | Potion of Mana | Misc | Misc | Unequippable | POTION_OF_MANA | MANA | Null | 0 | 0 | - | - | -/-/- | 50 | 0 | NONE | - |
| 27 | IDI_IDENTIFY | Scroll of Identify | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Identify | 1 | 0 | - | - | -/-/- | 200 | 0 | NONE | - |
| 28 | IDI_PORTAL | Scroll of Town Portal | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | TownPortal | 4 | 0 | - | - | -/-/- | 200 | 0 | NONE | - |
| 29 | IDI_ARMOFVAL | Arkaine's Valor | Armor | MediumArmor | Armor | ARKAINES_VALOR | UNIQUE | Null | 0 | 40 | - | - | -/-/- | 0 | 0 | ARMOFVAL | - |
| 30 | IDI_FULLHEAL | Potion of Full Healing | Misc | Misc | Unequippable | POTION_OF_FULL_HEALING | FULLHEAL | Null | 1 | 0 | - | - | -/-/- | 150 | 0 | NONE | - |
| 31 | IDI_FULLMANA | Potion of Full Mana | Misc | Misc | Unequippable | POTION_OF_FULL_MANA | FULLMANA | Null | 1 | 0 | - | - | -/-/- | 150 | 0 | NONE | - |
| 32 | IDI_GRISWOLD | Griswold's Edge | Weapon | Sword | One-handed | BROAD_SWORD | UNIQUE | Null | 8 | 50 | 4-12 | - | 40/-/- | 750 | 0 | GRISWOLD | - |
| 33 | IDI_LGTFORGE | Bovine Plate | Armor | HeavyArmor | Armor | BOVINE | UNIQUE | Null | 0 | 40 | - | - | 50/-/- | 0 | 0 | BOVINE | - |
| 34 | IDI_LAZSTAFF | Staff of Lazarus | Misc | Misc | Unequippable | STAFF_OF_LAZARUS | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | LAZSTAFF | - |
| 35 | IDI_RESURRECT | Scroll of Resurrect | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Resurrect | 1 | 0 | - | - | -/-/- | 250 | 0 | NONE | - |
| 36 | IDI_OIL | Blacksmith Oil | Misc | Misc | Unequippable | OIL | OILBSMTH | Null | 1 | 0 | - | - | -/-/- | 100 | 0 | NONE | - |
| 37 | IDI_SHORTSTAFF | Short Staff | Weapon | Staff | Two-handed | SHORT_STAFF | NONE | Null | 1 | 25 | 2-4 | - | -/-/- | 20 | 0 | NONE | - |
| 38 | IDI_BARDSWORD | Sword | Weapon | Sword | One-handed | SHORT_SWORD | NONE | Null | 2 | 8 | 1-5 | - | 15/-/20 | 20 | 0 | NONE | - |
| 39 | IDI_BARDDAGGER | Dagger | Weapon | Sword | One-handed | DAGGER | NONE | Null | 1 | 16 | 1-4 | - | -/-/- | 20 | 0 | NONE | - |
| 40 | IDI_RUNEBOMB | Rune Bomb | Quest | Misc | Unequippable | RUNE_BOMB | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 41 | IDI_THEODORE | Theodore | Quest | Misc | Unequippable | THEODORE | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 42 | IDI_AURIC | Auric Amulet | Misc | Misc | Amulet | AURIC_AMULET | AURIC | Null | 0 | 0 | - | - | -/-/- | 100 | 0 | NONE | - |
#### items rows 43–84 of 168
| # | id | name | class | itemType | equip | cursor | misc | spell | lvl | dur | dmg | ac | req | val | dr | uniqBase | special |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 43 | IDI_NOTE1 | Torn Note 1 | Quest | Misc | Unequippable | TORN_NOTE_1 | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 44 | IDI_NOTE2 | Torn Note 2 | Quest | Misc | Unequippable | TORN_NOTE_2 | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 45 | IDI_NOTE3 | Torn Note 3 | Quest | Misc | Unequippable | TORN_NOTE_3 | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 46 | IDI_FULLNOTE | Reconstructed Note | Quest | Misc | Unequippable | RECONSTRUCTED_NOTE | NOTE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 47 | IDI_BROWNSUIT | Brown Suit | Quest | Misc | Unequippable | BROWN_SUIT | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 48 | IDI_GREYSUIT | Grey Suit | Quest | Misc | Unequippable | GREY_SUIT | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 49 | - | Cap | Armor | Helm | Helm | CAP | NONE | Null | 1 | 15 | - | 1-3 | -/-/- | 15 | 1 | NONE | - |
| 50 | - | Skull Cap | Armor | Helm | Helm | SKULL_CAP | NONE | Null | 4 | 20 | - | 2-4 | -/-/- | 25 | 1 | SKULLCAP | - |
| 51 | - | Helm | Armor | Helm | Helm | HELM | NONE | Null | 8 | 30 | - | 4-6 | 25/-/- | 40 | 1 | HELM | - |
| 52 | - | Full Helm | Armor | Helm | Helm | FULL_HELM | NONE | Null | 12 | 35 | - | 6-8 | 35/-/- | 90 | 1 | NONE | - |
| 53 | - | Crown | Armor | Helm | Helm | CROWN | NONE | Null | 16 | 40 | - | 8-12 | -/-/- | 200 | 1 | CROWN | - |
| 54 | - | Great Helm | Armor | Helm | Helm | GREAT_HELM | NONE | Null | 20 | 60 | - | 10-15 | 50/-/- | 400 | 1 | GREATHELM | - |
| 55 | - | Cape | Armor | LightArmor | Armor | CAPE | NONE | Null | 1 | 12 | - | 1-5 | -/-/- | 10 | 1 | CAPE | - |
| 56 | - | Rags | Armor | LightArmor | Armor | RAGS | NONE | Null | 1 | 6 | - | 2-6 | -/-/- | 5 | 1 | RAGS | - |
| 57 | - | Cloak | Armor | LightArmor | Armor | CLOAK | NONE | Null | 2 | 18 | - | 3-7 | -/-/- | 40 | 1 | CLOAK | - |
| 58 | - | Robe | Armor | LightArmor | Armor | ROBE | NONE | Null | 3 | 24 | - | 4-7 | -/-/- | 75 | 1 | ROBE | - |
| 59 | - | Quilted Armor | Armor | LightArmor | Armor | QUILTED_ARMOR | NONE | Null | 4 | 30 | - | 7-10 | -/-/- | 200 | 1 | NONE | - |
| 60 | - | Leather Armor | Armor | LightArmor | Armor | LEATHER_ARMOR | NONE | Null | 6 | 35 | - | 10-13 | -/-/- | 300 | 1 | LEATHARMOR | - |
| 61 | - | Hard Leather Armor | Armor | LightArmor | Armor | HARD_LEATHER_ARMOR | NONE | Null | 7 | 40 | - | 11-14 | -/-/- | 450 | 1 | NONE | - |
| 62 | - | Studded Leather Armor | Armor | LightArmor | Armor | STUDDED_LEATHER_ARMOR | NONE | Null | 9 | 45 | - | 15-17 | 20/-/- | 700 | 1 | STUDARMOR | - |
| 63 | - | Ring Mail | Armor | MediumArmor | Armor | RING_MAIL | NONE | Null | 11 | 50 | - | 17-20 | 25/-/- | 900 | 1 | NONE | - |
| 64 | - | Chain Mail | Armor | MediumArmor | Armor | CHAIN_MAIL | NONE | Null | 13 | 55 | - | 18-22 | 30/-/- | 1250 | 1 | CHAINMAIL | - |
| 65 | - | Scale Mail | Armor | MediumArmor | Armor | SCALE_MAIL | NONE | Null | 15 | 60 | - | 23-28 | 35/-/- | 2300 | 1 | NONE | - |
| 66 | - | Breast Plate | Armor | HeavyArmor | Armor | BREAST_PLATE | NONE | Null | 16 | 80 | - | 20-24 | 40/-/- | 2800 | 1 | BREASTPLATE | - |
| 67 | - | Splint Mail | Armor | MediumArmor | Armor | SPLINT_MAIL | NONE | Null | 17 | 65 | - | 30-35 | 40/-/- | 3250 | 1 | NONE | - |
| 68 | - | Plate Mail | Armor | HeavyArmor | Armor | FIELD_PLATE | NONE | Null | 19 | 75 | - | 42-50 | 60/-/- | 4600 | 1 | PLATEMAIL | - |
| 69 | - | Field Plate | Armor | HeavyArmor | Armor | FIELD_PLATE | NONE | Null | 21 | 80 | - | 40-45 | 65/-/- | 5800 | 1 | NONE | - |
| 70 | - | Gothic Plate | Armor | HeavyArmor | Armor | GOTHIC_PLATE | NONE | Null | 23 | 100 | - | 50-60 | 80/-/- | 8000 | 1 | NONE | - |
| 71 | - | Full Plate Mail | Armor | HeavyArmor | Armor | FULL_PLATE_MAIL | NONE | Null | 25 | 90 | - | 60-75 | 90/-/- | 6500 | 1 | FULLPLATE | - |
| 72 | - | Buckler | Armor | Shield | One-handed | BUCKLER | NONE | Null | 1 | 16 | - | 1-5 | -/-/- | 30 | 1 | BUCKLER | - |
| 73 | - | Small Shield | Armor | Shield | One-handed | SMALL_SHIELD | NONE | Null | 5 | 24 | - | 3-8 | 25/-/- | 90 | 1 | SMALLSHIELD | - |
| 74 | - | Large Shield | Armor | Shield | One-handed | LARGE_SHIELD | NONE | Null | 9 | 32 | - | 5-10 | 40/-/- | 200 | 1 | LARGESHIELD | - |
| 75 | - | Kite Shield | Armor | Shield | One-handed | KITE_SHIELD | NONE | Null | 14 | 40 | - | 8-15 | 50/-/- | 400 | 1 | KITESHIELD | - |
| 76 | - | Tower Shield | Armor | Shield | One-handed | TOWER_SHIELD | NONE | Null | 20 | 50 | - | 12-20 | 60/-/- | 850 | 1 | GOTHSHIELD | - |
| 77 | - | Gothic Shield | Armor | Shield | One-handed | GOTHIC_SHIELD | NONE | Null | 23 | 60 | - | 14-18 | 80/-/- | 2300 | 1 | GOTHSHIELD | - |
| 78 | - | Potion of Healing | Misc | Misc | Unequippable | POTION_OF_HEALING | HEAL | Null | 1 | 0 | - | - | -/-/- | 50 | 1 | NONE | - |
| 79 | - | Potion of Full Healing | Misc | Misc | Unequippable | POTION_OF_FULL_HEALING | FULLHEAL | Null | 1 | 0 | - | - | -/-/- | 150 | 1 | NONE | - |
| 80 | - | Potion of Mana | Misc | Misc | Unequippable | POTION_OF_MANA | MANA | Null | 1 | 0 | - | - | -/-/- | 50 | 1 | NONE | - |
| 81 | - | Potion of Full Mana | Misc | Misc | Unequippable | POTION_OF_FULL_MANA | FULLMANA | Null | 1 | 0 | - | - | -/-/- | 150 | 1 | NONE | - |
| 82 | - | Potion of Rejuvenation | Misc | Misc | Unequippable | POTION_OF_REJUVENATION | REJUV | Null | 3 | 0 | - | - | -/-/- | 120 | 1 | NONE | - |
| 83 | - | Potion of Full Rejuvenation | Misc | Misc | Unequippable | POTION_OF_FULL_REJUVENATION | FULLREJUV | Null | 7 | 0 | - | - | -/-/- | 600 | 1 | NONE | - |
| 84 | - | Blacksmith Oil | Misc | Misc | Unequippable | OIL | OILBSMTH | Null | 1 | 0 | - | - | -/-/- | 100 | 1 | NONE | - |
#### items rows 85–126 of 168
| # | id | name | class | itemType | equip | cursor | misc | spell | lvl | dur | dmg | ac | req | val | dr | uniqBase | special |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 85 | - | Oil of Accuracy | Misc | Misc | Unequippable | OIL | OILACC | Null | 1 | 0 | - | - | -/-/- | 500 | 1 | NONE | - |
| 86 | - | Oil of Sharpness | Misc | Misc | Unequippable | OIL | OILSHARP | Null | 1 | 0 | - | - | -/-/- | 500 | 1 | NONE | - |
| 87 | - | Oil | Misc | Misc | Unequippable | OIL | OILOF | Null | 10 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 88 | - | Elixir of Strength | Misc | Misc | Unequippable | ELIXIR_OF_STRENGTH | ELIXSTR | Null | 15 | 0 | - | - | -/-/- | 5000 | 1 | NONE | - |
| 89 | - | Elixir of Magic | Misc | Misc | Unequippable | ELIXIR_OF_MAGIC | ELIXMAG | Null | 15 | 0 | - | - | -/-/- | 5000 | 1 | NONE | - |
| 90 | - | Elixir of Dexterity | Misc | Misc | Unequippable | ELIXIR_OF_DEXTERITY | ELIXDEX | Null | 15 | 0 | - | - | -/-/- | 5000 | 1 | NONE | - |
| 91 | - | Elixir of Vitality | Misc | Misc | Unequippable | ELIXIR_OF_VITALITY | ELIXVIT | Null | 20 | 0 | - | - | -/-/- | 5000 | 1 | NONE | - |
| 92 | - | Scroll of Healing | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Healing | 1 | 0 | - | - | -/-/- | 50 | 1 | NONE | - |
| 93 | - | Scroll of Search | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Search | 1 | 0 | - | - | -/-/- | 50 | 1 | NONE | - |
| 94 | - | Scroll of Lightning | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Lightning | 4 | 0 | - | - | -/-/- | 150 | 1 | NONE | - |
| 95 | - | Scroll of Identify | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Identify | 1 | 0 | - | - | -/-/- | 100 | 1 | NONE | - |
| 96 | - | Scroll of Resurrect | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Resurrect | 1 | 0 | - | - | -/-/- | 250 | 1 | NONE | - |
| 97 | - | Scroll of Fire Wall | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | FireWall | 4 | 0 | - | - | -/17/- | 400 | 1 | NONE | - |
| 98 | - | Scroll of Inferno | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Inferno | 1 | 0 | - | - | -/19/- | 100 | 1 | NONE | - |
| 99 | - | Scroll of Town Portal | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | TownPortal | 4 | 0 | - | - | -/-/- | 200 | 1 | NONE | - |
| 100 | - | Scroll of Flash | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Flash | 6 | 0 | - | - | -/21/- | 500 | 1 | NONE | - |
| 101 | - | Scroll of Infravision | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Infravision | 8 | 0 | - | - | -/23/- | 600 | 1 | NONE | - |
| 102 | - | Scroll of Phasing | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Phasing | 6 | 0 | - | - | -/25/- | 200 | 1 | NONE | - |
| 103 | - | Scroll of Mana Shield | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | ManaShield | 8 | 0 | - | - | -/-/- | 1200 | 1 | NONE | - |
| 104 | - | Scroll of Flame Wave | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | FlameWave | 10 | 0 | - | - | -/29/- | 650 | 1 | NONE | - |
| 105 | - | Scroll of Fireball | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Fireball | 8 | 0 | - | - | -/31/- | 300 | 1 | NONE | - |
| 106 | - | Scroll of Stone Curse | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | StoneCurse | 6 | 0 | - | - | -/33/- | 800 | 1 | NONE | - |
| 107 | - | Scroll of Chain Lightning | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | ChainLightning | 10 | 0 | - | - | -/35/- | 750 | 1 | NONE | - |
| 108 | - | Scroll of Guardian | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Guardian | 12 | 0 | - | - | -/47/- | 950 | 1 | NONE | - |
| 109 | - | (empty) | Misc | Misc | Unequippable | - | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 110 | - | Scroll of Nova | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Nova | 14 | 0 | - | - | -/57/- | 1300 | 1 | NONE | - |
| 111 | - | Scroll of Golem | Misc | Misc | Unequippable | SCROLL_OF | SCROLLT | Golem | 10 | 0 | - | - | -/51/- | 1100 | 1 | NONE | - |
| 112 | - | (empty) | Misc | Misc | Unequippable | - | NONE | Null | 0 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
| 113 | - | Scroll of Teleport | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Teleport | 14 | 0 | - | - | -/81/- | 3000 | 1 | NONE | - |
| 114 | - | Scroll of Apocalypse | Misc | Misc | Unequippable | SCROLL_OF | SCROLL | Apocalypse | 22 | 0 | - | - | -/117/- | 2000 | 1 | NONE | - |
| 115 | - | Book of | Misc | Misc | Unequippable | BOOK_BLUE | BOOK | Null | 2 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 116 | - | Book of | Misc | Misc | Unequippable | BOOK_BLUE | BOOK | Null | 8 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 117 | - | Book of | Misc | Misc | Unequippable | BOOK_BLUE | BOOK | Null | 14 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 118 | - | Book of | Misc | Misc | Unequippable | BOOK_BLUE | BOOK | Null | 20 | 0 | - | - | -/-/- | 0 | 1 | NONE | - |
| 119 | - | Dagger | Weapon | Sword | One-handed | DAGGER | NONE | Null | 1 | 16 | 1-4 | - | -/-/- | 60 | 1 | DAGGER | - |
| 120 | - | Short Sword | Weapon | Sword | One-handed | SHORT_SWORD | NONE | Null | 1 | 24 | 2-6 | - | 18/-/- | 120 | 1 | NONE | - |
| 121 | - | Falchion | Weapon | Sword | One-handed | FALCHION | NONE | Null | 2 | 20 | 4-8 | - | 30/-/- | 250 | 1 | FALCHION | - |
| 122 | - | Scimitar | Weapon | Sword | One-handed | SCIMITAR | NONE | Null | 4 | 28 | 3-7 | - | 23/-/23 | 200 | 1 | SCIMITAR | - |
| 123 | - | Claymore | Weapon | Sword | One-handed | CLAYMORE | NONE | Null | 5 | 36 | 1-12 | - | 35/-/- | 450 | 1 | CLAYMORE | - |
| 124 | - | Blade | Weapon | Sword | One-handed | BLADE | NONE | Null | 4 | 30 | 3-8 | - | 25/-/30 | 280 | 1 | NONE | - |
| 125 | - | Sabre | Weapon | Sword | One-handed | SABRE | NONE | Null | 1 | 45 | 1-8 | - | 17/-/- | 170 | 1 | SABRE | - |
| 126 | - | Long Sword | Weapon | Sword | One-handed | LONG_SWORD | NONE | Null | 6 | 40 | 2-10 | - | 30/-/30 | 350 | 1 | LONGSWR | - |
#### items rows 127–168 of 168
| # | id | name | class | itemType | equip | cursor | misc | spell | lvl | dur | dmg | ac | req | val | dr | uniqBase | special |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 127 | - | Broad Sword | Weapon | Sword | One-handed | BROAD_SWORD | NONE | Null | 8 | 50 | 4-12 | - | 40/-/- | 750 | 1 | BROADSWR | - |
| 128 | - | Bastard Sword | Weapon | Sword | One-handed | BASTARD_SWORD | NONE | Null | 10 | 60 | 6-15 | - | 50/-/- | 1000 | 1 | BASTARDSWR | - |
| 129 | - | Two-Handed Sword | Weapon | Sword | Two-handed | TWO_HANDED_SWORD | NONE | Null | 14 | 75 | 8-16 | - | 65/-/- | 1800 | 1 | TWOHANDSWR | - |
| 130 | - | Great Sword | Weapon | Sword | Two-handed | GREAT_SWORD | NONE | Null | 17 | 100 | 10-20 | - | 75/-/- | 3000 | 1 | GREATSWR | - |
| 131 | - | Small Axe | Weapon | Axe | Two-handed | SMALL_AXE | NONE | Null | 2 | 24 | 2-10 | - | -/-/- | 150 | 1 | SMALLAXE | - |
| 132 | - | Axe | Weapon | Axe | Two-handed | AXE | NONE | Null | 4 | 32 | 4-12 | - | 22/-/- | 450 | 1 | NONE | - |
| 133 | - | Large Axe | Weapon | Axe | Two-handed | LARGE_AXE | NONE | Null | 6 | 40 | 6-16 | - | 30/-/- | 750 | 1 | LARGEAXE | - |
| 134 | - | Broad Axe | Weapon | Axe | Two-handed | BROAD_AXE | NONE | Null | 8 | 50 | 8-20 | - | 50/-/- | 1000 | 1 | BROADAXE | - |
| 135 | - | Battle Axe | Weapon | Axe | Two-handed | BATTLE_AXE | NONE | Null | 10 | 60 | 10-25 | - | 65/-/- | 1500 | 1 | BATTLEAXE | - |
| 136 | - | Great Axe | Weapon | Axe | Two-handed | GREAT_AXE | NONE | Null | 12 | 75 | 12-30 | - | 80/-/- | 2500 | 1 | GREATAXE | - |
| 137 | - | Mace | Weapon | Mace | One-handed | MACE | NONE | Null | 2 | 32 | 1-8 | - | 16/-/- | 200 | 1 | MACE | - |
| 138 | - | Morning Star | Weapon | Mace | One-handed | MORNING_STAR | NONE | Null | 3 | 40 | 1-10 | - | 26/-/- | 300 | 1 | MORNSTAR | - |
| 139 | - | War Hammer | Weapon | Mace | One-handed | WAR_HAMMER | NONE | Null | 5 | 50 | 5-9 | - | 40/-/- | 600 | 1 | WARHAMMER | - |
| 140 | - | Spiked Club | Weapon | Mace | One-handed | SPIKED_CLUB | NONE | Null | 4 | 20 | 3-6 | - | 18/-/- | 225 | 1 | SPIKCLUB | - |
| 141 | - | Club | Weapon | Mace | One-handed | CLUB | NONE | Null | 1 | 20 | 1-6 | - | -/-/- | 20 | 1 | SPIKCLUB | - |
| 142 | - | Flail | Weapon | Mace | One-handed | FLAIL | NONE | Null | 7 | 36 | 2-12 | - | 30/-/- | 500 | 1 | FLAIL | - |
| 143 | - | Maul | Weapon | Mace | Two-handed | MAUL | NONE | Null | 10 | 50 | 6-20 | - | 55/-/- | 900 | 1 | MAUL | - |
| 144 | - | Short Bow | Weapon | Bow | Two-handed | SHORT_BOW | NONE | Null | 1 | 30 | 1-4 | - | -/-/- | 100 | 2 | SHORTBOW | - |
| 145 | - | Hunter's Bow | Weapon | Bow | Two-handed | HUNTERS_BOW | NONE | Null | 3 | 40 | 2-5 | - | 20/-/35 | 350 | 2 | HUNTBOW | - |
| 146 | - | Long Bow | Weapon | Bow | Two-handed | HUNTERS_BOW | NONE | Null | 5 | 35 | 1-6 | - | 25/-/30 | 250 | 2 | LONGBOW | - |
| 147 | - | Composite Bow | Weapon | Bow | Two-handed | COMPOSITE_BOW | NONE | Null | 7 | 45 | 3-6 | - | 25/-/40 | 600 | 2 | COMPBOW | - |
| 148 | IDI_SHORT_BATTLE_BOW | Short Battle Bow | Weapon | Bow | Two-handed | SHORT_BATTLE_BOW | NONE | Null | 9 | 45 | 3-7 | - | 30/-/50 | 750 | 2 | NONE | - |
| 149 | - | Long Battle Bow | Weapon | Bow | Two-handed | LONG_BATTLE_BOW | NONE | Null | 11 | 50 | 1-10 | - | 30/-/60 | 1000 | 2 | BATTLEBOW | - |
| 150 | - | Short War Bow | Weapon | Bow | Two-handed | SHORT_WAR_BOW | NONE | Null | 15 | 55 | 4-8 | - | 35/-/70 | 1500 | 2 | NONE | - |
| 151 | - | Long War Bow | Weapon | Bow | Two-handed | LONG_WAR_BOW | NONE | Null | 19 | 60 | 1-14 | - | 45/-/80 | 2000 | 2 | WARBOW | - |
| 152 | - | Short Staff | Weapon | Staff | Two-handed | SHORT_STAFF | STAFF | Null | 1 | 25 | 2-4 | - | -/-/- | 30 | 1 | SHORTSTAFF | - |
| 153 | - | Long Staff | Weapon | Staff | Two-handed | LONG_STAFF | STAFF | Null | 4 | 35 | 4-8 | - | -/-/- | 100 | 1 | LONGSTAFF | - |
| 154 | - | Composite Staff | Weapon | Staff | Two-handed | COMPOSITE_STAFF | STAFF | Null | 6 | 45 | 5-10 | - | -/-/- | 500 | 1 | COMPSTAFF | - |
| 155 | - | Quarter Staff | Weapon | Staff | Two-handed | SHORT_STAFF | STAFF | Null | 9 | 55 | 6-12 | - | 20/-/- | 1000 | 1 | QUARSTAFF | - |
| 156 | - | War Staff | Weapon | Staff | Two-handed | WAR_STAFF | STAFF | Null | 12 | 75 | 8-16 | - | 30/-/- | 1500 | 1 | WARSTAFF | - |
| 157 | - | Ring | Misc | Ring | Ring | RING | RING | Null | 5 | 0 | - | - | -/-/- | 1000 | 1 | RING | - |
| 158 | - | Ring | Misc | Ring | Ring | RING | RING | Null | 10 | 0 | - | - | -/-/- | 1000 | 1 | RING | - |
| 159 | - | Ring | Misc | Ring | Ring | RING | RING | Null | 15 | 0 | - | - | -/-/- | 1000 | 1 | RING | - |
| 160 | - | Amulet | Misc | Amulet | Amulet | AMULET | AMULET | Null | 8 | 0 | - | - | -/-/- | 1200 | 1 | AMULET | - |
| 161 | - | Amulet | Misc | Amulet | Amulet | AMULET | AMULET | Null | 16 | 0 | - | - | -/-/- | 1200 | 1 | AMULET | - |
| 162 | - | Rune of Fire | Misc | Misc | Unequippable | RUNE_OF_FIRE | RUNEF | Null | 1 | 0 | - | - | -/-/- | 100 | 1 | NONE | - |
| 163 | - | Rune of Lightning | Misc | Misc | Unequippable | RUNE_OF_LIGHTNING | RUNEL | Null | 3 | 0 | - | - | -/13/- | 200 | 1 | NONE | - |
| 164 | - | Greater Rune of Fire | Misc | Misc | Unequippable | GREATER_RUNE_OF_FIRE | GR_RUNEF | Null | 7 | 0 | - | - | -/42/- | 400 | 1 | NONE | - |
| 165 | - | Greater Rune of Lightning | Misc | Misc | Unequippable | GREATER_RUNE_OF_LIGHTNING | GR_RUNEL | Null | 7 | 0 | - | - | -/42/- | 500 | 1 | NONE | - |
| 166 | - | Rune of Stone | Misc | Misc | Unequippable | RUNE_OF_STONE | RUNES | Null | 7 | 0 | - | - | -/25/- | 300 | 1 | NONE | - |
| 167 | - | Short Staff of Charged Bolt | Weapon | Staff | Two-handed | SHORT_STAFF | STAFF | ChargedBolt | 1 | 25 | 2-4 | - | -/25/- | 470 | 0 | NONE | - |
| 168 | IDI_ARENAPOT | Arena Potion | Misc | Misc | Unequippable | ARENA_POTION | ARENAPOT | Null | 7 | 0 | - | - | -/-/- | 0 | 0 | NONE | - |
Spot-checks: row 1 = `IDI_GOLD` (Gold, `GOLD_SMALL`, value 0, dropRate 1); row 168 = `IDI_ARENAPOT` / Arena Potion.
### V1.3 Prefixes — item_prefixes.tsv rows 2-84 (83 rows, complete)
Schema (`PLStruct`, parsed at `tables/itemdat.cpp:684-730`): `name | power (IPL_*) | power.value1 |
power.value2 | minLevel | itemTypes (AffixItemType list) | alignment (goodorevil) | chance
(PLChance = weight) | useful (PLOk) | minVal | maxVal | multVal`.
#### prefixes rows 1–42 of 83
| # | name | power | v1 | v2 | lvl | itemTypes | align | chance | ok | minVal | maxVal | multVal |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Tin | TOHIT_CURSE | 6 | 10 | 3 | Weapon,Bow,Misc | Any | 2 | false | 0 | 0 | -3 |
| 2 | Brass | TOHIT_CURSE | 1 | 5 | 1 | Weapon,Bow,Misc | Any | 2 | false | 0 | 0 | -2 |
| 3 | Bronze | TOHIT | 1 | 5 | 1 | Weapon,Bow,Misc | Any | 2 | true | 100 | 500 | 2 |
| 4 | Iron | TOHIT | 6 | 10 | 4 | Weapon,Bow,Misc | Any | 2 | true | 600 | 1000 | 3 |
| 5 | Steel | TOHIT | 11 | 15 | 6 | Weapon,Bow,Misc | Any | 2 | true | 1100 | 1500 | 5 |
| 6 | Silver | TOHIT | 16 | 20 | 9 | Weapon,Bow,Misc | Good | 2 | true | 1600 | 2000 | 7 |
| 7 | Gold | TOHIT | 21 | 30 | 12 | Weapon,Bow,Misc | Good | 2 | true | 2100 | 3000 | 9 |
| 8 | Platinum | TOHIT | 31 | 40 | 16 | Weapon,Bow | Good | 2 | true | 3100 | 4000 | 11 |
| 9 | Mithril | TOHIT | 41 | 60 | 20 | Weapon,Bow | Good | 2 | true | 4100 | 6000 | 13 |
| 10 | Meteoric | TOHIT | 61 | 80 | 23 | Weapon,Bow | Any | 2 | true | 6100 | 10000 | 15 |
| 11 | Weird | TOHIT | 81 | 100 | 35 | Weapon,Bow | Any | 2 | true | 10100 | 14000 | 17 |
| 12 | Strange | TOHIT | 101 | 150 | 60 | Weapon,Bow | Any | 2 | true | 14100 | 20000 | 20 |
| 13 | Useless | DAMP_CURSE | 100 | 100 | 5 | Weapon,Bow | Any | 2 | false | 0 | 0 | -8 |
| 14 | Bent | DAMP_CURSE | 50 | 75 | 3 | Weapon,Bow | Any | 2 | false | 0 | 0 | -4 |
| 15 | Weak | DAMP_CURSE | 25 | 45 | 1 | Weapon,Bow | Any | 2 | false | 0 | 0 | -3 |
| 16 | Jagged | DAMP | 20 | 35 | 4 | Weapon,Bow | Any | 2 | true | 250 | 450 | 3 |
| 17 | Deadly | DAMP | 36 | 50 | 6 | Weapon,Bow | Any | 2 | true | 500 | 700 | 4 |
| 18 | Heavy | DAMP | 51 | 65 | 9 | Weapon,Bow | Any | 2 | true | 750 | 950 | 5 |
| 19 | Vicious | DAMP | 66 | 80 | 12 | Weapon,Bow | Evil | 2 | true | 1000 | 1450 | 8 |
| 20 | Brutal | DAMP | 81 | 95 | 16 | Weapon,Bow | Any | 2 | true | 1500 | 1950 | 10 |
| 21 | Massive | DAMP | 96 | 110 | 20 | Weapon,Bow | Any | 2 | true | 2000 | 2450 | 13 |
| 22 | Savage | DAMP | 111 | 125 | 23 | Weapon,Bow | Any | 2 | true | 2500 | 3000 | 15 |
| 23 | Ruthless | DAMP | 126 | 150 | 35 | Weapon,Bow | Any | 2 | true | 10100 | 15000 | 17 |
| 24 | Merciless | DAMP | 151 | 175 | 60 | Weapon,Bow | Any | 2 | true | 15000 | 20000 | 20 |
| 25 | Clumsy | TOHIT_DAMP_CURSE | 50 | 75 | 5 | Weapon,Staff,Bow | Any | 2 | false | 0 | 0 | -7 |
| 26 | Dull | TOHIT_DAMP_CURSE | 25 | 45 | 1 | Weapon,Staff,Bow | Any | 2 | false | 0 | 0 | -5 |
| 27 | Sharp | TOHIT_DAMP | 20 | 35 | 1 | Weapon,Staff,Bow | Any | 2 | false | 350 | 950 | 5 |
| 28 | Fine | TOHIT_DAMP | 36 | 50 | 6 | Weapon,Staff,Bow | Any | 2 | true | 1100 | 1700 | 7 |
| 29 | Warrior's | TOHIT_DAMP | 51 | 65 | 10 | Weapon,Staff,Bow | Any | 2 | true | 1850 | 2450 | 13 |
| 30 | Soldier's | TOHIT_DAMP | 66 | 80 | 15 | Weapon,Staff | Any | 2 | true | 2600 | 3950 | 17 |
| 31 | Lord's | TOHIT_DAMP | 81 | 95 | 19 | Weapon,Staff | Any | 2 | true | 4100 | 5950 | 21 |
| 32 | Knight's | TOHIT_DAMP | 96 | 110 | 23 | Weapon,Staff | Any | 2 | true | 6100 | 8450 | 26 |
| 33 | Master's | TOHIT_DAMP | 111 | 125 | 28 | Weapon,Staff | Any | 2 | true | 8600 | 13000 | 30 |
| 34 | Champion's | TOHIT_DAMP | 126 | 150 | 40 | Weapon,Staff | Any | 2 | true | 15200 | 24000 | 33 |
| 35 | King's | TOHIT_DAMP | 151 | 175 | 28 | Weapon,Staff | Any | 2 | true | 24100 | 35000 | 38 |
| 36 | Vulnerable | ACP_CURSE | 51 | 100 | 3 | Armor,Shield | Any | 2 | false | 0 | 0 | -3 |
| 37 | Rusted | ACP_CURSE | 25 | 50 | 1 | Armor,Shield | Any | 2 | false | 0 | 0 | -2 |
| 38 | Fine | ACP | 20 | 30 | 1 | Armor,Shield | Any | 2 | true | 20 | 100 | 2 |
| 39 | Strong | ACP | 31 | 40 | 3 | Armor,Shield | Any | 2 | true | 120 | 200 | 3 |
| 40 | Grand | ACP | 41 | 55 | 6 | Armor,Shield | Any | 2 | true | 220 | 300 | 5 |
| 41 | Valiant | ACP | 56 | 70 | 10 | Armor,Shield | Any | 2 | true | 320 | 400 | 7 |
| 42 | Glorious | ACP | 71 | 90 | 14 | Armor,Shield | Good | 2 | true | 420 | 600 | 9 |
#### prefixes rows 43–83 of 83
| # | name | power | v1 | v2 | lvl | itemTypes | align | chance | ok | minVal | maxVal | multVal |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 43 | Blessed | ACP | 91 | 110 | 19 | Armor,Shield | Good | 2 | true | 620 | 800 | 11 |
| 44 | Saintly | ACP | 111 | 130 | 24 | Armor,Shield | Good | 2 | true | 820 | 1200 | 13 |
| 45 | Awesome | ACP | 131 | 150 | 28 | Armor,Shield | Good | 2 | true | 1220 | 2000 | 15 |
| 46 | Holy | ACP | 151 | 170 | 35 | Armor,Shield | Good | 2 | true | 5200 | 6000 | 17 |
| 47 | Godly | ACP | 171 | 200 | 60 | Armor,Shield | Good | 2 | true | 6200 | 7000 | 20 |
| 48 | Red | FIRERES | 10 | 20 | 4 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 500 | 1500 | 2 |
| 49 | Crimson | FIRERES | 21 | 30 | 10 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 2100 | 3000 | 2 |
| 50 | Crimson | FIRERES | 31 | 40 | 16 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 3100 | 4000 | 2 |
| 51 | Garnet | FIRERES | 41 | 50 | 20 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 8200 | 12000 | 3 |
| 52 | Ruby | FIRERES | 51 | 60 | 26 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 17100 | 20000 | 5 |
| 53 | Blue | LIGHTRES | 10 | 20 | 4 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 500 | 1500 | 2 |
| 54 | Azure | LIGHTRES | 21 | 30 | 10 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 2100 | 3000 | 2 |
| 55 | Lapis | LIGHTRES | 31 | 40 | 16 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 3100 | 4000 | 2 |
| 56 | Cobalt | LIGHTRES | 41 | 50 | 20 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 8200 | 12000 | 3 |
| 57 | Sapphire | LIGHTRES | 51 | 60 | 26 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 17100 | 20000 | 5 |
| 58 | White | MAGICRES | 10 | 20 | 4 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 500 | 1500 | 2 |
| 59 | Pearl | MAGICRES | 21 | 30 | 10 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 2100 | 3000 | 2 |
| 60 | Ivory | MAGICRES | 31 | 40 | 16 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 3100 | 4000 | 2 |
| 61 | Crystal | MAGICRES | 41 | 50 | 20 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 8200 | 12000 | 3 |
| 62 | Diamond | MAGICRES | 51 | 60 | 26 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 17100 | 20000 | 5 |
| 63 | Topaz | ALLRES | 10 | 15 | 8 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 2000 | 5000 | 3 |
| 64 | Amber | ALLRES | 16 | 20 | 12 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 7400 | 10000 | 3 |
| 65 | Jade | ALLRES | 21 | 30 | 18 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 11000 | 15000 | 3 |
| 66 | Obsidian | ALLRES | 31 | 40 | 24 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 24000 | 40000 | 4 |
| 67 | Emerald | ALLRES | 41 | 50 | 31 | Shield,Weapon,Staff,Bow | Any | 1 | true | 61000 | 75000 | 7 |
| 68 | Hyena's | MANA_CURSE | 11 | 25 | 4 | Staff,Misc | Any | 1 | false | 100 | 1000 | -2 |
| 69 | Frog's | MANA_CURSE | 1 | 10 | 1 | Staff,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 70 | Spider's | MANA | 10 | 15 | 1 | Staff,Misc | Evil | 1 | true | 500 | 1000 | 2 |
| 71 | Raven's | MANA | 15 | 20 | 5 | Staff,Misc | Any | 1 | true | 1100 | 2000 | 3 |
| 72 | Snake's | MANA | 21 | 30 | 9 | Staff,Misc | Any | 1 | true | 2100 | 4000 | 5 |
| 73 | Serpent's | MANA | 30 | 40 | 15 | Staff,Misc | Any | 1 | true | 4100 | 6000 | 7 |
| 74 | Drake's | MANA | 41 | 50 | 21 | Staff,Misc | Any | 1 | true | 6100 | 10000 | 9 |
| 75 | Dragon's | MANA | 51 | 60 | 27 | Staff,Misc | Any | 1 | true | 10100 | 15000 | 11 |
| 76 | Wyrm's | MANA | 61 | 80 | 35 | Staff | Any | 1 | true | 15100 | 19000 | 12 |
| 77 | Hydra's | MANA | 81 | 100 | 60 | Staff | Any | 1 | true | 19100 | 30000 | 13 |
| 78 | Angel's | SPLLVLADD | 1 | 1 | 15 | Staff | Good | 1 | true | 25000 | 25000 | 2 |
| 79 | Arch-Angel's | SPLLVLADD | 2 | 2 | 25 | Staff | Good | 1 | true | 50000 | 50000 | 3 |
| 80 | Plentiful | CHARGES | 2 | 2 | 4 | Staff | Any | 1 | true | 2000 | 2000 | 2 |
| 81 | Bountiful | CHARGES | 3 | 3 | 9 | Staff | Any | 1 | true | 3000 | 3000 | 3 |
| 82 | Flaming | FIREDAM | 1 | 10 | 7 | Weapon,Staff | Any | 1 | true | 5000 | 5000 | 2 |
| 83 | Lightning | LIGHTDAM | 2 | 20 | 18 | Weapon,Staff | Any | 1 | true | 10000 | 10000 | 2 |
### V1.4 Suffixes — item_suffixes.tsv rows 2-96 (95 rows, complete)
#### suffixes rows 1–42 of 95
| # | name | power | v1 | v2 | lvl | itemTypes | align | chance | ok | minVal | maxVal | multVal |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | quality | DAMMOD | 1 | 2 | 2 | Weapon,Bow | Any | 1 | true | 100 | 200 | 2 |
| 2 | maiming | DAMMOD | 3 | 5 | 7 | Weapon,Bow | Any | 1 | true | 1300 | 1500 | 3 |
| 3 | slaying | DAMMOD | 6 | 8 | 15 | Weapon | Any | 1 | true | 2600 | 3000 | 5 |
| 4 | gore | DAMMOD | 9 | 12 | 25 | Weapon | Any | 1 | true | 4100 | 5000 | 8 |
| 5 | carnage | DAMMOD | 13 | 16 | 35 | Weapon | Any | 1 | true | 5100 | 10000 | 10 |
| 6 | slaughter | DAMMOD | 17 | 20 | 60 | Weapon | Any | 1 | true | 10100 | 15000 | 13 |
| 7 | pain | GETHIT_CURSE | 2 | 4 | 4 | Armor,Shield,Misc | Evil | 1 | false | 0 | 0 | -4 |
| 8 | tears | GETHIT_CURSE | 1 | 1 | 2 | Armor,Shield,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 9 | health | GETHIT | 1 | 1 | 2 | Armor,Shield,Misc | Good | 1 | true | 200 | 200 | 2 |
| 10 | protection | GETHIT | 2 | 2 | 6 | Armor,Shield | Good | 1 | true | 400 | 800 | 4 |
| 11 | absorption | GETHIT | 3 | 3 | 12 | Armor,Shield | Good | 1 | true | 1001 | 2500 | 10 |
| 12 | deflection | GETHIT | 4 | 4 | 20 | Armor | Good | 1 | true | 2500 | 6500 | 15 |
| 13 | osmosis | GETHIT | 5 | 6 | 50 | Armor | Good | 1 | true | 7500 | 10000 | 20 |
| 14 | frailty | STR_CURSE | 6 | 10 | 3 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -3 |
| 15 | weakness | STR_CURSE | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 16 | strength | STR | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 200 | 1000 | 2 |
| 17 | might | STR | 6 | 10 | 5 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 1200 | 2000 | 3 |
| 18 | power | STR | 11 | 15 | 11 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 2200 | 3000 | 4 |
| 19 | giants | STR | 16 | 20 | 17 | Armor,Weapon,Bow,Misc | Any | 1 | true | 3200 | 5000 | 7 |
| 20 | titans | STR | 21 | 30 | 23 | Weapon,Misc | Any | 1 | true | 5200 | 10000 | 10 |
| 21 | paralysis | DEX_CURSE | 6 | 10 | 3 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -3 |
| 22 | atrophy | DEX_CURSE | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 23 | dexterity | DEX | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 200 | 1000 | 2 |
| 24 | skill | DEX | 6 | 10 | 5 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 1200 | 2000 | 3 |
| 25 | accuracy | DEX | 11 | 15 | 11 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 2200 | 3000 | 4 |
| 26 | precision | DEX | 16 | 20 | 17 | Armor,Weapon,Bow,Misc | Any | 1 | true | 3200 | 5000 | 7 |
| 27 | perfection | DEX | 21 | 30 | 23 | Bow,Misc | Any | 1 | true | 5200 | 10000 | 10 |
| 28 | the fool | MAG_CURSE | 6 | 10 | 3 | Armor,Shield,Weapon,Staff,Bow,Misc | Evil | 1 | false | 0 | 0 | -3 |
| 29 | dyslexia | MAG_CURSE | 1 | 5 | 1 | Armor,Shield,Weapon,Staff,Bow,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 30 | magic | MAG | 1 | 5 | 1 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 200 | 1000 | 2 |
| 31 | the mind | MAG | 6 | 10 | 5 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 1200 | 2000 | 3 |
| 32 | brilliance | MAG | 11 | 15 | 11 | Armor,Shield,Weapon,Staff,Bow,Misc | Any | 1 | true | 2200 | 3000 | 4 |
| 33 | sorcery | MAG | 16 | 20 | 17 | Armor,Weapon,Staff,Bow,Misc | Any | 1 | true | 3200 | 5000 | 7 |
| 34 | wizardry | MAG | 21 | 30 | 23 | Staff,Misc | Any | 1 | true | 5200 | 10000 | 10 |
| 35 | illness | VIT_CURSE | 6 | 10 | 3 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -3 |
| 36 | disease | VIT_CURSE | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 37 | vitality | VIT | 1 | 5 | 1 | Armor,Shield,Weapon,Bow,Misc | Good | 1 | true | 200 | 1000 | 2 |
| 38 | zest | VIT | 6 | 10 | 5 | Armor,Shield,Weapon,Bow,Misc | Good | 1 | true | 1200 | 2000 | 3 |
| 39 | vim | VIT | 11 | 15 | 11 | Armor,Shield,Weapon,Bow,Misc | Good | 1 | true | 2200 | 3000 | 4 |
| 40 | vigor | VIT | 16 | 20 | 17 | Armor,Weapon,Bow,Misc | Good | 1 | true | 3200 | 5000 | 7 |
| 41 | life | VIT | 21 | 30 | 23 | Misc | Good | 1 | true | 5200 | 10000 | 10 |
| 42 | trouble | ATTRIBS_CURSE | 6 | 10 | 12 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -10 |
#### suffixes rows 43–84 of 95
| # | name | power | v1 | v2 | lvl | itemTypes | align | chance | ok | minVal | maxVal | multVal |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 43 | the pit | ATTRIBS_CURSE | 1 | 5 | 5 | Armor,Shield,Weapon,Bow,Misc | Evil | 1 | false | 0 | 0 | -5 |
| 44 | the sky | ATTRIBS | 1 | 3 | 5 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 800 | 4000 | 5 |
| 45 | the moon | ATTRIBS | 4 | 7 | 11 | Armor,Shield,Weapon,Bow,Misc | Any | 1 | true | 4800 | 8000 | 10 |
| 46 | the stars | ATTRIBS | 8 | 11 | 17 | Armor,Weapon,Bow,Misc | Any | 1 | true | 8800 | 12000 | 15 |
| 47 | the heavens | ATTRIBS | 12 | 15 | 25 | Weapon,Bow,Misc | Any | 1 | true | 12800 | 20000 | 20 |
| 48 | the zodiac | ATTRIBS | 16 | 20 | 30 | Misc | Any | 1 | true | 20800 | 40000 | 30 |
| 49 | the vulture | LIFE_CURSE | 11 | 25 | 4 | Armor,Shield,Misc | Evil | 1 | false | 0 | 0 | -4 |
| 50 | the jackal | LIFE_CURSE | 1 | 10 | 1 | Armor,Shield,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 51 | the fox | LIFE | 10 | 15 | 1 | Armor,Shield,Misc | Any | 1 | true | 100 | 1000 | 2 |
| 52 | the jaguar | LIFE | 16 | 20 | 5 | Armor,Shield,Misc | Any | 1 | true | 1100 | 2000 | 3 |
| 53 | the eagle | LIFE | 21 | 30 | 9 | Armor,Shield,Misc | Any | 1 | true | 2100 | 4000 | 5 |
| 54 | the wolf | LIFE | 30 | 40 | 15 | Armor,Shield,Misc | Any | 1 | true | 4100 | 6000 | 7 |
| 55 | the tiger | LIFE | 41 | 50 | 21 | Armor,Shield,Misc | Any | 1 | true | 6100 | 10000 | 9 |
| 56 | the lion | LIFE | 51 | 60 | 27 | Armor,Misc | Any | 1 | true | 10100 | 15000 | 11 |
| 57 | the mammoth | LIFE | 61 | 80 | 35 | Armor | Any | 1 | true | 15100 | 19000 | 12 |
| 58 | the whale | LIFE | 81 | 100 | 60 | Armor | Any | 1 | true | 19100 | 30000 | 13 |
| 59 | fragility | DUR_CURSE | 100 | 100 | 3 | Armor,Shield,Weapon | Evil | 1 | false | 0 | 0 | -4 |
| 60 | brittleness | DUR_CURSE | 26 | 75 | 1 | Armor,Shield,Weapon | Evil | 1 | false | 0 | 0 | -2 |
| 61 | sturdiness | DUR | 26 | 75 | 1 | Armor,Shield,Weapon | Any | 1 | true | 100 | 100 | 2 |
| 62 | craftsmanship | DUR | 51 | 100 | 6 | Armor,Shield,Weapon | Any | 1 | true | 200 | 200 | 2 |
| 63 | structure | DUR | 101 | 200 | 12 | Armor,Shield,Weapon | Any | 1 | true | 300 | 300 | 2 |
| 64 | the ages | INDESTRUCTIBLE | - | - | 25 | Armor,Shield,Weapon | Any | 1 | true | 600 | 600 | 5 |
| 65 | the dark | LIGHT_CURSE | 4 | 4 | 6 | Armor,Weapon,Misc | Evil | 1 | false | 0 | 0 | -3 |
| 66 | the night | LIGHT_CURSE | 2 | 2 | 3 | Armor,Weapon,Misc | Evil | 1 | false | 0 | 0 | -2 |
| 67 | light | LIGHT | 2 | 2 | 4 | Armor,Weapon,Misc | Good | 1 | true | 750 | 750 | 2 |
| 68 | radiance | LIGHT | 4 | 4 | 8 | Armor,Weapon,Misc | Good | 1 | true | 1500 | 1500 | 3 |
| 69 | flame | FIRE_ARROWS | 1 | 3 | 1 | Bow | Any | 1 | true | 2000 | 2000 | 2 |
| 70 | fire | FIRE_ARROWS | 1 | 6 | 11 | Bow | Any | 1 | true | 4000 | 4000 | 4 |
| 71 | burning | FIRE_ARROWS | 1 | 16 | 35 | Bow | Any | 1 | true | 6000 | 6000 | 6 |
| 72 | shock | LIGHT_ARROWS | 1 | 6 | 13 | Bow | Any | 1 | true | 6000 | 6000 | 2 |
| 73 | lightning | LIGHT_ARROWS | 1 | 10 | 21 | Bow | Any | 1 | true | 8000 | 8000 | 4 |
| 74 | thunder | LIGHT_ARROWS | 1 | 20 | 60 | Bow | Any | 1 | true | 12000 | 12000 | 6 |
| 75 | many | DUR | 100 | 100 | 3 | Bow | Any | 1 | true | 750 | 750 | 2 |
| 76 | plenty | DUR | 200 | 200 | 7 | Bow | Any | 1 | true | 1500 | 1500 | 3 |
| 77 | thorns | THORNS | 1 | 3 | 1 | Armor,Shield | Any | 1 | true | 500 | 500 | 2 |
| 78 | corruption | NOMANA | - | - | 5 | Armor,Shield,Weapon | Evil | 1 | false | -1000 | -1000 | 2 |
| 79 | thieves | ABSHALFTRAP | - | - | 11 | Armor,Shield,Misc | Any | 1 | true | 1500 | 1500 | 2 |
| 80 | the bear | KNOCKBACK | - | - | 5 | Weapon,Staff,Bow | Evil | 1 | true | 750 | 750 | 2 |
| 81 | the bat | STEALMANA | 3 | 3 | 8 | Weapon | Any | 1 | true | 7500 | 7500 | 3 |
| 82 | vampires | STEALMANA | 5 | 5 | 19 | Weapon | Any | 1 | true | 15000 | 15000 | 3 |
| 83 | the leech | STEALLIFE | 3 | 3 | 8 | Weapon | Any | 1 | true | 7500 | 7500 | 3 |
| 84 | blood | STEALLIFE | 5 | 5 | 19 | Weapon | Any | 1 | true | 15000 | 15000 | 3 |
#### suffixes rows 85–95 of 95
| # | name | power | v1 | v2 | lvl | itemTypes | align | chance | ok | minVal | maxVal | multVal |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 85 | piercing | TARGAC | 2 | 6 | 1 | Weapon,Bow | Any | 1 | true | 1000 | 1000 | 3 |
| 86 | puncturing | TARGAC | 4 | 12 | 9 | Weapon,Bow | Any | 1 | true | 2000 | 2000 | 6 |
| 87 | bashing | TARGAC | 8 | 24 | 17 | Weapon | Any | 1 | true | 4000 | 4000 | 12 |
| 88 | readiness | FASTATTACK | 1 | 1 | 1 | Weapon,Staff,Bow | Any | 1 | true | 2000 | 2000 | 2 |
| 89 | swiftness | FASTATTACK | 2 | 2 | 10 | Weapon,Staff,Bow | Any | 1 | true | 4000 | 4000 | 4 |
| 90 | speed | FASTATTACK | 3 | 3 | 19 | Weapon,Staff | Any | 1 | true | 8000 | 8000 | 8 |
| 91 | haste | FASTATTACK | 4 | 4 | 27 | Weapon,Staff | Any | 1 | true | 16000 | 16000 | 16 |
| 92 | balance | FASTRECOVER | 1 | 1 | 1 | Armor,Misc | Any | 1 | true | 2000 | 2000 | 2 |
| 93 | stability | FASTRECOVER | 2 | 2 | 10 | Armor,Misc | Any | 1 | true | 4000 | 4000 | 4 |
| 94 | harmony | FASTRECOVER | 3 | 3 | 20 | Armor,Misc | Any | 1 | true | 8000 | 8000 | 8 |
| 95 | blocking | FASTBLOCK | 1 | 1 | 5 | Shield | Any | 1 | true | 4000 | 4000 | 4 |
Power-id vocabulary across prefixes+suffixes (`IPL_*` / effect names): ABSHALFTRAP, ACP, ACP_CURSE, ALLRES, ATTRIBS, ATTRIBS_CURSE, CHARGES, DAMMOD, DAMP, DAMP_CURSE, DEX, DEX_CURSE, DUR, DUR_CURSE, FASTATTACK, FASTBLOCK, FASTRECOVER, FIREDAM, FIRERES, FIRE_ARROWS, GETHIT, GETHIT_CURSE, INDESTRUCTIBLE, KNOCKBACK, LIFE, LIFE_CURSE, LIGHT, LIGHTDAM, LIGHTRES, LIGHT_ARROWS, LIGHT_CURSE, MAG, MAGICRES, MAG_CURSE, MANA, MANA_CURSE, NOMANA, SPLLVLADD, STEALLIFE, STEALMANA, STR, STR_CURSE, TARGAC, THORNS, TOHIT, TOHIT_CURSE, TOHIT_DAMP, TOHIT_DAMP_CURSE, VIT, VIT_CURSE.
### V1.5 Uniques — unique_itemdat.tsv rows 2-91 (90 rows, complete)
Schema (`UniqueItem`, parsed `tables/itemdat.cpp:636-667`): `name | cursorGraphic | uniqueBaseItem |
minLevel (UIMinLvl) | value (UIValue) | power0..power5 (type + .value1/.value2)`; `UINumPL` is
derived by stopping at the first empty power cell (`:651-656`).
#### uniques rows 1–45 of 90
| # | name | uniqBase | lvl | value | cursor | powers (type v1:v2 …) |
|---|---|---|---|---|---|---|
| 1 | The Butcher's Cleaver | CLEAVER | 1 | 3650 | - | STR 10:10; SETDAM 4:24; SETDUR 10:10 |
| 2 | The Undead Crown | SKCROWN | 1 | 16650 | HELM_OF_SPIRITS | RNDSTEALLIFE -:-; SETAC 8:8 |
| 3 | Empyrean Band | INFRARING | 1 | 8000 | - | ATTRIBS 2:2; LIGHT 2:2; FASTRECOVER 1:1; ABSHALFTRAP -:- |
| 4 | Optic Amulet | OPTAMULET | 1 | 9750 | - | LIGHT 2:2; LIGHTRES 20:20; GETHIT 1:1; MAG 5:5 |
| 5 | Ring of Truth | TRING | 1 | 9100 | - | LIFE 10:10; GETHIT 1:1; ALLRES 10:10 |
| 6 | Harlequin Crest | HARCREST | 1 | 4000 | - | AC_CURSE 3:3; GETHIT 1:1; ATTRIBS 2:2; LIFE 7:7; MANA 7:7 |
| 7 | Veil of Steel | STEELVEIL | 1 | 63800 | - | ALLRES 50:50; LIGHT_CURSE 2:2; ACP 60:60; MANA_CURSE 30:30; STR 15:15; VIT 15:15 |
| 8 | Arkaine's Valor | ARMOFVAL | 1 | 42000 | - | SETAC 25:25; VIT 10:10; GETHIT 3:3; FASTRECOVER 3:3 |
| 9 | Griswold's Edge | GRISWOLD | 1 | 42000 | - | FIREDAM 1:10; TOHIT 25:25; FASTATTACK 2:2; KNOCKBACK -:-; MANA 20:20; LIFE_CURSE 20:20 |
| 10 | Lightforge | LGTFORGE | 1 | 26675 | - | LIGHT 4:4; DAMP 150:150; TOHIT 25:25; FIREDAM 10:20; INDESTRUCTIBLE -:-; ATTRIBS 8:8 |
| 11 | The Rift Bow | SHORTBOW | 1 | 1800 | - | RNDARROWVEL -:-; DAMMOD 2:2; DEX_CURSE 3:3 |
| 12 | The Needler | SHORTBOW | 2 | 8900 | THE_NEEDLER | TOHIT 50:50; SETDAM 1:3; FASTATTACK 2:2 |
| 13 | The Celestial Bow | LONGBOW | 2 | 1200 | COMPOSITE_BOW | NOMINSTR -:-; DAMMOD 2:2; SETAC 5:5 |
| 14 | Deadly Hunter | COMPBOW | 3 | 8750 | DEADLY_HUNTER | 3XDAMVDEM -:-; TOHIT 20:20; MAG_CURSE 5:5 |
| 15 | Bow of the Dead | COMPBOW | 5 | 2500 | DEADLY_HUNTER | TOHIT 10:10; DEX 4:4; VIT_CURSE 3:3; LIGHT_CURSE 2:2; SETDUR 30:30 |
| 16 | The Blackoak Bow | LONGBOW | 5 | 2500 | - | DEX 10:10; VIT_CURSE 10:10; DAMP 50:50; LIGHT_CURSE 1:1 |
| 17 | Flamedart | HUNTBOW | 10 | 14250 | - | FIRE_ARROWS 1:6; TOHIT 20:20; FIRERES 40:40 |
| 18 | Fleshstinger | LONGBOW | 13 | 16500 | - | DEX 15:15; TOHIT 40:40; DAMP 80:80; DUR 6:6 |
| 19 | Windforce | WARBOW | 17 | 37750 | WINDFORCE | STR 5:5; DAMP 200:200; KNOCKBACK -:- |
| 20 | Eaglehorn | BATTLEBOW | 26 | 42500 | DEADLY_HUNTER | DEX 20:20; TOHIT 50:50; DAMP 100:100; INDESTRUCTIBLE -:- |
| 21 | Gonnagal's Dirk | DAGGER | 1 | 7040 | GONNAGALS_DIRK | DEX_CURSE 5:5; DAMMOD 4:4; FASTATTACK 2:2; FIRERES 25:25 |
| 22 | The Defender | SABRE | 1 | 2000 | - | SETAC 5:5; VIT 5:5; TOHIT_CURSE 5:5 |
| 23 | Gryphon's Claw | FALCHION | 1 | 1000 | GRYPHONS_CLAW | DAMP 100:100; MAG_CURSE 2:2; DEX_CURSE 5:5 |
| 24 | Black Razor | DAGGER | 1 | 2000 | BLACK_RAZOR | DAMP 150:150; VIT 2:2; SETDUR 5:5 |
| 25 | Gibbous Moon | BROADSWR | 2 | 6660 | - | ATTRIBS 2:2; DAMP 25:25; MANA 15:15; LIGHT_CURSE 3:3 |
| 26 | Ice Shank | LONGSWR | 3 | 5250 | - | FIRERES 40:40; SETDUR 15:15; STR 5:10 |
| 27 | The Executioner's Blade | FALCHION | 3 | 7080 | THE_EXECUTIONERS_BLADE | DAMP 150:150; LIFE_CURSE 10:10; LIGHT_CURSE 1:1; DUR 200:200 |
| 28 | The Bonesaw | CLAYMORE | 6 | 4400 | - | DAMMOD 10:10; STR 10:10; MAG_CURSE 5:5; DEX_CURSE 5:5; LIFE 10:10; MANA_CURSE 10:10 |
| 29 | Shadowhawk | BROADSWR | 8 | 13750 | - | LIGHT_CURSE 2:2; STEALLIFE 5:5; TOHIT 15:15; ALLRES 5:5 |
| 30 | Wizardspike | DAGGER | 11 | 12920 | WIZARDSPIKE | MAG 15:15; MANA 35:35; TOHIT 25:25; ALLRES 15:15 |
| 31 | Lightsabre | SABRE | 13 | 19150 | - | LIGHT 2:2; LIGHTDAM 1:10; TOHIT 20:20; LIGHTRES 50:50 |
| 32 | The Falcon's Talon | SCIMITAR | 15 | 7867 | GRYPHONS_CLAW | FASTATTACK 4:4; TOHIT 20:20; DAMP_CURSE 33:33; DEX 10:10 |
| 33 | Inferno | LONGSWR | 17 | 34600 | - | FIREDAM 2:12; LIGHT 3:3; MANA 20:20; FIRERES 80:80 |
| 34 | Doombringer | BASTARDSWR | 19 | 18250 | - | TOHIT 25:25; DAMP 250:250; ATTRIBS_CURSE 5:5; LIFE_CURSE 25:25; LIGHT_CURSE 2:2 |
| 35 | The Grizzly | TWOHANDSWR | 23 | 50000 | THE_GRIZZLY | STR 20:20; VIT_CURSE 5:5; DAMP 200:200; KNOCKBACK -:-; DUR 100:100 |
| 36 | The Grandfather | GREATSWR | 27 | 119800 | THE_GRANDFATHER | ONEHAND -:-; ATTRIBS 5:5; TOHIT 20:20; DAMP 70:70; LIFE 20:20 |
| 37 | The Mangler | LARGEAXE | 2 | 2850 | AXE | DAMP 200:200; DEX_CURSE 5:5; MAG_CURSE 5:5; MANA_CURSE 10:10 |
| 38 | Sharp Beak | LARGEAXE | 2 | 2850 | GREAT_AXE | LIFE 20:20; MAG_CURSE 10:10; MANA_CURSE 10:10 |
| 39 | Bloodslayer | BROADAXE | 3 | 2500 | AXE | DAMP 100:100; 3XDAMVDEM -:-; ATTRIBS_CURSE 5:5; SPLLVLADD -1:-1 |
| 40 | The Celestial Axe | BATTLEAXE | 4 | 14100 | - | NOMINSTR -:-; TOHIT 15:15; LIFE 15:15; STR_CURSE 15:15 |
| 41 | Wicked Axe | LARGEAXE | 5 | 31150 | GREAT_AXE | TOHIT 30:30; DEX 10:10; VIT_CURSE 10:10; GETHIT 1:6; INDESTRUCTIBLE -:- |
| 42 | Stonecleaver | BROADAXE | 7 | 23900 | STONECLEAVER | LIFE 30:30; TOHIT 20:20; DAMP 50:50; LIGHTRES 40:40 |
| 43 | Aguinara's Hatchet | SMALLAXE | 12 | 24800 | - | SPLLVLADD 1:1; MAG 10:10; MAGICRES 80:80 |
| 44 | Hellslayer | BATTLEAXE | 15 | 26200 | - | STR 8:8; VIT 8:8; DAMP 100:100; LIFE 25:25; MANA_CURSE 25:25 |
| 45 | Messerschmidt's Reaver | GREATAXE | 25 | 58000 | MESSERSCHMIDTS_REAVER | DAMP 200:200; DAMMOD 15:15; ATTRIBS 5:5; LIFE_CURSE 50:50; FIREDAM 2:12 |
#### uniques rows 46–90 of 90
| # | name | uniqBase | lvl | value | cursor | powers (type v1:v2 …) |
|---|---|---|---|---|---|---|
| 46 | Crackrust | MACE | 1 | 11375 | - | ATTRIBS 2:2; INDESTRUCTIBLE -:-; ALLRES 15:15; DAMP 50:50; SPLLVLADD -1:-1 |
| 47 | Hammer of Jholm | MAUL | 1 | 8700 | - | SETDAM 4:10; INDESTRUCTIBLE -:-; STR 3:3; TOHIT 15:15 |
| 48 | Civerb's Cudgel | MACE | 1 | 2000 | - | 3XDAMVDEM -:-; DEX_CURSE 5:5; MAG_CURSE 2:2 |
| 49 | The Celestial Star | FLAIL | 2 | 7810 | - | NOMINSTR -:-; LIGHT 2:2; DAMMOD 10:10; AC_CURSE 8:8 |
| 50 | Baranar's Star | MORNSTAR | 5 | 6850 | - | TOHIT 12:12; DAMP 80:80; FASTATTACK 1:1; VIT 4:4; DEX_CURSE 4:4; SETDUR 60:60 |
| 51 | Gnarled Root | SPIKCLUB | 9 | 9820 | - | TOHIT 20:20; DAMP 300:300; DEX 10:10; MAG 5:5; ALLRES 10:10; AC_CURSE 10:10 |
| 52 | The Cranium Basher | MAUL | 12 | 36500 | - | DAMMOD 20:20; STR 15:15; INDESTRUCTIBLE -:-; MANA_CURSE 150:150; ALLRES 5:5 |
| 53 | Schaefer's Hammer | WARHAMMER | 16 | 56125 | - | DAMP_CURSE 100:100; LIGHTDAM 1:50; LIFE 50:50; TOHIT 30:30; LIGHTRES 80:80; LIGHT 1:1 |
| 54 | Dreamflange | MACE | 26 | 26450 | - | MAG 30:30; MANA 50:50; MAGICRES 50:50; LIGHT 2:2; SPLLVLADD 1:1 |
| 55 | Staff of Shadows | LONGSTAFF | 2 | 1250 | - | MAG_CURSE 10:10; TOHIT 10:10; DAMP 60:60; LIGHT_CURSE 2:2; FASTATTACK 1:1 |
| 56 | Immolator | LONGSTAFF | 4 | 3900 | - | FIRERES 20:20; FIREDAM 4:4; MANA 10:10; VIT_CURSE 5:5 |
| 57 | Storm Spire | WARSTAFF | 8 | 22500 | - | LIGHTRES 50:50; LIGHTDAM 2:8; STR 10:10; MAG_CURSE 10:10 |
| 58 | Gleamsong | SHORTSTAFF | 8 | 6520 | - | MANA 25:25; STR_CURSE 3:3; VIT_CURSE 3:3; SPELL 10:76 |
| 59 | Thundercall | COMPSTAFF | 14 | 22250 | - | TOHIT 35:35; LIGHTDAM 1:10; SPELL 3:76; LIGHTRES 30:30; LIGHT 2:2 |
| 60 | The Protector | SHORTSTAFF | 16 | 17240 | THE_PROTECTOR | VIT 5:5; GETHIT 5:5; SETAC 40:40; SPELL 2:86; THORNS 1:3 |
| 61 | Naj's Puzzler | LONGSTAFF | 18 | 34000 | - | MAG 20:20; DEX 10:10; ALLRES 20:20; SPELL 23:57; LIFE_CURSE 25:25 |
| 62 | Mindcry | QUARSTAFF | 20 | 41500 | - | MAG 15:15; SPELL 13:69; ALLRES 15:15; SPLLVLADD 1:1 |
| 63 | Rod of Onan | WARSTAFF | 22 | 44167 | - | SPELL 21:50; DAMP 100:100; ATTRIBS 5:5 |
| 64 | Helm of Spirits | HELM | 1 | 7525 | HELM_OF_SPIRITS | STEALLIFE 5:5 |
| 65 | Thinking Cap | SKULLCAP | 6 | 2020 | THINKING_CAP | MANA 30:30; SPLLVLADD 2:2; ALLRES 20:20; SETDUR 1:1 |
| 66 | OverLord's Helm | HELM | 7 | 12500 | OVERLORDS_HELM | STR 20:20; DEX 15:15; VIT 5:5; MAG_CURSE 20:20; SETDUR 15:15 |
| 67 | Fool's Crest | HELM | 12 | 10150 | FOOLS_CREST | ATTRIBS_CURSE 4:4; LIFE 100:100; GETHIT_CURSE 1:6; THORNS 1:3 |
| 68 | Gotterdamerung | GREATHELM | 21 | 54900 | VEIL_OF_STEEL | ATTRIBS 20:20; SETAC 60:60; GETHIT 4:4; ALLRESZERO -:-; LIGHT_CURSE 4:4 |
| 69 | Royal Circlet | CROWN | 27 | 24875 | ROYAL_CIRCLET | ATTRIBS 10:10; MANA 40:40; SETAC 40:40; LIGHT 1:1 |
| 70 | Torn Flesh of Souls | RAGS | 2 | 4825 | TORN_FLESH_OF_SOULS | SETAC 8:8; VIT 10:10; GETHIT 1:1; INDESTRUCTIBLE -:- |
| 71 | The Gladiator's Bane | STUDARMOR | 6 | 3450 | - | SETAC 25:25; GETHIT 2:2; DUR 200:200; ATTRIBS_CURSE 3:3 |
| 72 | The Rainbow Cloak | CLOAK | 2 | 4900 | THE_RAINBOW_CLOAK | SETAC 10:10; ATTRIBS 1:1; ALLRES 10:10; LIFE 5:5; DUR 50:50 |
| 73 | Leather of Aut | LEATHARMOR | 4 | 10550 | - | SETAC 15:15; STR 5:5; MAG_CURSE 5:5; DEX 5:5; INDESTRUCTIBLE -:- |
| 74 | Wisdom's Wrap | ROBE | 5 | 6200 | THE_RAINBOW_CLOAK | MAG 5:5; MANA 10:10; LIGHTRES 25:25; SETAC 15:15; GETHIT 1:1 |
| 75 | Sparking Mail | CHAINMAIL | 9 | 15750 | - | SETAC 30:30; LIGHTDAM 1:10 |
| 76 | Scavenger Carapace | BREASTPLATE | 13 | 14000 | - | GETHIT 15:15; AC_CURSE 30:30; DEX 5:5; LIGHTRES 40:40 |
| 77 | Nightscape | CAPE | 16 | 11600 | THE_RAINBOW_CLOAK | FASTRECOVER 2:2; LIGHT_CURSE 4:4; SETAC 15:15; DEX 3:3; ALLRES 20:20 |
| 78 | Naj's Light Plate | PLATEMAIL | 19 | 78700 | NAJS_LIGHT_PLATE | NOMINSTR -:-; MAG 5:5; MANA 20:20; ALLRES 20:20; SPLLVLADD 1:1 |
| 79 | Demonspike Coat | FULLPLATE | 25 | 251175 | - | SETAC 100:100; GETHIT 6:6; STR 10:10; INDESTRUCTIBLE -:-; FIRERES 50:50 |
| 80 | The Deflector | BUCKLER | 1 | 1500 | - | SETAC 7:7; ALLRES 10:10; DAMP_CURSE 20:20; TOHIT_CURSE 5:5 |
| 81 | Split Skull Shield | BUCKLER | 1 | 2025 | SPLIT_SKULL_SHIELD | SETAC 10:10; LIFE 10:10; STR 2:2; LIGHT_CURSE 1:1; SETDUR 15:15 |
| 82 | Dragon's Breach | KITESHIELD | 2 | 19200 | DRAGONS_BREACH | FIRERES 25:25; STR 5:5; SETAC 20:20; MAG_CURSE 5:5; INDESTRUCTIBLE -:- |
| 83 | Blackoak Shield | SMALLSHIELD | 4 | 5725 | BLACKOAK_SHIELD | DEX 10:10; VIT_CURSE 10:10; SETAC 18:18; LIGHT_CURSE 1:1; DUR 150:150 |
| 84 | Holy Defender | LARGESHIELD | 10 | 13800 | BLACKOAK_SHIELD | SETAC 15:15; GETHIT 2:2; FIRERES 20:20; DUR 200:200; FASTBLOCK 1:1 |
| 85 | Stormshield | GOTHSHIELD | 24 | 49000 | GOTHIC_SHIELD | SETAC 40:40; GETHIT_CURSE 4:4; STR 10:10; INDESTRUCTIBLE -:-; FASTBLOCK 1:1; LIGHTRES 50:50 |
| 86 | Bramble | RING | 1 | 1000 | BRAMBLE | ATTRIBS_CURSE 2:2; DAMMOD 3:3; MANA 10:10 |
| 87 | Ring of Regha | RING | 1 | 4175 | RING_OF_REGHA | MAG 10:10; MAGICRES 10:10; LIGHT 1:1; STR_CURSE 3:3; DEX_CURSE 3:3 |
| 88 | The Bleeder | RING | 2 | 8500 | THE_BLEEDER | MAGICRES 20:20; MANA 30:30; LIFE_CURSE 10:10 |
| 89 | Constricting Ring | RING | 5 | 62000 | CONSTRICTING_RING | ALLRES 75:75; DRAINLIFE -:- |
| 90 | Ring of Engagement | RING | 11 | 12476 | RING_OF_ENGAGEMENT | GETHIT 1:2; THORNS 1:3; SETAC 5:5; TARGAC 2:2 |
First unique: `The Butcher's Cleaver` (base `CLEAVER`, lvl 1, 3650, STR 10:10 / SETDAM 4:24 /
SETDUR 10:10) — the drop is forced by `MT_CLEAVER` treasure `Uniq(CLEAVER)` (V2.1 row 53).
Last unique: `Ring of Engagement` (row 90, base `RING`).
### V1.6 Affix selection algorithm — the exact C++ (verified line by line)
1. **Eligible pool + weighting** — `SelectAffix(affixList, type, minlvl, maxlvl, onlygood, goe,`
   `excludeChargesForStaffs)` (`Source/items.cpp:1063-1093`): skip when (a) `type` not in
   `affix.PLIType` (`:1074`), (b) `affix.PLMinLvl < minlvl || > maxlvl` (`:1076`), (c)
   `onlygood && !affix.PLOk` (`:1078`), (d) good/evil exclusion `GOE_GOOD↔GOE_EVIL` (`:1080`),
   (e) `excludeChargesForStaffs && type==Staff && power==IPL_CHARGES` (`:1082`). **Then the affix
   is appended `PLChance` times** (`:1085-1087`) into a `StaticVector<const PLStruct*,256>`, and the
   pick is `eligibleAffixes[GenerateRnd(size)]` (`:1092`); empty pool ⇒ `nullopt` (`:1089-1090`).
   So `chance` is a pure integer weight over a duplicated list — replicate by materialising the
   same vector, or roll a cumulative-weight table **with identical tie order**.
2. **Prefix/suffix allocation + afflock** — `GetItemPowerPrefixAndSuffix` (`items.cpp:1179-1207`):
   `allocatePrefix = FlipCoin(4)` (25%), `allocateSuffix = !FlipCoin(3)` (66.67%); if neither, one
   is forced with `FlipCoin()` (`:1182-1187`); `if (!onlygood && !FlipCoin(3)) onlygood = true`
   (`:1189-1191`). Prefix is drawn first (`:1194-1199`) and **its `alignment` is written into the
   local `goe` before the suffix draw** (`:1195` → used at `:1201`) — this is the affix lock: an
   Evil prefix forbids every Good suffix and vice versa.
3. **Apply** — `SaveItemAffix` (`items.cpp:1048-1061`): `value = PLVal(value, power.param1,
   power.param2, affix.minVal, affix.maxVal)` (`:1053`) then `multVal` lands in `_iVMult2` when
   `param2` set else `_iVMult1` (`:1055-1059`); `GetItemPower` sets `_iMagical =
   ITEM_QUALITY_MAGIC` + `_iPrePower`/`_iSufPower` (`:1216-1224`).
4. **Name + value** — `GenerateMagicItemName` (`:1103-1113`, `"{0} {1} of {2}"` with short-name
   fallback via `StringInPanel`); `CalcItemValue` (`items.cpp:600-610`):
   `v = _iVMult1+_iVMult2; v>0 ⇒ v *= _ivalue; v<0 ⇒ v = _ivalue/v;`
   `_iIvalue = max(_iVAdd1+_iVAdd2+v, 1)`.
5. **Dispatch by item type** — `GetItemBonus` (`items.cpp:1309-1344`): `minlvl = min(minlvl, 25)`
   then Sword/Axe/Mace⇒Weapon, Bow⇒Bow, Shield⇒Shield, Light/Helm/Medium/Heavy⇒Armor,
   Staff⇒`GetStaffSpell` when `allowspells` (`:1236+`) else Staff affixes, Ring/Amulet⇒Misc,
   None/Misc/Gold⇒nothing.
**Worked example (weighting + afflock, all numbers from the real tables).**
Context: a Sword with `minlvl=1, maxlvl=6, onlygood=false, goe=GOE_ANY`, prefix draw succeeded.
- Prefix pool membership per rule 1 (`itemTypes` contains `Weapon`, `1 ≤ minLevel ≤ 6`):
  `Tin` (TOHIT_CURSE, chance **2**, alignment Evil), `Brass` (TOHIT_CURSE, chance **2**, Evil),
  `quality`-class weapon affixes with `minLevel ≤ 6` etc. Suppose the eligible set is exactly
  {Tin×2, Brass×2, `sharpness`×1, `mending`×1}: the vector is `[Tin,Tin,Brass,Brass,sharpness,mending]`
  of length 6, so `GenerateRnd(6)` gives Tin 2/6, Brass 2/6, each single-chance affix 1/6.
  Rolled index 0 ⇒ **Tin** (TOHIT_CURSE v1=6 v2=10, minVal 0 maxVal 0 multVal −3, alignment *Evil*).
- Afflock: `goe = Tin.PLGOE = GOE_EVIL` (`items.cpp:1195`).
- Suffix draw over the same level window now drops every suffix with `alignment=Good`
  (`items.cpp:1080`); e.g. `mayhem` (DAMMOD 9-13, lvl 7 — also filtered by the `minLevel>maxlvl`
  rule), leaving Evil/Any suffixes. If a `quality`-family Evil/Any suffix is drawn at index k,
  `SaveItemAffix` computes `_iVAdd*`/`_iVMult*` and `CalcItemValue` yields the final `_iIvalue`
  used by every vendor price in V5.3 (buy = `_iIvalue`, sell = `_iIvalue/4`).
- Any divergence in (a) duplicate-weight expansion, (b) `PLMinLvl` range direction, (c) the
  prefix→`goe` ordering, or (d) `FlipCoin` denominators (4/3/2) shifts the drop distribution.
### V1.7 Drop rules: magic / unique / no-drop (verified)
- **Entry** — `SpawnItem(monster, position, sendmsg, spawn)` (`items.cpp:3422-3481`):
  - `dropsSpecialTreasure = (monster.data().treasure & T_UNIQ) != 0` (`:3427`); treasure bit layout
    `tables/monstdat.h:86-88`: `T_MASK = 0xFFF`, `T_NODROP = 0x4000` (no loot), `T_UNIQ = 0x8000`
    (combine with the unique id, e.g. `Uniq(CLEAVER)` written by `monstdat.cpp:39`).
  - SP + special treasure ⇒ `SpawnUnique((_unique_items)(treasure & T_MASK), position)` (`:3431`, `:3195`)
    and return — **no normal drop**. (This is why the two monstdat rows with `Uniq(...)` always give
    their unique in single player; in MP-quest mode they fall through to the normal table.)
  - Unique monster or special treasure ⇒ base index = `RndUItem(monster)` (`:1376-1390`): allowed =
    books always, else `iMinMLvl <= monster->level(difficulty)` and **not** Gold/Misc.
  - Mushroom-quest brain override: `Q_MUSHROOM._qactive==ACTIVE && _qvar1==QS_MUSHGIVEN` ⇒ SP
    replaces the drop with `IDI_BRAIN` (`:3435-3443`); MP adds an extra `SpawnQuestItem(IDI_BRAIN,`
    `GetSuperItemLoc(pos))` so every client sees it (`:3446-3453`, comment explains the desync risk).
  - Normal monster: `T_NODROP` ⇒ return (`:3455`); else `onlygood = false` and
    `idx = RndItemForMonsterLevel(monster.level(difficulty))` (`:3459`).
  - **uper** (unique chance denominator): `uper = monster.isUnique() ? 15 : 1` (`:3464`).
  - Generation: `SetupAllItems(*MyPlayer, item, idx, AdvanceRndSeed(), mLevel, uper, onlygood, false)`
    (`:3468`) → `TryRandomUniqueItem` (`:3469`, `:3315-3360`) → `SetupItem` (`:3470`).
- **Base roll for normal monsters** — `RndItemForMonsterLevel(int8_t mLevel)` (`items.cpp:3256-3266`):
  `GenerateRnd(100) > 40 ⇒ IDI_NONE` (≈59% no drop), then `GenerateRnd(100) > 25 ⇒ IDI_GOLD`)
  (≈37.5% of drops are gold), else `GetItemIndexForDroppableItem(considerDropRate=true,`
  `iMinMLvl <= mLevel)` — a cumulative-weight table over `dropRate` rebuilt per call
  (`items.cpp:1333-1375`; `IsItemAvailable` filters spawn/Hellfire rows).
- **Magic-quality decision** — `GetItemBLevel(lvl, miscId, onlygood, uper15)` (`items.cpp:1480-1491`):
  `iblvl = lvl` if `GenerateRnd(100) <= 10` **or** `GenerateRnd(100) <= lvl` **or** `onlygood` **or**
  `miscId ∈ {STAFF, RING, AMULET}`; `uper15 ⇒ iblvl = lvl + 4`. `-1` ⇒ item stays normal.
- **Unique check** — `CheckUnique(item, lvl, uper, uidOffset)` (`items.cpp:1431-1453`):
  `GenerateRnd(100) > uper ⇒ UITEM_INVALID` (⇒ `uper=15`: 15/100 ≈ 15% for unique monsters;
  `uper=1`: 1%). Then `GetValidUniques(lvl, AllItemsList[item.IDidx].iItemId)` (`:1417-1429`: same
  base **and** `lvl >= UIMinLvl`), `DiscardRandomValues(1)` (`:1442`), take
  `validUniques[size-1-uidOffset]` (`:1447`) — **highest-lvl-eligible unique first**, `uidOffset`
  is how quest uniques are forced. `GetUniqueItem` (`:1455-1475`) applies up to 6 powers, sets
  `_iIName`, `_iCurs` when `UICurs != ICURS_DEFAULT`, `_iIvalue = UIValue`, `_iUid`,
  `_iMagical = UNIQUE`, `CF_UNIQUE`; for `IMISC_UNIQUE` items `_iSeed = uid`.
- **Already-dropped filter** — `TryRandomUniqueItem` (`items.cpp:3315-3360`): non-quest uniques
  re-roll by forcing the seed (`SetRndSeed(item._iSeed)`), filtering `UniqueItemFlags[]` in SP only
  (`:3330-3336`, MP allows repeats); if nothing remains it re-generates with `uper=1, mLevel+=4`.
- **Otherwise** magic: `GetItemBonus(item, iblvl/2, iblvl, onlygood, true)` (`items.cpp:3298`) then
  `ItemRndDur` (`:1475-1478`) unless the item became unique (`:3300-3303`).
- **Other generators** — `RndAllItems` (`:1392-1401`, >25% ⇒ gold), `RndTypeItems(type,imid,lvl)`
  (`:1403-1413`, `lvl*2` gate), `CreateRndItem`/`CreateRndUseful`/`CreateTypeItem` (`:3483-3516`),
  `SetupBaseItem` (`:1493+`), `SpawnQuestItem` (`:3700+`), `SpawnRewardItem`/`SpawnRuneBomb` (`:3724`).
- Quality tiers are NORMAL/MAGIC/UNIQUE only (`item_quality`, `Source/items.h:41-45`); there is no
  rare tier in this data set.
---
## V2. MONSTERS — 112 monstdat rows + 100 unique rows + AI families + pack rules
### V2.1 monstdat.tsv column list (41 columns, header rows 1; loader `tables/monstdat.cpp:409`)
`.tsv` header verbatim: `_monster_id, name, assetsSuffix, soundSuffix, trnFile, availability, width, image, hasSpecial, hasSpecialSound, frames[6], rate[6], minDunLvl, maxDunLvl, level, hitPointsMinimum, hitPointsMaximum, ai, abilityFlags, intelligence, toHit, animFrameNum, minDamage, maxDamage, toHitSpecial, animFrameNumSpecial, minDamageSpecial, maxDamageSpecial, reducePlayerStrength, reducePlayerMagic, reducePlayerDexterity, reducePlayerVitality, reducePlayerMaxHP, reducePlayerMaxMana, armorClass, monsterClass, resistance, resistanceHell, selectionRegion, treasure, exp`
Compressed presentation columns: `lvls`=minDunLvl-maxDunLvl, `hp`=hitPoints min-max,
`dmg`=min-max, `sdmg`=special min-max, `intel`=intelligence, `cls`=monsterClass,
`res`=resistance (+`resistanceHell` after `|H:`), `red`=stat reductions, `tr`=treasure,
`reg`=selectionRegion, `ai`=MonsterAIID.
#### monsters rows 1–38 of 112
| # | _monster_id | name | avail | lvls | lvl | hp | ac | toHit | dmg | sdmg | ai | intel | cls | res | red | tr | exp | reg |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | MT_NZOMBIE | Zombie | Always | 1-2 | 1 | 4-7 | 5 | 10 | 2-5 | - | Zombie | 0 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 54 | Bottom,Middle |
| 2 | MT_BZOMBIE | Ghoul | Always | 2-3 | 2 | 7-11 | 10 | 10 | 3-10 | - | Zombie | 1 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 58 | Bottom,Middle |
| 3 | MT_GZOMBIE | Rotting Carcass | Always | 2-4 | 4 | 15-25 | 15 | 25 | 5-15 | - | Zombie | 2 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC,RESIST_FIRE | - | - | 136 | Bottom,Middle |
| 4 | MT_YZOMBIE | Black Death | Always | 3-5 | 6 | 25-40 | 20 | 30 | 6-22 | - | Zombie | 3 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | HP1 | - | 240 | Bottom,Middle |
| 5 | MT_RFALLSP | Fallen One | Always | 1-2 | 1 | 1-4 | 0 | 15 | 1-3 | - | Fallen | 0 | Animal | - | - | - | 46 | Bottom,Middle |
| 6 | MT_DFALLSP | Carver | Always | 2-3 | 3 | 4-8 | 5 | 20 | 2-5 | - | Fallen | 2 | Animal | - | - | - | 80 | Bottom,Middle |
| 7 | MT_YFALLSP | Devil Kin | Always | 2-4 | 5 | 12-24 | 10 | 25 | 3-7 | - | Fallen | 2 | Animal | H:RESIST_FIRE | - | - | 155 | Bottom,Middle |
| 8 | MT_BFALLSP | Dark One | Always | 3-5 | 7 | 20-36 | 15 | 30 | 4-8 | - | Fallen | 3 | Animal | H:RESIST_LIGHTNING | - | - | 255 | Bottom,Middle |
| 9 | MT_WSKELAX | Skeleton | Always | 1-2 | 1 | 2-4 | 0 | 20 | 1-4 | - | SkeletonMelee | 0 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 64 | Bottom,Middle |
| 10 | MT_TSKELAX | Corpse Axe | Always | 2-3 | 2 | 4-7 | 0 | 25 | 3-5 | - | SkeletonMelee | 1 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 68 | Bottom,Middle |
| 11 | MT_RSKELAX | Burning Dead | Always | 2-4 | 4 | 8-12 | 5 | 30 | 3-7 | - | SkeletonMelee | 2 | Undead | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 154 | Bottom,Middle |
| 12 | MT_XSKELAX | Horror | Always | 3-5 | 6 | 12-20 | 15 | 35 | 4-9 | - | SkeletonMelee | 3 | Undead | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 264 | Bottom,Middle |
| 13 | MT_RFALLSD | Fallen One | Always | 1-2 | 1 | 2-5 | 10 | 15 | 1-4 | - | Fallen | 0 | Animal | - | - | - | 52 | Bottom,Middle |
| 14 | MT_DFALLSD | Carver | Always | 2-3 | 3 | 5-9 | 15 | 20 | 2-7 | - | Fallen | 1 | Animal | - | - | - | 90 | Bottom,Middle |
| 15 | MT_YFALLSD | Devil Kin | Always | 2-4 | 5 | 16-24 | 20 | 25 | 4-10 | - | Fallen | 2 | Animal | H:RESIST_FIRE | - | - | 180 | Bottom,Middle |
| 16 | MT_BFALLSD | Dark One | Always | 3-5 | 7 | 24-36 | 25 | 30 | 4-12 | - | Fallen | 3 | Animal | H:RESIST_LIGHTNING | - | - | 280 | Bottom,Middle |
| 17 | MT_NSCAV | Scavenger | Always | 1-3 | 2 | 3-6 | 10 | 20 | 1-5 | - | Scavenger | 0 | Animal | H:RESIST_FIRE | - | - | 80 | Bottom,Middle |
| 18 | MT_BSCAV | Plague Eater | Always | 2-4 | 4 | 12-24 | 20 | 30 | 1-8 | - | Scavenger | 1 | Animal | H:RESIST_LIGHTNING | - | - | 188 | Bottom,Middle |
| 19 | MT_WSCAV | Shadow Beast | Always | 3-5 | 6 | 24-36 | 25 | 35 | 3-12 | - | Scavenger | 2 | Animal | H:RESIST_FIRE | - | - | 375 | Bottom,Middle |
| 20 | MT_YSCAV | Bone Gasher | Always | 4-6 | 8 | 28-40 | 30 | 35 | 5-15 | - | Scavenger | 3 | Animal | RESIST_MAGIC \|H:RESIST_LIGHTNING | - | - | 552 | Bottom,Middle |
| 21 | MT_WSKELBW | Skeleton | Always | 2-3 | 3 | 2-4 | 0 | 15 | 1-2 | - | SkeletonRanged | 0 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 110 | Bottom,Middle |
| 22 | MT_TSKELBW | Corpse Bow | Always | 2-4 | 5 | 8-16 | 0 | 25 | 1-4 | - | SkeletonRanged | 1 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 210 | Bottom,Middle |
| 23 | MT_RSKELBW | Burning Dead | Always | 3-5 | 7 | 10-24 | 5 | 30 | 1-6 | - | SkeletonRanged | 2 | Undead | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 364 | Bottom,Middle |
| 24 | MT_XSKELBW | Horror | Always | 4-6 | 9 | 15-45 | 15 | 35 | 2-9 | - | SkeletonRanged | 3 | Undead | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 594 | Bottom,Middle |
| 25 | MT_WSKELSD | Skeleton Captain | Always | 1-3 | 2 | 3-6 | 10 | 20 | 2-7 | - | SkeletonMelee | 0 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 90 | Bottom,Middle |
| 26 | MT_TSKELSD | Corpse Captain | Always | 2-4 | 4 | 12-20 | 5 | 30 | 3-9 | - | SkeletonMelee | 1 | Undead | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 200 | Bottom,Middle |
| 27 | MT_RSKELSD | Burning Dead Captain | Always | 3-5 | 6 | 16-30 | 15 | 35 | 4-10 | - | SkeletonMelee | 2 | Undead | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 393 | Bottom,Middle |
| 28 | MT_XSKELSD | Horror Captain | Always | 4-6 | 8 | 35-50 | 30 | 40 | 5-14 | - | SkeletonMelee | 3 | Undead | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 604 | Bottom,Middle |
| 29 | MT_INVILORD | Invisible Lord | Never | 19-20 | 14 | 278-278 | 60 | 65 | 16-30 | - | SkeletonMelee | 3 | Demon | RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | - | - | 2000 | Bottom,Middle |
| 30 | MT_SNEAK | Hidden | Retail | 2-5 | 5 | 8-24 | 25 | 35 | 3-6 | - | Sneak | 0 | Demon | - | - | - | 278 | Bottom,Middle |
| 31 | MT_STALKER | Stalker | Retail | 5-7 | 9 | 30-45 | 30 | 40 | 8-16 | - | Sneak | 1 | Demon | - | - | - | 630 | Bottom,Middle |
| 32 | MT_UNSEEN | Unseen | Retail | 6-8 | 11 | 35-50 | 30 | 45 | 12-20 | - | Sneak | 2 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC | - | - | 935 | Bottom,Middle |
| 33 | MT_ILLWEAV | Illusion Weaver | Retail | 8-10 | 13 | 40-60 | 30 | 60 | 16-24 | - | Sneak | 3 | Demon | RESIST_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,RESIST_FIRE | - | - | 1500 | Bottom,Middle |
| 34 | MT_LRDSAYTR | Satyr Lord | Never | 21-22 | 28 | 160-200 | 70 | 90 | 20-30 | - | SkeletonMelee | 3 | Animal | RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 2800 | Bottom,Middle |
| 35 | MT_NGOATMC | Flesh Clan | Retail | 4-6 | 8 | 30-45 | 40 | 50 | 4-10 | - | GoatMelee | 0 | Demon | - | - | - | 460 | Bottom,Middle |
| 36 | MT_BGOATMC | Stone Clan | Retail | 5-7 | 10 | 40-55 | 40 | 60 | 6-12 | - | GoatMelee | 1 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC | - | - | 685 | Bottom,Middle |
| 37 | MT_RGOATMC | Fire Clan | Retail | 6-8 | 12 | 50-65 | 45 | 70 | 8-16 | - | GoatMelee | 2 | Demon | RESIST_FIRE \|H:IMMUNE_FIRE | - | - | 906 | Bottom,Middle |
| 38 | MT_GGOATMC | Night Clan | Retail | 7-9 | 14 | 55-70 | 50 | 80 | 10-20 | 30-30 | GoatMelee | 3 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC | - | - | 1190 | Bottom,Middle |
#### monsters rows 39–76 of 112
| # | _monster_id | name | avail | lvls | lvl | hp | ac | toHit | dmg | sdmg | ai | intel | cls | res | red | tr | exp | reg |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 39 | MT_FIEND | Fiend | Always | 2-3 | 3 | 3-6 | 0 | 35 | 1-6 | - | Bat | 0 | Animal | - | - | None | 102 | Middle,Top |
| 40 | MT_BLINK | Blink | Always | 3-5 | 7 | 12-28 | 15 | 45 | 1-8 | - | Bat | 1 | Animal | - | - | None | 340 | Middle,Top |
| 41 | MT_GLOOM | Gloom | Always | 4-6 | 9 | 28-36 | 35 | 70 | 4-12 | - | Bat | 2 | Animal | RESIST_MAGIC \|H:RESIST_MAGIC | - | None | 509 | Middle,Top |
| 42 | MT_FAMILIAR | Familiar | Always | 6-8 | 13 | 20-35 | 35 | 50 | 4-16 | - | Bat | 3 | Demon | RESIST_MAGIC,IMMUNE_LIGHTNING \|H:RESIST_MAGIC,IMMUNE_LIGHTNING | - | None | 448 | Middle,Top |
| 43 | MT_NGOATBW | Flesh Clan | Retail | 4-6 | 8 | 20-35 | 35 | 35 | 1-7 | - | GoatRanged | 0 | Demon | - | - | - | 448 | Bottom,Middle |
| 44 | MT_BGOATBW | Stone Clan | Retail | 5-7 | 10 | 30-40 | 35 | 40 | 2-9 | - | GoatRanged | 1 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC | - | - | 645 | Bottom,Middle |
| 45 | MT_RGOATBW | Fire Clan | Retail | 6-8 | 12 | 40-50 | 35 | 45 | 3-11 | - | GoatRanged | 2 | Demon | RESIST_FIRE \|H:IMMUNE_FIRE | - | - | 822 | Bottom,Middle |
| 46 | MT_GGOATBW | Night Clan | Retail | 7-9 | 14 | 50-65 | 40 | 50 | 4-13 | - | GoatRanged | 3 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC | - | - | 1092 | Bottom,Middle |
| 47 | MT_NACID | Acid Beast | Retail | 6-8 | 11 | 40-66 | 30 | 40 | 4-12 | - | Acid | 0 | Animal | IMMUNE_ACID \|H:IMMUNE_MAGIC,IMMUNE_ACID | - | - | 846 | Bottom,Middle |
| 48 | MT_RACID | Poison Spitter | Retail | 8-10 | 15 | 60-85 | 30 | 45 | 4-16 | - | Acid | 1 | Animal | IMMUNE_ACID \|H:IMMUNE_MAGIC,IMMUNE_ACID | - | - | 1248 | Bottom,Middle |
| 49 | MT_BACID | Pit Beast | Retail | 10-12 | 21 | 80-110 | 35 | 55 | 8-18 | - | Acid | 2 | Animal | RESIST_MAGIC,IMMUNE_ACID \|H:IMMUNE_MAGIC,RESIST_LIGHTNING,IMMUNE_ACID | - | - | 2060 | Bottom,Middle |
| 50 | MT_XACID | Lava Maw | Retail | 12-14 | 25 | 100-150 | 35 | 65 | 10-20 | - | Acid | 3 | Animal | RESIST_MAGIC,IMMUNE_FIRE,IMMUNE_ACID \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_ACID | - | - | 2940 | Bottom,Middle |
| 51 | MT_SKING | Skeleton King | Never | 4-4 | 9 | 140-140 | 70 | 60 | 6-16 | - | SkeletonKing | 3 | Undead | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | Uniq(SKCROWN) | 570 | Bottom,Middle,Top |
| 52 | MT_CLEAVER | The Butcher | Never | 1-1 | 1 | 320-320 | 50 | 50 | 6-12 | - | Butcher | 3 | Demon | RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | Uniq(CLEAVER) | 710 | Bottom,Middle |
| 53 | MT_FAT | Overlord | Retail | 5-7 | 10 | 60-80 | 55 | 55 | 6-12 | - | Fat | 0 | Demon | H:RESIST_FIRE | - | - | 635 | Bottom,Middle |
| 54 | MT_MUDMAN | Mud Man | Retail | 7-9 | 14 | 100-125 | 60 | 60 | 8-16 | - | Fat | 1 | Demon | H:IMMUNE_LIGHTNING | - | - | 1165 | Bottom,Middle |
| 55 | MT_TOAD | Toad Demon | Retail | 8-10 | 16 | 135-160 | 65 | 70 | 8-16 | 8-20 | Fat | 2 | Demon | IMMUNE_MAGIC \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 1380 | Bottom,Middle |
| 56 | MT_FLAYED | Flayed One | Retail | 10-12 | 20 | 160-200 | 70 | 85 | 10-20 | - | Fat | 3 | Demon | RESIST_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 2058 | Bottom,Middle |
| 57 | MT_WYRM | Wyrm | Never | 5-7 | 11 | 60-90 | 25 | 40 | 4-10 | - | SkeletonMelee | 0 | Animal | RESIST_MAGIC \|H:RESIST_MAGIC | - | - | 660 | Bottom,Middle |
| 58 | MT_CAVSLUG | Cave Slug | Never | 6-8 | 13 | 75-110 | 30 | 50 | 6-13 | - | SkeletonMelee | 1 | Animal | RESIST_MAGIC \|H:RESIST_MAGIC | - | - | 994 | Bottom,Middle |
| 59 | MT_DVLWYRM | Devil Wyrm | Never | 7-9 | 15 | 100-140 | 30 | 55 | 8-16 | - | SkeletonMelee | 2 | Animal | RESIST_MAGIC,RESIST_FIRE \|H:RESIST_MAGIC,RESIST_FIRE | - | - | 1320 | Bottom,Middle |
| 60 | MT_DEVOUR | Devourer | Never | 8-10 | 17 | 125-200 | 35 | 60 | 10-20 | - | SkeletonMelee | 3 | Animal | RESIST_MAGIC,RESIST_FIRE \|H:RESIST_MAGIC,RESIST_FIRE | - | - | 1827 | Bottom,Middle |
| 61 | MT_NMAGMA | Magma Demon | Retail | 8-9 | 13 | 50-70 | 45 | 45 | 2-10 | - | Magma | 0 | Demon | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 1076 | Bottom,Middle,Top |
| 62 | MT_YMAGMA | Blood Stone | Retail | 8-10 | 14 | 55-75 | 45 | 50 | 2-12 | - | Magma | 1 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 1309 | Bottom,Middle,Top |
| 63 | MT_BMAGMA | Hell Stone | Retail | 9-11 | 16 | 60-80 | 50 | 60 | 2-20 | - | Magma | 2 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 1680 | Bottom,Middle,Top |
| 64 | MT_WMAGMA | Lava Lord | Retail | 9-11 | 18 | 70-85 | 60 | 75 | 4-24 | - | Magma | 3 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 2124 | Bottom,Middle,Top |
| 65 | MT_HORNED | Horned Demon | Retail | 7-9 | 13 | 40-80 | 40 | 60 | 2-16 | 5-32 | Rhino | 0 | Animal | H:RESIST_FIRE | - | - | 1172 | Bottom,Middle,Top |
| 66 | MT_MUDRUN | Mud Runner | Retail | 8-10 | 15 | 50-90 | 45 | 70 | 6-18 | 12-36 | Rhino | 1 | Animal | H:RESIST_FIRE | - | - | 1404 | Bottom,Middle,Top |
| 67 | MT_FROSTC | Frost Charger | Retail | 9-11 | 17 | 60-100 | 50 | 80 | 8-20 | 20-40 | Rhino | 2 | Animal | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 1720 | Bottom,Middle,Top |
| 68 | MT_OBLORD | Obsidian Lord | Retail | 10-12 | 19 | 70-110 | 55 | 90 | 10-22 | 20-50 | Rhino | 3 | Animal | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 1809 | Bottom,Middle,Top |
| 69 | MT_BONEDMN | oldboned | Never | 24-24 | 12 | 70-70 | 50 | 60 | 6-14 | - | Storm | 0 | Demon | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 1344 | Bottom,Middle,Top |
| 70 | MT_REDDTH | Red Death | Never | 8-10 | 16 | 96-96 | 60 | 75 | 10-20 | - | Storm | 1 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 2168 | Bottom,Middle,Top |
| 71 | MT_LTCHDMN | Litch Demon | Never | 9-11 | 18 | 110-110 | 45 | 80 | 10-24 | - | Storm | 2 | Demon | IMMUNE_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 2736 | Bottom,Middle,Top |
| 72 | MT_UDEDBLRG | Undead Balrog | Never | 11-13 | 22 | 130-130 | 65 | 85 | 12-30 | - | Storm | 3 | Demon | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | - | - | 3575 | Bottom,Middle,Top |
| 73 | MT_INCIN | Incinerator | Never | 21-22 | 16 | 30-45 | 25 | 75 | 8-16 | - | FireMan | 0 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 1888 | Bottom,Middle |
| 74 | MT_FLAMLRD | Flame Lord | Never | 22-23 | 18 | 40-55 | 25 | 75 | 10-20 | - | FireMan | 1 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 2250 | Bottom,Middle |
| 75 | MT_DOOMFIRE | Doom Fire | Never | 23-24 | 20 | 50-65 | 30 | 80 | 12-24 | - | FireMan | 2 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 2740 | Bottom,Middle |
| 76 | MT_HELLBURN | Hell Burner | Never | 24-24 | 22 | 60-80 | 30 | 85 | 15-30 | - | FireMan | 3 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 3355 | Bottom,Middle |
#### monsters rows 77–112 of 112
| # | _monster_id | name | avail | lvls | lvl | hp | ac | toHit | dmg | sdmg | ai | intel | cls | res | red | tr | exp | reg |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 77 | MT_STORM | Red Storm | Retail | 9-11 | 18 | 55-110 | 30 | 80 | 8-18 | 4-16 | Storm | 0 | Demon | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 2160 | Bottom,Middle,Top |
| 78 | MT_RSTORM | Storm Rider | Retail | 10-12 | 20 | 60-120 | 30 | 80 | 8-18 | 4-16 | Storm | 1 | Demon | RESIST_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 2391 | Bottom,Middle,Top |
| 79 | MT_STORML | Storm Lord | Retail | 11-13 | 22 | 75-135 | 35 | 85 | 12-24 | 4-16 | Storm | 2 | Demon | RESIST_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 2775 | Bottom,Middle,Top |
| 80 | MT_MAEL | Maelstrom | Retail | 12-14 | 24 | 90-150 | 40 | 90 | 12-28 | 4-16 | Storm | 3 | Demon | RESIST_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 3177 | Bottom,Middle,Top |
| 81 | MT_BIGFALL | Devil Kin Brute | Never | 21-22 | 27 | 120-160 | 70 | 100 | 18-24 | - | SkeletonMelee | 3 | Animal | RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | - | - | 2400 | Bottom,Middle |
| 82 | MT_WINGED | Winged-Demon | Retail | 5-7 | 9 | 45-60 | 45 | 50 | 10-16 | - | Gargoyle | 0 | Demon | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 662 | Middle,Top |
| 83 | MT_GARGOYLE | Gargoyle | Retail | 7-9 | 13 | 60-90 | 45 | 65 | 10-16 | - | Gargoyle | 1 | Demon | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_LIGHTNING | - | - | 1205 | Middle,Top |
| 84 | MT_BLOODCLW | Blood Claw | Retail | 9-11 | 19 | 75-125 | 50 | 80 | 14-22 | - | Gargoyle | 2 | Demon | IMMUNE_MAGIC,IMMUNE_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 1873 | Middle,Top |
| 85 | MT_DEATHW | Death Wing | Retail | 10-12 | 23 | 90-150 | 60 | 95 | 16-28 | - | Gargoyle | 3 | Demon | IMMUNE_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | - | - | 2278 | Middle,Top |
| 86 | MT_MEGA | Slayer | Retail | 10-12 | 20 | 120-140 | 60 | 100 | 12-20 | - | Mega | 0 | Demon | RESIST_MAGIC,IMMUNE_FIRE \|H:RESIST_MAGIC,IMMUNE_FIRE | - | - | 2300 | Bottom,Middle,Top |
| 87 | MT_GUARD | Guardian | Retail | 11-13 | 22 | 140-160 | 65 | 110 | 14-22 | - | Mega | 1 | Demon | RESIST_MAGIC,IMMUNE_FIRE \|H:RESIST_MAGIC,IMMUNE_FIRE | - | - | 2714 | Bottom,Middle,Top |
| 88 | MT_VTEXLRD | Vortex Lord | Retail | 12-14 | 24 | 160-180 | 70 | 120 | 18-24 | - | Mega | 2 | Demon | RESIST_MAGIC,IMMUNE_FIRE \|H:RESIST_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 3252 | Bottom,Middle,Top |
| 89 | MT_BALROG | Balrog | Retail | 13-15 | 26 | 180-200 | 75 | 130 | 22-30 | - | Mega | 3 | Demon | RESIST_MAGIC,IMMUNE_FIRE \|H:RESIST_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 3643 | Bottom,Middle,Top |
| 90 | MT_NSNAKE | Cave Viper | Retail | 11-13 | 21 | 100-150 | 60 | 90 | 8-20 | - | Snake | 0 | Demon | IMMUNE_MAGIC \|H:IMMUNE_MAGIC | - | - | 2725 | Bottom,Middle,Top |
| 91 | MT_RSNAKE | Fire Drake | Retail | 12-14 | 23 | 120-170 | 65 | 105 | 12-24 | - | Snake | 1 | Demon | IMMUNE_MAGIC,RESIST_FIRE \|H:IMMUNE_MAGIC,IMMUNE_FIRE | - | - | 3139 | Bottom,Middle,Top |
| 92 | MT_BSNAKE | Gold Viper | Retail | 13-14 | 25 | 140-180 | 70 | 120 | 15-26 | - | Snake | 2 | Demon | IMMUNE_MAGIC,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 3540 | Bottom,Middle,Top |
| 93 | MT_GSNAKE | Azure Drake | Retail | 15-16 | 27 | 160-200 | 75 | 130 | 18-30 | - | Snake | 3 | Demon | RESIST_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | - | - | 3791 | Bottom,Middle,Top |
| 94 | MT_NBLACK | Black Knight | Retail | 12-14 | 24 | 150-150 | 75 | 110 | 15-20 | - | SkeletonMelee | 0 | Demon | RESIST_MAGIC,RESIST_LIGHTNING \|H:RESIST_MAGIC,IMMUNE_LIGHTNING | - | - | 3360 | Bottom,Middle,Top |
| 95 | MT_RTBLACK | Doom Guard | Retail | 13-15 | 26 | 165-165 | 75 | 130 | 18-25 | - | SkeletonMelee | 0 | Demon | RESIST_MAGIC,RESIST_FIRE \|H:RESIST_MAGIC,IMMUNE_FIRE | - | - | 3650 | Bottom,Middle,Top |
| 96 | MT_BTBLACK | Steel Lord | Retail | 14-16 | 28 | 180-180 | 80 | 120 | 20-30 | - | SkeletonMelee | 1 | Demon | RESIST_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 4252 | Bottom,Middle,Top |
| 97 | MT_RBLACK | Blood Knight | Retail | 13-14 | 30 | 200-200 | 85 | 130 | 25-35 | - | SkeletonMelee | 1 | Demon | IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | - | - | 5130 | Bottom,Middle,Top |
| 98 | MT_UNRAV | The Shredded | Never | 17-18 | 23 | 70-90 | 65 | 75 | 4-12 | - | SkeletonMelee | 0 | Undead | RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_FIRE,RESIST_LIGHTNING | - | - | 900 | Bottom,Middle |
| 99 | MT_HOLOWONE | Hollow One | Never | 18-19 | 27 | 135-240 | 75 | 75 | 12-24 | - | SkeletonMelee | 1 | Undead | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 4374 | Bottom,Middle |
| 100 | MT_PAINMSTR | Pain Master | Never | 19-20 | 29 | 110-200 | 80 | 80 | 16-30 | - | SkeletonMelee | 2 | Undead | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 5147 | Bottom,Middle |
| 101 | MT_REALWEAV | Reality Weaver | Never | 20-20 | 30 | 135-240 | 85 | 85 | 20-35 | - | SkeletonMelee | 3 | Undead | RESIST_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING \|H:RESIST_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 5925 | Bottom,Middle |
| 102 | MT_SUCCUBUS | Succubus | Retail | 12-14 | 24 | 120-150 | 60 | 100 | 1-20 | - | Succubus | 0 | Demon | RESIST_MAGIC \|H:IMMUNE_MAGIC,RESIST_FIRE | - | - | 3696 | Bottom,Middle |
| 103 | MT_SNOWWICH | Snow Witch | Retail | 13-15 | 26 | 135-175 | 65 | 110 | 1-24 | - | Succubus | 1 | Demon | RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_LIGHTNING | - | - | 4084 | Bottom,Middle |
| 104 | MT_HLSPWN | Hell Spawn | Retail | 14-16 | 28 | 150-200 | 75 | 115 | 1-30 | - | Succubus | 2 | Demon | RESIST_MAGIC,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 4480 | Bottom,Middle |
| 105 | MT_SOLBRNR | Soul Burner | Retail | 15-16 | 30 | 140-225 | 85 | 120 | 1-35 | - | Succubus | 3 | Demon | RESIST_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 4644 | Bottom,Middle |
| 106 | MT_COUNSLR | Counselor | Retail | 13-14 | 25 | 70-70 | 0 | 90 | 8-20 | - | Counselor | 0 | Demon | RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | - | - | 4070 | Bottom,Middle,Top |
| 107 | MT_MAGISTR | Magistrate | Retail | 14-15 | 27 | 85-85 | 0 | 100 | 10-24 | - | Counselor | 1 | Demon | RESIST_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | - | - | 4478 | Bottom,Middle,Top |
| 108 | MT_CABALIST | Cabalist | Retail | 15-16 | 29 | 120-120 | 0 | 110 | 14-30 | - | Counselor | 2 | Demon | RESIST_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | - | - | 4929 | Bottom,Middle,Top |
| 109 | MT_ADVOCATE | Advocate | Retail | 16-16 | 30 | 145-145 | 0 | 120 | 15-25 | - | Counselor | 3 | Demon | IMMUNE_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 4968 | Bottom,Middle,Top |
| 110 | MT_GOLEM | Golem | Never | 1-1 | 12 | 1-1 | 1 | 0 | 1-1 | - | Golem | 0 | Demon | - | - | - | 0 | - |
| 111 | MT_DIABLO | The Dark Lord | Never | 26-26 | 30 | 1666-1666 | 90 | 220 | 30-60 | - | Diablo | 3 | Demon | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | - | - | 31666 | Bottom,Middle,Top |
| 112 | MT_DARKMAGE | The Arch-Litch Malignus | Never | 21-21 | 30 | 160-160 | 70 | 120 | 20-40 | - | Counselor | 3 | Demon | RESIST_MAGIC,RESIST_FIRE,RESIST_LIGHTNING \|H:IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | - | - | 4968 | Bottom,Middle,Top |
Row 1 = `MT_NZOMBIE` Zombie; row 112 = `MT_DARKMAGE` The Arch-Litch Malignus (monstdat.tsv physical lines 2 and 113).
- **`frames[6]` / `rate[6]`**: 112 rows carry animation counts + tick rates for the six
  animation groups (neutral/walk/wait-attack-hit-dead ordering per `MonsterGraphic`,
  `MonsterData` fields, monstdat.h:100-160). Example row 1: `11,24,12,6,16,0` frames with
  `4,1,1,1,1,1` tick rates; `animFrameNum`=8 for the attack animation.
- **treasure** column vocabulary: `-` (normal table), `None` (=T_NODROP, `monstdat.cpp:304`)
  on 4 rows, and exactly 2 forced uniques: row 53 `MT_SKING` → `Uniq(SKCROWN)`,
  row 52 `MT_CLEAVER` → `Uniq(CLEAVER)`. Histogram: `(empty)`=106, `None`=4, `Uniq(SKCROWN)`=1, `Uniq(CLEAVER)`=1.
- **availability** histogram: Always=32, Never=24, Retail=56 — `MonsterAvailability` (`tables/monstdat.h:92-95`: Never/Retail/Always/Hellfire variants),
  filtered by `IsMonsterAvailable` (`monster.cpp:3159-3167`): `Never` always out, `Retail` out when
  `gbIsSpawn`, plus the level window `minDunLvl <= currlevel <= maxDunLvl`.
- **ai** histogram over 112 rows: SkeletonMelee=23, Fallen=8, Storm=8, Counselor=5, Zombie=4, Scavenger=4, SkeletonRanged=4, Sneak=4, GoatMelee=4, Bat=4, GoatRanged=4, Acid=4, Fat=4, Magma=4, Rhino=4, FireMan=4, Gargoyle=4, Mega=4, Snake=4, Succubus=4, SkeletonKing=1, Butcher=1, Golem=1, Diablo=1.
- **monsterClass** histogram: Animal=29, Demon=62, Undead=21 (`MonsterClass`, monstdat.h:66-70; UI text `GetMonsterTypeText` monster.cpp:3066-3078).
### V2.2 unique_monstdat.tsv rows 2-101 (100 rows, complete; loader `tables/monstdat.cpp:459`)
#### unique monsters rows 1–34 of 100
| # | name | base | trn | lvl | maxHp | ai | int | dmg | red | res | pack | toHit | ac | talk |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Gharbad the Weak | MT_NGOATMC | bsdb | 4 | 120 | Gharbad | 3 | 8-16 | - | IMMUNE_LIGHTNING | None | 0 | 0 | TEXT_GARBUD1 |
| 2 | Skeleton King | MT_SKING | genrl | 0 | 240 | SkeletonKing | 3 | 6-16 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Independent | 0 | 0 | - |
| 3 | Zhar the Mad | MT_COUNSLR | general | 8 | 360 | Zhar | 3 | 16-40 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | None | 0 | 0 | TEXT_ZHAR1 |
| 4 | Snotspill | MT_BFALLSP | bng | 4 | 220 | Snotspill | 3 | 10-18 | - | RESIST_LIGHTNING | None | 0 | 0 | TEXT_BANNER10 |
| 5 | Arch-Bishop Lazarus | MT_ADVOCATE | general | 0 | 600 | Lazarus | 3 | 30-50 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | None | 0 | 0 | TEXT_VILE13 |
| 6 | Red Vex | MT_HLSPWN | redv | 0 | 400 | LazarusSuccubus | 3 | 30-50 | - | IMMUNE_MAGIC,RESIST_FIRE | None | 0 | 0 | TEXT_VILE13 |
| 7 | Black Jade | MT_HLSPWN | blkjd | 0 | 400 | LazarusSuccubus | 3 | 30-50 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | None | 0 | 0 | TEXT_VILE13 |
| 8 | Lachdanan | MT_RBLACK | bhka | 14 | 500 | Lachdanan | 3 | 0-0 | - | - | None | 0 | 0 | TEXT_VEIL9 |
| 9 | Warlord of Blood | MT_BTBLACK | general | 13 | 850 | Warlord | 3 | 35-50 | - | IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | None | 0 | 0 | TEXT_WARLRD9 |
| 10 | The Butcher | MT_CLEAVER | genrl | 0 | 220 | Butcher | 3 | 6-12 | - | RESIST_FIRE,RESIST_LIGHTNING | None | 0 | 0 | - |
| 11 | Hork Demon | MT_HORKDMN | genrl | 19 | 300 | HorkDemon | 3 | 20-35 | - | RESIST_LIGHTNING | None | 0 | 0 | - |
| 12 | The Defiler | MT_DEFILER | genrl | 20 | 480 | SkeletonMelee | 3 | 30-40 | - | RESIST_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | None | 0 | 0 | - |
| 13 | Na-Krul | MT_NAKRUL | genrl | 0 | 1332 | SkeletonMelee | 3 | 40-50 | - | IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 14 | Bonehead Keenaxe | MT_TSKELAX | bhka | 2 | 91 | SkeletonMelee | 2 | 4-10 | - | IMMUNE_MAGIC | Leashed | 100 | 0 | - |
| 15 | Bladeskin the Slasher | MT_RFALLSD | bsts | 2 | 51 | Fallen | 0 | 6-18 | - | RESIST_FIRE | Leashed | 0 | 45 | - |
| 16 | Soulpus | MT_NZOMBIE | general | 2 | 133 | Zombie | 0 | 4-8 | - | RESIST_FIRE,RESIST_LIGHTNING | None | 0 | 0 | - |
| 17 | Pukerat the Unclean | MT_RFALLSP | ptu | 2 | 77 | Fallen | 3 | 1-5 | - | RESIST_FIRE | None | 0 | 0 | - |
| 18 | Boneripper | MT_WSKELAX | br | 2 | 54 | Bat | 0 | 6-15 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 19 | Rotfeast the Hungry | MT_NZOMBIE | eth | 2 | 85 | SkeletonMelee | 3 | 4-12 | - | IMMUNE_MAGIC | Leashed | 0 | 0 | - |
| 20 | Gutshank the Quick | MT_DFALLSD | gtq | 3 | 66 | Bat | 2 | 6-16 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 21 | Brokenhead Bangshield | MT_TSKELSD | bhbs | 3 | 108 | SkeletonMelee | 3 | 12-20 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 22 | Bongo | MT_YFALLSP | bng | 3 | 178 | Fallen | 3 | 9-21 | - | - | Leashed | 0 | 0 | - |
| 23 | Rotcarnage | MT_BZOMBIE | rcrn | 3 | 102 | Zombie | 3 | 9-24 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 0 | 45 | - |
| 24 | Shadowbite | MT_NSCAV | shbt | 2 | 60 | SkeletonMelee | 3 | 3-20 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 25 | Deadeye | MT_WSKELBW | de | 2 | 49 | GoatRanged | 0 | 6-9 | - | IMMUNE_MAGIC,RESIST_FIRE | None | 0 | 0 | - |
| 26 | Madeye the Dead | MT_RSKELAX | mtd | 4 | 75 | Bat | 0 | 9-21 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 30 | - |
| 27 | El Chupacabras | MT_BSCAV | general | 3 | 120 | GoatMelee | 0 | 10-18 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 28 | Skullfire | MT_TSKELBW | skfr | 3 | 125 | GoatRanged | 1 | 6-10 | - | IMMUNE_FIRE | None | 0 | 0 | - |
| 29 | Warpskull | MT_SNEAK | tspo | 3 | 117 | Sneak | 2 | 6-18 | - | RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 30 | Goretongue | MT_GZOMBIE | pmr | 3 | 156 | SkeletonMelee | 1 | 15-30 | - | IMMUNE_MAGIC | None | 0 | 0 | - |
| 31 | Pulsecrawler | MT_WSCAV | bhka | 4 | 150 | Scavenger | 0 | 16-20 | - | IMMUNE_FIRE,RESIST_LIGHTNING | Leashed | 0 | 45 | - |
| 32 | Moonbender | MT_BLINK | general | 4 | 135 | Bat | 0 | 9-27 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 33 | Wrathraven | MT_BLINK | general | 5 | 135 | Bat | 2 | 9-22 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 34 | Spineeater | MT_YSCAV | general | 4 | 180 | Scavenger | 1 | 18-25 | - | IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
#### unique monsters rows 35–68 of 100
| # | name | base | trn | lvl | maxHp | ai | int | dmg | red | res | pack | toHit | ac | talk |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 35 | Blackash the Burning | MT_RSKELBW | bashtb | 4 | 120 | GoatRanged | 0 | 6-16 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 36 | Shadowcrow | MT_BFALLSD | general | 5 | 270 | Sneak | 2 | 12-25 | - | - | Leashed | 0 | 0 | - |
| 37 | Blightstone the Weak | MT_LRDSAYTR | bhka | 4 | 360 | SkeletonMelee | 0 | 4-12 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 70 | 0 | - |
| 38 | Bilefroth the Pit Master | MT_FAT | bftp | 6 | 210 | Bat | 1 | 16-23 | - | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 39 | Bloodskin Darkbow | MT_NGOATBW | bsdb | 5 | 207 | GoatRanged | 0 | 3-16 | - | RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 55 | - |
| 40 | Foulwing | MT_GLOOM | db | 5 | 246 | Rhino | 3 | 12-28 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 41 | Shadowdrinker | MT_XSKELSD | shdr | 5 | 300 | Sneak | 1 | 18-26 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | None | 0 | 45 | - |
| 42 | Hazeshifter | MT_UNSEEN | bhka | 5 | 285 | Sneak | 3 | 18-30 | - | IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 43 | Deathspit | MT_NACID | bfds | 6 | 303 | AcidUnique | 0 | 12-32 | - | RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 44 | Bloodgutter | MT_RGOATMC | bgbl | 6 | 315 | Bat | 1 | 24-34 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 45 | Deathshade Fleshmaul | MT_BGOATMC | dsfm | 6 | 276 | Rhino | 0 | 12-24 | - | IMMUNE_MAGIC,RESIST_FIRE | None | 0 | 65 | - |
| 46 | Warmaggot the Mad | MT_WYRM | general | 6 | 246 | Bat | 3 | 15-30 | - | RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 47 | Glasskull the Jagged | MT_STORM | bhka | 7 | 354 | Storm | 0 | 18-30 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 48 | Blightfire | MT_RGOATBW | blf | 7 | 321 | Succubus | 2 | 13-21 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 49 | Nightwing the Cold | MT_GARGOYLE | general | 7 | 342 | Bat | 1 | 18-26 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 50 | Gorestone | MT_GGOATBW | general | 7 | 303 | GoatRanged | 1 | 15-28 | - | RESIST_LIGHTNING | Leashed | 70 | 0 | - |
| 51 | Bronzefist Firestone | MT_BMAGMA | general | 8 | 360 | Magma | 0 | 30-36 | - | IMMUNE_MAGIC,RESIST_FIRE | Leashed | 0 | 0 | - |
| 52 | Wrathfire the Doomed | MT_INCIN | wftd | 8 | 270 | SkeletonMelee | 2 | 20-30 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 53 | Firewound the Grim | MT_NMAGMA | bhka | 8 | 303 | Magma | 0 | 18-22 | - | IMMUNE_MAGIC,RESIST_FIRE | Leashed | 0 | 0 | - |
| 54 | Baron Sludge | MT_MUDMAN | bsm | 8 | 315 | Sneak | 3 | 25-34 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 75 | - |
| 55 | Blighthorn Steelmace | MT_GGOATMC | bhsm | 7 | 250 | Rhino | 0 | 20-28 | - | RESIST_LIGHTNING | Leashed | 0 | 45 | - |
| 56 | Chaoshowler | MT_RACID | general | 8 | 240 | AcidUnique | 0 | 12-20 | - | - | Leashed | 0 | 0 | - |
| 57 | Doomgrin the Rotting | MT_REDDTH | general | 8 | 405 | Storm | 3 | 25-50 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 58 | Madburner | MT_FLAMLRD | general | 9 | 270 | Storm | 0 | 20-40 | - | IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 59 | Bonesaw the Litch | MT_LTCHDMN | general | 9 | 495 | Storm | 2 | 30-55 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 60 | Breakspine | MT_MUDRUN | general | 9 | 351 | Rhino | 0 | 25-34 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 61 | Devilskull Sharpbone | MT_REDDTH | general | 9 | 444 | Storm | 1 | 25-40 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 62 | Brokenstorm | MT_STORM | general | 9 | 411 | Storm | 2 | 25-36 | - | IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 63 | Stormbane | MT_RSTORM | general | 9 | 555 | Storm | 3 | 30-30 | - | IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 64 | Oozedrool | MT_TOAD | general | 9 | 483 | Fat | 3 | 25-30 | - | RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 65 | Goldblight of the Flame | MT_BLOODCLW | general | 10 | 405 | Gargoyle | 0 | 15-35 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 80 | - |
| 66 | Blackstorm | MT_OBLORD | general | 10 | 525 | Rhino | 3 | 20-40 | - | IMMUNE_MAGIC,IMMUNE_LIGHTNING | Leashed | 0 | 90 | - |
| 67 | Plaguewrath | MT_RACID | general | 10 | 450 | AcidUnique | 2 | 20-30 | - | IMMUNE_MAGIC,RESIST_FIRE | Leashed | 0 | 0 | - |
| 68 | The Flayer | MT_RSTORM | general | 10 | 501 | Storm | 1 | 20-35 | - | RESIST_MAGIC,RESIST_FIRE,IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
#### unique monsters rows 69–100 of 100
| # | name | base | trn | lvl | maxHp | ai | int | dmg | red | res | pack | toHit | ac | talk |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 69 | Bluehorn | MT_FROSTC | general | 11 | 477 | Rhino | 1 | 25-30 | - | IMMUNE_MAGIC,RESIST_FIRE | Leashed | 0 | 90 | - |
| 70 | Warpfire Hellspawn | MT_HELLBURN | general | 11 | 525 | FireMan | 3 | 10-40 | - | RESIST_MAGIC,IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 71 | Fangspeir | MT_NSNAKE | general | 11 | 444 | SkeletonMelee | 1 | 15-32 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 72 | Festerskull | MT_UDEDBLRG | general | 11 | 600 | Storm | 2 | 15-30 | - | IMMUNE_MAGIC | Leashed | 0 | 0 | - |
| 73 | Lionskull the Bent | MT_NBLACK | general | 12 | 525 | SkeletonMelee | 2 | 25-25 | - | IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 74 | Blacktongue | MT_COUNSLR | general | 12 | 360 | Counselor | 3 | 15-30 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 75 | Viletouch | MT_DEATHW | general | 12 | 525 | Gargoyle | 3 | 20-40 | - | IMMUNE_LIGHTNING | Leashed | 0 | 0 | - |
| 76 | Viperflame | MT_RSNAKE | general | 12 | 570 | SkeletonMelee | 1 | 25-35 | - | IMMUNE_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 77 | Fangskin | MT_BSNAKE | bhka | 14 | 681 | SkeletonMelee | 2 | 15-50 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 78 | Witchfire the Unholy | MT_SUCCUBUS | general | 12 | 444 | Succubus | 3 | 10-20 | - | IMMUNE_MAGIC,IMMUNE_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 79 | Blackskull | MT_BALROG | bhka | 13 | 750 | SkeletonMelee | 3 | 25-40 | - | IMMUNE_MAGIC,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 80 | Soulslash | MT_UNRAV | general | 12 | 450 | SkeletonMelee | 0 | 25-25 | - | IMMUNE_MAGIC | Leashed | 0 | 0 | - |
| 81 | Windspawn | MT_VTEXLRD | general | 12 | 711 | SkeletonMelee | 1 | 35-40 | - | IMMUNE_MAGIC,IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 82 | Lord of the Pit | MT_GSNAKE | general | 13 | 762 | SkeletonMelee | 2 | 25-42 | - | RESIST_FIRE | Leashed | 0 | 0 | - |
| 83 | Rustweaver | MT_RTBLACK | general | 13 | 400 | SkeletonMelee | 3 | 1-60 | - | IMMUNE_MAGIC,IMMUNE_FIRE,IMMUNE_LIGHTNING | None | 0 | 0 | - |
| 84 | Howlingire the Shade | MT_HOLOWONE | general | 13 | 450 | SkeletonMelee | 2 | 40-75 | - | RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 85 | Doomcloud | MT_MAEL | general | 13 | 612 | Storm | 1 | 1-60 | - | RESIST_FIRE,IMMUNE_LIGHTNING | None | 0 | 0 | - |
| 86 | Bloodmoon Soulfire | MT_PAINMSTR | general | 13 | 684 | SkeletonMelee | 1 | 15-40 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 87 | Witchmoon | MT_SNOWWICH | general | 13 | 310 | Succubus | 3 | 30-40 | - | RESIST_LIGHTNING | None | 0 | 0 | - |
| 88 | Gorefeast | MT_VTEXLRD | general | 13 | 771 | SkeletonMelee | 3 | 20-55 | - | RESIST_FIRE | None | 0 | 0 | - |
| 89 | Graywar the Slayer | MT_RTBLACK | general | 14 | 672 | SkeletonMelee | 1 | 30-50 | - | RESIST_LIGHTNING | None | 0 | 0 | - |
| 90 | Dreadjudge | MT_MAGISTR | general | 14 | 540 | Counselor | 1 | 30-40 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 91 | Stareye the Witch | MT_HLSPWN | general | 14 | 726 | Succubus | 2 | 30-50 | - | IMMUNE_FIRE | None | 0 | 0 | - |
| 92 | Steelskull the Hunter | MT_BTBLACK | general | 14 | 831 | SkeletonMelee | 3 | 40-50 | - | RESIST_LIGHTNING | None | 0 | 0 | - |
| 93 | Sir Gorash | MT_RBLACK | general | 16 | 1050 | SkeletonMelee | 1 | 20-60 | - | - | None | 0 | 0 | - |
| 94 | The Vizier | MT_CABALIST | general | 15 | 850 | Counselor | 2 | 25-40 | - | IMMUNE_FIRE | Leashed | 0 | 0 | - |
| 95 | Zamphir | MT_REALWEAV | general | 15 | 891 | SkeletonMelee | 2 | 30-50 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
| 96 | Bloodlust | MT_HLSPWN | general | 15 | 825 | Succubus | 1 | 20-55 | - | IMMUNE_MAGIC,IMMUNE_LIGHTNING | None | 0 | 0 | - |
| 97 | Webwidow | MT_HLSPWN | general | 16 | 774 | Succubus | 1 | 20-50 | - | IMMUNE_MAGIC,IMMUNE_FIRE | None | 0 | 0 | - |
| 98 | Fleshdancer | MT_SOLBRNR | general | 16 | 999 | Succubus | 3 | 30-50 | - | IMMUNE_MAGIC,RESIST_FIRE | None | 0 | 0 | - |
| 99 | Grimspike | MT_OBLORD | general | 19 | 534 | Sneak | 1 | 25-40 | - | IMMUNE_MAGIC,RESIST_FIRE | Leashed | 0 | 0 | - |
| 100 | Doomlock | MT_STORML | general | 28 | 534 | Sneak | 1 | 35-55 | - | IMMUNE_MAGIC,RESIST_FIRE,RESIST_LIGHTNING | Leashed | 0 | 0 | - |
Row 1 = `Gharbad the Weak` (base MT_NGOATMC, trn `bsdb`, ai `Gharbad`, IMMUNE_LIGHTNING, pack
`none`, talk TEXT_GARBUD1); row 100 = `Doomlock` (base `MT_STORML`).
`monsterPack` histogram: Leashed=70, None=29, Independent=1. `talkMessage` non-empty on 8 rows (drives `InitQTextMsg` dialogs).
### V2.3 Monster AI — `AiProc[]` function table (`Source/monster.cpp:3090-3130`), dispatched at `:4323`
| # | MonsterAIID (monstdat.h:22-63) | AiProc entry (monster.cpp:3090+) | implementation | behavior |
|---|---|---|---|---|
| 1 | Zombie | AiProc[0] | ZombieAi | shambles to the target, melee only, door checks; low intelligence random-walks. |
| 2 | Fat | AiProc[1] | OverlordAi | Hellfire Omega/overlord: ranged spew, avoids melee. |
| 3 | SkeletonMelee | AiProc[2] | SkeletonAi | melee rusher; skeleton resurrection handling. |
| 4 | SkeletonRanged | AiProc[3] | SkeletonBowAi | bow AI: Arrow projectiles, keeps range. |
| 5 | Scavenger | AiProc[4] | ScavengerAi | feeds on corpses, flees when disturbed (L3). |
| 6 | Rhino | AiProc[5] | RhinoAi | telegraphed charge (`Viper`/rhino dash) across the room. |
| 7 | GoatMelee | AiProc[6] | AiAvoidance (`:1893`) | melee + avoidance; special attack when hp < maxHp/2 with `!FlipCoin()` (`:1930-1933`), else `StartAttack` (`:1935`); `GetMissileType` not used. |
| 8 | GoatRanged | AiProc[7] | AiRanged (`:1976`) | ranged Arrow; `AiDelay(GenerateRnd(20))` cadence (`:1990-1996`), closes when target within 4 tiles. |
| 9 | Fallen | AiProc[8] | FallenAi | pack behaviour: scatter/flee when the leader dies; knife throw. |
| 10 | Magma | AiProc[9] | AiRangedAvoidance (`:2013`) | MagmaBall ranged + avoidance. |
| 11 | SkeletonKing | AiProc[10] | LeoricAi | Leoric: summons minions/soul-fire missiles. |
| 12 | Bat | AiProc[11] | BatAi | fast erratic melee (L3/L4). |
| 13 | Gargoyle | AiProc[12] | GargoyleAi | stone until player near; `PlaceGroup` refuses to overwrite it (`monster.cpp:357`). |
| 14 | Butcher | AiProc[13] | ButcherAi | level-2 sprinter, no pathing; `MT_CLEAVER` greeting SFX on activation (`monster.cpp:4278-4281`). |
| 15 | Succubus | AiProc[14] | AiRanged | BloodStar ranged, flying. |
| 16 | Sneak | AiProc[15] | SneakAi | Hellfire sneak: invisible until it strikes. |
| 17 | Storm | AiProc[16] | AiRangedAvoidance | ThinLightningControl chain lightning. |
| 18 | FireMan | AiProc[17] | nullptr (`:3107`) | no AI — elemental behaviour lives in the missile code. |
| 19 | Gharbad | AiProc[18] | GharbadAi | quest NPC: dialog-driven (`TEXT_GARBUD*`), attacks when double-crossed. |
| 20 | Acid | AiProc[19] | AiRangedAvoidance | Acid spitter + puddle missiles (Hive). |
| 21 | AcidUnique | AiProc[20] | AiRanged | unique variant of the acid spitter. |
| 22 | Golem | AiProc[21] | GolumAi | player-summoned golem: follow + melee. |
| 23 | Zhar | AiProc[22] | ZharAi | quest caster, dialog gated (`TEXT_ZHAR*`), Fireball ranged. |
| 24 | Snotspill | AiProc[23] | SnotSpilAi | Zhar's pet; state driven by the quest. |
| 25 | Snake | AiProc[24] | SnakeAi | Lazarus-quest snakes, idle until triggered. |
| 26 | Counselor | AiProc[25] | CounselorAi | Hellfire counselor: flare attack. |
| 27 | Mega | AiProc[26] | MegaAi | Na-Krul mega demon, heavy ranged. |
| 28 | Diablo | AiProc[27] | AiRangedAvoidance | boss: DiabloApocalypse + phasing fireballs, HP-phase gates. |
| 29 | Lazarus | AiProc[28] | LazarusAi | boss: teleports, ranged, `TEXT_VILE*` dialogs. |
| 30 | LazarusSuccubus | AiProc[29] | LazarusMinionAi | Lazarus's succubus minions (BloodStar). |
| 31 | Lachdanan | AiProc[30] | LachdananAi | quest warrior: dialog → attack (`TEXT_VEIL*`). |
| 32 | Warlord | AiProc[31] | WarlordAi | Bloodwarlord: summons his pack (`RedVex`/`BlackJade`/`Vile`, `monster.cpp:566-607`). |
| 33 | FireBat | AiProc[32] | AiRanged | Firebolt ranged. |
| 34 | Torchant | AiProc[33] | AiRanged | Fireball ranged. |
| 35 | HorkDemon | AiProc[34] | HorkDemonAi | Hellfire hork: teleport + `HorkSpawn`; on death spawns Theodore when `bTheoQuest` else `CreateAmulet` (`monster.cpp:925-929`). |
| 36 | Lich | AiProc[35] | AiRanged | OrangeFlare. |
| 37 | ArchLich | AiProc[36] | AiRanged | YellowFlare. |
| 38 | Psychorb | AiProc[37] | AiRanged | BlueFlare. |
| 39 | Necromorb | AiProc[38] | AiRanged | RedFlare. |
| 40 | BoneDemon | AiProc[39] | AiRangedAvoidance | BlueFlare2 (Nest boss). |
| 41 | Invalid = -1 | AiProc[40] | — | ERRATUM (devilGo-hy8e): the table lists 41 rows but AiProc has exactly **40** entries (monster.cpp:3089-3130, indices 0-39); `Invalid = -1` is NOT an array slot. Exactly one entry is nullptr (FireMan). |
Ranged missile selection — `GetMissileType(MonsterAIID)` (`Source/monster.cpp:1941-1972`):
GoatRanged→Arrow, Succubus/LazarusSuccubus→BloodStar, Acid/AcidUnique→Acid, FireBat→Firebolt,
Torchant→Fireball, Lich→OrangeFlare, ArchLich→YellowFlare, Psychorb→BlueFlare,
Necromorb→RedFlare, Magma→MagmaBall, Storm→ThinLightningControl, Diablo→DiabloApocalypse,
BoneDemon→BlueFlare2, default→Arrow. Special-attack damage uses `toHitSpecial`/`min/maxDamageSpecial`.
Shared primitives: `AiDelay` (`:755`), `StartAttack` (`:823`), `StartRangedAttack` (`:832`),
`RandomWalk` with intelligence roll `v < 2*intelligence+33` (or `+83` while mid-move, `:3050-3058`),
`MonstCheckDoors`, `IsRelativeMoveOK` (`:3132+`, per-direction solid checks), `MFLAG_*`
(`MFLAG_TARGETS_MONSTER` used by AiRanged `:1984`), `checkStandAnimationIsLoaded` (`:1938`, `:3063`).
### V2.4 Pack/spawn placement rules (`Source/monster.cpp`)
- `PlaceGroup(size_t typeIndex, size_t num, Monster *leader, bool leashed)` (`:308-380`):
  - outer `try1 < 10` loop rewinds the whole group on failure (`:312-319`).
  - anchor: leader ⇒ `leader.tile + Direction(GenerateRnd(8))` (`:321-325`); else `do { x = 16+GenerateRnd(80); y = 16+GenerateRnd(80) } while (!CanPlaceMonster)` (`:326-331`) — the 40×40 playable band.
  - budget: `if (num + ActiveMonsterCount > totalmonsters) num = totalmonsters - ActiveMonsterCount` (`:333-335`).
  - scatter: up to 100 tries, step `Displacement(Direction(GenerateRnd(8)))` in x **and** the
    documented vanilla bug `yp += Displacement(...).deltaX` (`:338` — keep verbatim for parity).
  - placement gates: `CanPlaceMonster`, **same `dTransVal` as the anchor** (`:341`), leashed ⇒
    `|dx| < 4 && |dy| < 4` (`:342`).
  - leader minions: `maxHitPoints *= 2; hitPoints = maxHitPoints; intelligence = leader->intelligence`
    (`:352-355`), `setLeader` when leashed (`:357-359`), and `ai != Gargoyle` extra handling (`:360`).
- `AddMonsterType(type, placeType)` / `AddMonster` (`:566-603`, `:1592`) — per-level type lists with
  `PLACE_*` flags; `IsMonsterAvailable` (`:3159`) filters spawn/level window.
- `PlaceUniqueMonst(uniqindex, minionType, bosspacksize)` (`:463`), `PlaceUniqueMonsters()` (`:504-610`):
  e.g. `PlaceUniqueMonst(SkeletonKing, i, 30)` (`:541`), Warlord/Lachdanan scatter (`:566-569`),
  Lazarus+RedVex (`:576-579`), BlackJade/Na-Krul Hellfire branches (`:580`, `:598`, `:601-607`).
- Diablo and quest monsters are level-specific: `!gbIsSpawn && !setlevel && currlevel == 16` (`:3688`),
  `if (!gbIsSpawn)` gates (`:3701-3704`), `gbIsSpawn` early-out (`:4087`).
- Containers: `MaxMonsters`/`ActiveMonsters` (`Source/monsters.h`), `dMonster` occupancy 4-tile trick.
---
## V3. MISSILES — 68 misdat rows + 46 missile_sprites rows + behavior registry
### V3.1 misdat.tsv rows 2-69 (68 rows, complete; loader `tables/misdat.cpp:373-396`)
Columns: `id | addFn | processFn | castSound | hitSound | graphic | flags | movementDistribution`.
Parsing (`misdat.cpp:377-390`): `addFn`/`processFn` via `ParseMissileAddFn` (`:181`) /
`ParseMissileProcessFn` (`:261`), sounds via `ParseCastSound`/`ParseHitSound`, `graphic` via
`ParseMissileGraphicID` (`:67`), `flags` as enum list, `movementDistribution` via `:344-349`
(``⇒Disabled, `Blockable`, `Unblockable`). Damage type = low 3 bits of `flags`
(`MissileData::damageType()`, `misdat.h:190-193`); `isDrawn()` = not `Invisible` (`:182`).
Row order **is** the `MissileID` index: `GetMissileData(id)` = `MissilesData[id]` (`:448-451`),
and `misdat.cpp:393` asserts `MissileID::LastDiablo + 1 == size` (68 = `DiabloApocalypse` index 67).
#### misdat rows 1–34 of 68
| # | id | addFn | processFn | castSound | hitSound | graphic | flags | movement |
|---|---|---|---|---|---|---|---|---|
| 1 | Arrow | AddArrow | ProcessArrow | - | - | Arrow | Physical,Arrow | Blockable |
| 2 | Firebolt | AddFirebolt | ProcessGenericProjectile | SpellFirebolt | SpellFireHit | Fireball | Fire | Blockable |
| 3 | Guardian | AddGuardian | ProcessGuardian | SpellGuardian | - | Guardian | Physical | Disabled |
| 4 | Phasing | AddPhasing | ProcessTeleport | SpellTeleport | - | - | Physical,Invisible | Disabled |
| 5 | NovaBall | AddNovaBall | ProcessNovaBall | - | - | Lightning | Lightning | Unblockable |
| 6 | FireWall | AddFireWall | ProcessFireWall | SpellFireWall | SpellFireHit | FireWall | Fire | Disabled |
| 7 | Fireball | AddFireball | ProcessFireball | SpellFirebolt | SpellFireHit | Fireball | Fire | Blockable |
| 8 | LightningControl | AddLightningControl | ProcessLightningControl | - | - | Lightning | Lightning,Invisible | Disabled |
| 9 | Lightning | AddLightning | ProcessLightning | SpellLightning | SpellLightningHit | Lightning | Lightning | Disabled |
| 10 | MagmaBallExplosion | AddMissileExplosion | ProcessMissileExplosion | - | - | MagmaBallExplosion | Physical | Disabled |
| 11 | TownPortal | AddTownPortal | ProcessTownPortal | SpellPortal | - | TownPortal | Magic | Disabled |
| 12 | FlashBottom | AddFlashBottom | ProcessFlashBottom | SpellNova | SpellLightningHit | FlashBottom | Magic | Disabled |
| 13 | FlashTop | AddFlashTop | ProcessFlashTop | - | - | FlashTop | Magic | Disabled |
| 14 | ManaShield | AddManaShield | - | SpellManaShield | - | ManaShield | Magic,Invisible | Disabled |
| 15 | FlameWave | AddFlameWave | ProcessFlameWave | - | - | FireWall | Fire | Unblockable |
| 16 | ChainLightning | AddChainLightning | ProcessChainLightning | SpellLightning | SpellLightningHit | Lightning | Lightning | Disabled |
| 17 | ChainBall | - | - | - | - | Lightning | Lightning | Disabled |
| 18 | BloodHit | - | - | SpellBloodStar | SpellBloodStarHit | BloodHit | Physical | Disabled |
| 19 | BoneHit | - | - | - | - | BoneHit | Physical | Disabled |
| 20 | MetalHit | - | - | - | - | MetalHit | Physical | Disabled |
| 21 | Rhino | AddRhino | ProcessRhino | - | - | - | Physical | Blockable |
| 22 | MagmaBall | AddMagmaBall | ProcessGenericProjectile | - | - | MagmaBall | Fire | Blockable |
| 23 | ThinLightningControl | AddLightningControl | ProcessLightningControl | - | - | ThinLightning | Lightning,Invisible | Disabled |
| 24 | ThinLightning | AddLightning | ProcessLightning | - | - | ThinLightning | Lightning | Disabled |
| 25 | BloodStar | AddGenericMagicMissile | ProcessGenericProjectile | - | - | BloodStar | Magic | Blockable |
| 26 | BloodStarExplosion | AddMissileExplosion | ProcessMissileExplosion | - | - | BloodStarExplosion | Magic | Disabled |
| 27 | Teleport | AddTeleport | ProcessTeleport | SpellElemental | - | - | Physical,Invisible | Disabled |
| 28 | FireArrow | AddElementalArrow | ProcessElementalArrow | - | - | FireArrow | Fire,Arrow | Blockable |
| 29 | DoomSerpents | - | - | SpellDoomSerpents | - | DoomSerpents | Magic,Invisible | Disabled |
| 30 | FireOnly | - | - | - | - | FireWall | Fire | Disabled |
| 31 | StoneCurse | AddStoneCurse | ProcessStoneCurse | SpellStoneCurse | - | - | Magic,Invisible | Disabled |
| 32 | BloodRitual | - | - | - | - | - | Physical | Disabled |
| 33 | Invisibility | - | - | SpellInvisibility | - | - | Physical,Invisible | Disabled |
| 34 | Golem | AddGolem | - | SpellGolem | - | - | Physical,Invisible | Disabled |
#### misdat rows 35–68 of 68
| # | id | addFn | processFn | castSound | hitSound | graphic | flags | movement |
|---|---|---|---|---|---|---|---|---|
| 35 | Etherealize | - | - | SpellEtherealize | - | Etherealize | Physical | Disabled |
| 36 | Spurt | - | - | - | - | Spurt | Physical | Disabled |
| 37 | ApocalypseBoom | AddApocalypseBoom | ProcessApocalypseBoom | - | - | ApocalypseBoom | Physical | Disabled |
| 38 | Healing | AddHealing | - | - | - | - | Physical,Invisible | Disabled |
| 39 | FireWallControl | AddWallControl | ProcessWallControl | - | - | FireWall | Fire,Invisible | Disabled |
| 40 | Infravision | AddInfravision | ProcessInfravision | SpellInfravision | - | - | Physical,Invisible | Disabled |
| 41 | Identify | AddIdentify | - | - | - | - | Physical,Invisible | Disabled |
| 42 | FlameWaveControl | AddFlameWaveControl | ProcessFlameWaveControl | SpellFlameWave | - | FireWall | Fire | Disabled |
| 43 | Nova | AddNova | ProcessNova | SpellNova | - | Lightning | Lightning | Disabled |
| 44 | Rage | AddRage | ProcessRage | - | - | - | Physical,Invisible | Disabled |
| 45 | Apocalypse | AddApocalypse | ProcessApocalypse | SpellApocalypse | - | ApocalypseBoom | Magic | Disabled |
| 46 | ItemRepair | AddItemRepair | - | - | - | - | Physical,Invisible | Disabled |
| 47 | StaffRecharge | AddStaffRecharge | - | - | - | - | Physical,Invisible | Disabled |
| 48 | TrapDisarm | AddTrapDisarm | - | SpellTrapDisarm | - | - | Physical,Invisible | Disabled |
| 49 | Inferno | AddInferno | ProcessInferno | SpellInferno | - | Inferno | Fire | Disabled |
| 50 | InfernoControl | AddInfernoControl | ProcessInfernoControl | - | - | - | Fire,Invisible | Disabled |
| 51 | FireMan | - | - | - | - | - | Physical | Blockable |
| 52 | Krull | - | - | - | - | Krull | Fire,Arrow | Blockable |
| 53 | ChargedBolt | AddChargedBolt | ProcessChargedBolt | SpellChargedBolt | - | ChargedBolt | Lightning | Blockable |
| 54 | HolyBolt | AddHolyBolt | ProcessHolyBolt | SpellHolyBolt | SpellLightningHit | HolyBolt | Physical | Blockable |
| 55 | Resurrect | AddResurrect | - | - | SpellResurrect | - | Magic,Invisible | Disabled |
| 56 | Telekinesis | AddTelekinesis | - | SpellEtherealize | - | - | Physical,Invisible | Disabled |
| 57 | LightningArrow | AddElementalArrow | ProcessElementalArrow | - | - | LightningArrow | Lightning,Arrow | Blockable |
| 58 | Acid | AddAcid | ProcessGenericProjectile | SpellAcid | - | Acid | Acid | Blockable |
| 59 | AcidSplat | AddMissileExplosion | ProcessAcidSplate | - | - | AcidSplat | Acid | Disabled |
| 60 | AcidPuddle | AddAcidPuddle | ProcessAcidPuddle | SpellPuddle | - | AcidPuddle | Acid | Disabled |
| 61 | HealOther | AddHealOther | - | - | - | - | Physical,Invisible | Disabled |
| 62 | Elemental | AddElemental | ProcessElemental | SpellElemental | - | Elemental | Fire | Unblockable |
| 63 | ResurrectBeam | AddResurrectBeam | ProcessResurrectBeam | - | - | Resurrect | Physical | Disabled |
| 64 | BoneSpirit | AddBoneSpirit | ProcessBoneSpirit | SpellBoneSpirit | SpellBoneSpiritHit | BoneSpirit | Magic | Blockable |
| 65 | WeaponExplosion | AddWeaponExplosion | ProcessWeaponExplosion | - | - | - | Physical | Disabled |
| 66 | RedPortal | AddRedPortal | ProcessRedPortal | SpellPortal | - | RedPortal | Physical | Disabled |
| 67 | DiabloApocalypseBoom | AddApocalypseBoom | ProcessApocalypseBoom | - | - | DiabloApocalypseBoom | Physical | Disabled |
| 68 | DiabloApocalypse | AddDiabloApocalypse | - | - | - | - | Physical,Invisible | Disabled |
Flag vocabulary: Acid, Arrow, Fire, Invisible, Lightning, Magic, Physical.
addFn vocabulary (50 distinct): AddAcid, AddAcidPuddle, AddApocalypse, AddApocalypseBoom, AddArrow, AddBoneSpirit, AddChainLightning, AddChargedBolt, AddDiabloApocalypse, AddElemental, AddElementalArrow, AddFireWall, AddFireball, AddFirebolt, AddFlameWave, AddFlameWaveControl, AddFlashBottom, AddFlashTop, AddGenericMagicMissile, AddGolem, AddGuardian, AddHealOther, AddHealing, AddHolyBolt, AddIdentify, AddInferno, AddInfernoControl, AddInfravision, AddItemRepair, AddLightning, AddLightningControl, AddMagmaBall, AddManaShield, AddMissileExplosion, AddNova, AddNovaBall, AddPhasing, AddRage, AddRedPortal, AddResurrect, AddResurrectBeam, AddRhino, AddStaffRecharge, AddStoneCurse, AddTelekinesis, AddTeleport, AddTownPortal, AddTrapDisarm, AddWallControl, AddWeaponExplosion.
processFn vocabulary (36 distinct): ProcessAcidPuddle, ProcessAcidSplate, ProcessApocalypse, ProcessApocalypseBoom, ProcessArrow, ProcessBoneSpirit, ProcessChainLightning, ProcessChargedBolt, ProcessElemental, ProcessElementalArrow, ProcessFireWall, ProcessFireball, ProcessFlameWave, ProcessFlameWaveControl, ProcessFlashBottom, ProcessFlashTop, ProcessGenericProjectile, ProcessGuardian, ProcessHolyBolt, ProcessInferno, ProcessInfernoControl, ProcessInfravision, ProcessLightning, ProcessLightningControl, ProcessMissileExplosion, ProcessNova, ProcessNovaBall, ProcessRage, ProcessRedPortal, ProcessResurrectBeam, ProcessRhino, ProcessStoneCurse, ProcessTeleport, ProcessTownPortal, ProcessWallControl, ProcessWeaponExplosion.
### V3.2 Behavior classes — Go registry (replaces the C++ function-pointer table)
- **Add\* implementations**: `Source/missiles.cpp` — 74 `void Add...` functions, e.g. `AddArrow`,
  `AddFirebolt`, `AddOpenNest` (`:1238`, spawns 4 `BigExplosion` missiles at (80/81, 62/63) →
  (80,62) then self-deletes), `AddRune` (`:667`), `AddRuneOfFire/Light/Nova/Immolation/Stone`
  (`:1248/:1253/:1261/:1266/:1271`), `AddReflect` (`:1276`), `AddBerserk` (`:1293`),
  `AddHorkSpawn` (`:1344`), `AddJester` (`:1352`), `AddStealPotions` (`:1390`), `AddStealMana`
  (`:1456`), `AddSpectralArrow` (`:1477`, fixes the vanilla damage bug behind `gbIsHellfire`),
  `AddWarp` (`:1505`), `AddLightningWall` (`:1580`), `AddBigExplosion` (`:1601`),
  `AddImmolation` (`:1618`), `AddLightningBow` (`:1634`), `AddMana` (`:1653`), `AddRedPortal` (`:2773`).
- **Common Add path** — `AddMissile(src, dst, midir, mitype, micaster, id, midam, spllvl, parent,`
  `lSFX)` (`missiles.cpp:278+`): pool append with `Missiles.max_size()` guard, `GetMissileData`)
  drives `_miAnimType`, `isDrawn()`, default/8-direction frame group (`:300-305`), unique monster
  tinting (`:295-298`), then `addFn(missile, parameter)` (`:2840` region).
- **Process\***: `ProcessGenericProjectile`, `ProcessArrow`, `ProcessFireWall`/`ProcessBurn`,
  `ProcessLightning`, `ProcessLightningArrow`, `ProcessChargedBolt`, `ProcessElemental`,
  `ProcessNova`, `ProcessApocalypse`, `ProcessManaShield`, `ProcessEtherealize`, `ProcessGolem`,
  `ProcessAcid`/`ProcessAcidSplat`/`ProcessPuddle`, `ProcessRuneOf*`, `ProcessReflect`,
  `ProcessRedPortal` (`:4162-4180`), `ProcessMissiles` list driver (`:4198`).
- **Motion/collision core**: `CheckMissileCol(missile, damageType, min, max, isDamageShifted,`
  `position, dontDeleteOnCollision, onlyHitWalking)` (`missiles.cpp:479`);
  `MoveMissileAndCheckMissileCol` (`:636`) with fixed-point 16.16 `position.traveled`;
  `Blockable` deflection at `:648`; light radius at draw (`:1338` `DTYPE_NEST ? 9 : 3`).
- **Go decision**: `map[MissileID]struct{Add AddFn; Process ProcessFn}` built from the TSV; a
  missing entry must fail a count test (68 rows == registry size; dg-task-022 AC).
### V3.3 missile_sprites.tsv rows 2-47 (46 rows, complete; loader `tables/misdat.cpp:134`)
Columns `id | width | width2 | name | numFrames | flags | frameDelay[16] | frameLength[16]`;
delay/length arrays are de-duplicated into shared tables (`misdat.cpp:47-55` `ToIndex`).
Sprite file resolution: `MissileFileData::LoadGFX()` (`misdat.cpp:400-430`) — `animFAmt==1` ⇒
single `missiles\<name>` CL2 (`LoadCl2WithStatus`), else 8/16-direction sheet
(`LoadMultipleCl2Sheet<16>` with `.cl2`), or `missiles\<name>.clx` under `UNPACKED_MPQS`.
Graphics flags via `ParseMissileGraphicsFlag` (`misdat.cpp:57+`); `MonsterOwned` sprites are
skipped at init (`misdat.cpp:444-450`).
(`delays`/`lengths` shown as value×run-length over the 16 slots.)
#### missile_sprites rows 1–46 of 46
| # | id | width | width2 | name | numFrames | flags | delays | lengths |
|---|---|---|---|---|---|---|---|---|
| 1 | Arrow | 96 | 16 | arrows | 1 | NotAnimated | 0×16 | 16×16 |
| 2 | Fireball | 96 | 16 | fireba | 16 | - | 0×16 | 14×16 |
| 3 | Guardian | 96 | 16 | guard | 3 | - | 1×16 | 15×1 14×1 3×1 0×13 |
| 4 | Lightning | 96 | 16 | lghning | 1 | - | 0×16 | 8×16 |
| 5 | FireWall | 128 | 32 | firewal | 2 | - | 0×16 | 13×1 11×1 0×14 |
| 6 | MagmaBallExplosion | 128 | 32 | magblos | 1 | - | 1×16 | 10×16 |
| 7 | TownPortal | 96 | 16 | portal | 2 | - | 0×1 1×1 0×14 | 16×16 |
| 8 | FlashBottom | 160 | 48 | bluexfr | 1 | - | 0×16 | 19×16 |
| 9 | FlashTop | 160 | 48 | bluexbk | 1 | - | 0×16 | 19×16 |
| 10 | ManaShield | 96 | 16 | manashld | 1 | NotAnimated | 0×16 | 1×16 |
| 11 | BloodHit | 96 | 16 |  | 4 | - | 0×16 | 15×16 |
| 12 | BoneHit | 128 | 32 |  | 3 | - | 2×16 | 8×16 |
| 13 | MetalHit | 96 | 16 |  | 3 | - | 2×16 | 10×16 |
| 14 | FireArrow | 96 | 16 | farrow | 16 | - | 0×16 | 4×16 |
| 15 | DoomSerpents | 96 | 16 | doom | 9 | MonsterOwned | 1×16 | 15×16 |
| 16 | Golem | 0 | 0 |  | 1 | MonsterOwned | 0×16 | 0×16 |
| 17 | Spurt | 128 | 32 |  | 2 | - | 2×16 | 8×16 |
| 18 | ApocalypseBoom | 96 | 16 | newexp | 1 | - | 1×16 | 15×16 |
| 19 | StoneCurseShatter | 128 | 32 | shatter1 | 1 | - | 1×16 | 12×16 |
| 20 | BigExplosion | 160 | 48 | bigexp | 1 | - | 0×16 | 15×16 |
| 21 | Inferno | 96 | 16 | inferno | 1 | - | 0×16 | 20×16 |
| 22 | ThinLightning | 96 | 16 | thinlght | 1 | MonsterOwned | 0×16 | 8×16 |
| 23 | BloodStar | 128 | 32 | flare | 1 | - | 0×16 | 16×16 |
| 24 | BloodStarExplosion | 128 | 32 | flareexp | 1 | - | 0×16 | 7×16 |
| 25 | MagmaBall | 128 | 32 | magball | 8 | MonsterOwned | 1×16 | 16×16 |
| 26 | Krull | 96 | 16 | krull | 1 | MonsterOwned | 0×16 | 14×16 |
| 27 | ChargedBolt | 64 | 0 | miniltng | 1 | - | 1×16 | 8×16 |
| 28 | HolyBolt | 96 | 16 | holy | 16 | - | 1×1 0×15 | 14×16 |
| 29 | HolyBoltExplosion | 160 | 48 | holyexpl | 1 | - | 0×16 | 8×16 |
| 30 | LightningArrow | 96 | 16 | larrow | 16 | - | 0×16 | 4×16 |
| 31 | FireArrowExplosion | 64 | 0 |  | 1 | - | 0×16 | 6×16 |
| 32 | Acid | 96 | 16 | acidbf | 16 | MonsterOwned | 0×16 | 8×16 |
| 33 | AcidSplat | 96 | 16 | acidspla | 1 | MonsterOwned | 0×16 | 8×16 |
| 34 | AcidPuddle | 96 | 16 | acidpud | 2 | MonsterOwned | 0×16 | 9×1 4×1 0×14 |
| 35 | Etherealize | 96 | 16 |  | 1 | - | 0×16 | 1×16 |
| 36 | Elemental | 96 | 16 | firerun | 8 | - | 1×16 | 12×16 |
| 37 | Resurrect | 96 | 16 | ressur1 | 1 | - | 0×16 | 16×16 |
| 38 | BoneSpirit | 96 | 16 | sklball | 9 | - | 1×16 | 16×8 8×1 0×7 |
| 39 | RedPortal | 96 | 16 | rportal | 2 | - | 0×16 | 16×16 |
| 40 | DiabloApocalypseBoom | 160 | 48 | fireplar | 1 | MonsterOwned | 1×16 | 17×16 |
| 41 | BloodStarBlue | 96 | 16 | scubmisb | 1 | MonsterOwned | 0×16 | 16×16 |
| 42 | BloodStarBlueExplosion | 128 | 32 | scbsexpb | 1 | MonsterOwned | 0×16 | 6×16 |
| 43 | BloodStarYellow | 96 | 16 | scubmisc | 1 | MonsterOwned | 0×16 | 16×16 |
| 44 | BloodStarYellowExplosion | 128 | 32 | scbsexpc | 1 | MonsterOwned | 0×16 | 6×16 |
| 45 | BloodStarRed | 96 | 16 | scubmisd | 1 | MonsterOwned | 0×16 | 16×16 |
| 46 | BloodStarRedExplosion | 128 | 32 | scbsexpd | 1 | MonsterOwned | 0×16 | 6×16 |
First: `Arrow` (width 96, width2 16, name `arrows`, 1 frame, `NotAnimated`, all delays 0, all
lengths 16). Last: `BloodStarRedExplosion`.
---
## V4. OBJECTS — 109 objdat rows, object interaction, complete shrine list, books/tomes/elixirs
### V4.1 objdat.tsv rows 2-110 (109 rows, complete; loader `tables/objdat.cpp:259`)
Columns: `id | file | minLevel | maxLevel | levelType | theme | quest | flags | animDelay |
animLen | animWidth | selectionRegion` — `theme` = `theme_id` (`tables/objdat.h:18-37`),
`levelType` = `dungeon_type`, `flags` = ObjectFlag bitset.
#### objects rows 1–37 of 109
| # | id | file | lvls | levelType | theme | quest | flags | delay | len | width | region |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | OBJ_L1LIGHT | l1braz | 0-0 | DTYPE_CATHEDRAL | - | - | Animated,Solid,MissilesPassThrough | 1 | 26 | 64 | - |
| 2 | OBJ_L1LDOOR | l1doors | 0-0 | DTYPE_CATHEDRAL | - | - | Light,Trap | 1 | 0 | 64 | Bottom,Middle |
| 3 | OBJ_L1RDOOR | l1doors | 0-0 | DTYPE_CATHEDRAL | - | - | Light,Trap | 2 | 0 | 64 | Bottom,Middle |
| 4 | OBJ_SKFIRE | skulfire | 0-0 |  | THEME_SKELROOM | - | Animated,Solid,MissilesPassThrough | 2 | 11 | 96 | - |
| 5 | OBJ_LEVER | lever | 0-0 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 1 | 96 | Bottom |
| 6 | OBJ_CHEST1 | chest1 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 7 | OBJ_CHEST2 | chest2 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 8 | OBJ_CHEST3 | chest3 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 9 | OBJ_CANDLE1 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 10 | OBJ_CANDLE2 | candle2 | 0-0 |  | THEME_SHRINE | Q_PWATER | Animated,Solid,MissilesPassThrough,Light | 2 | 4 | 96 | - |
| 11 | OBJ_CANDLEO | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 12 | OBJ_BANNERL | banner | 0-0 |  | THEME_SKELROOM | - | Solid,MissilesPassThrough,Light | 2 | 0 | 96 | - |
| 13 | OBJ_BANNERM | banner | 0-0 |  | THEME_SKELROOM | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | - |
| 14 | OBJ_BANNERR | banner | 0-0 |  | THEME_SKELROOM | - | Solid,MissilesPassThrough,Light | 3 | 0 | 96 | - |
| 15 | OBJ_SKPILE | skulpile | 0-0 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 1 | 96 | - |
| 16 | OBJ_SKSTICK1 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 17 | OBJ_SKSTICK2 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 18 | OBJ_SKSTICK3 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 19 | OBJ_SKSTICK4 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 20 | OBJ_SKSTICK5 | l1braz | 0-0 |  | - | - | - | 0 | 0 | 0 | - |
| 21 | OBJ_CRUX1 | cruxsk1 | 0-0 |  | - | - | Solid,Light,Breakable | 1 | 15 | 96 | Bottom,Middle |
| 22 | OBJ_CRUX2 | cruxsk2 | 0-0 |  | - | - | Solid,Light,Breakable | 1 | 15 | 96 | Bottom,Middle |
| 23 | OBJ_CRUX3 | cruxsk3 | 0-0 |  | - | - | Solid,Light,Breakable | 1 | 15 | 96 | Bottom,Middle |
| 24 | OBJ_STAND | rockstan | 5-5 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | - |
| 25 | OBJ_ANGEL | angel | 0-0 |  | - | - | Solid,Light | 1 | 0 | 96 | - |
| 26 | OBJ_BOOK2L | book2 | 0-0 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 27 | OBJ_BCROSS | burncros | 0-0 |  | - | - | Animated,Solid | 0 | 10 | 160 | - |
| 28 | OBJ_NUDEW2R | nude2 | 0-0 |  | - | - | Animated,Solid,Light | 3 | 6 | 128 | - |
| 29 | OBJ_SWITCHSKL | switch4 | 16-16 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 30 | OBJ_TNUDEM1 | tnudem | 13-15 |  | - | Q_BUTCHER | Solid,Light | 1 | 0 | 128 | - |
| 31 | OBJ_TNUDEM2 | tnudem | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 2 | 0 | 128 | - |
| 32 | OBJ_TNUDEM3 | tnudem | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 3 | 0 | 128 | - |
| 33 | OBJ_TNUDEM4 | tnudem | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 4 | 0 | 128 | - |
| 34 | OBJ_TNUDEW1 | tnudew | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 1 | 0 | 128 | - |
| 35 | OBJ_TNUDEW2 | tnudew | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 2 | 0 | 128 | - |
| 36 | OBJ_TNUDEW3 | tnudew | 13-15 |  | THEME_TORTURE | Q_BUTCHER | Solid,Light | 3 | 0 | 128 | - |
| 37 | OBJ_TORTURE1 | tsoul | 13-15 |  | - | Q_BUTCHER | MissilesPassThrough,Light | 1 | 0 | 128 | - |
#### objects rows 38–74 of 109
| # | id | file | lvls | levelType | theme | quest | flags | delay | len | width | region |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 38 | OBJ_TORTURE2 | tsoul | 13-15 |  | - | Q_BUTCHER | MissilesPassThrough,Light | 2 | 0 | 128 | - |
| 39 | OBJ_TORTURE3 | tsoul | 13-15 |  | - | Q_BUTCHER | MissilesPassThrough,Light | 3 | 0 | 128 | - |
| 40 | OBJ_TORTURE4 | tsoul | 13-15 |  | - | Q_BUTCHER | MissilesPassThrough,Light | 4 | 0 | 128 | - |
| 41 | OBJ_TORTURE5 | tsoul | 13-15 |  | - | Q_BUTCHER | MissilesPassThrough,Light | 5 | 0 | 128 | - |
| 42 | OBJ_BOOK2R | book2 | 6-6 |  | - | - | Solid,MissilesPassThrough,Light | 4 | 0 | 96 | Bottom,Middle |
| 43 | OBJ_L2LDOOR | l2doors | 0-0 | DTYPE_CATACOMBS | - | - | Light,Trap | 1 | 0 | 64 | Bottom,Middle |
| 44 | OBJ_L2RDOOR | l2doors | 0-0 | DTYPE_CATACOMBS | - | - | Light,Trap | 2 | 0 | 64 | Bottom,Middle |
| 45 | OBJ_TORCHL | wtorch4 | 5-8 |  | - | - | Animated,MissilesPassThrough | 1 | 9 | 96 | - |
| 46 | OBJ_TORCHR | wtorch3 | 5-8 |  | - | - | Animated,MissilesPassThrough | 1 | 9 | 96 | - |
| 47 | OBJ_TORCHL2 | wtorch1 | 5-8 |  | - | - | Animated,MissilesPassThrough | 1 | 9 | 96 | - |
| 48 | OBJ_TORCHR2 | wtorch2 | 5-8 |  | - | - | Animated,MissilesPassThrough | 1 | 9 | 96 | - |
| 49 | OBJ_SARC | sarc | 1-4 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 5 | 128 | Bottom,Middle |
| 50 | OBJ_FLAMEHOLE | flame1 | 0-0 |  | - | - | MissilesPassThrough,Light | 1 | 20 | 96 | - |
| 51 | OBJ_FLAMELVR | lever | 0-0 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 2 | 96 | Bottom |
| 52 | OBJ_WATER | miniwatr | 0-0 |  | - | - | Animated,Solid,Light | 1 | 10 | 64 | - |
| 53 | OBJ_BOOKLVR | book1 | 0-0 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 54 | OBJ_TRAPL | traphole | 1-24 |  | - | - | MissilesPassThrough,Light | 1 | 0 | 64 | - |
| 55 | OBJ_TRAPR | traphole | 1-24 |  | - | - | MissilesPassThrough,Light | 2 | 0 | 64 | - |
| 56 | OBJ_BOOKSHELF | bcase | 0-0 |  | - | - | Solid,Light | 1 | 0 | 96 | - |
| 57 | OBJ_WEAPRACK | weapstnd | 0-0 |  | - | - | Solid,Light | 1 | 0 | 96 | - |
| 58 | OBJ_BARREL | barrel | 1-15 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 9 | 96 | Bottom,Middle |
| 59 | OBJ_BARRELEX | barrelex | 1-15 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 10 | 96 | Bottom,Middle |
| 60 | OBJ_SHRINEL | lshrineg | 0-0 |  | THEME_SHRINE | - | Light | 1 | 11 | 128 | Bottom,Middle |
| 61 | OBJ_SHRINER | rshrineg | 0-0 |  | THEME_SHRINE | - | Light | 1 | 11 | 128 | Bottom,Middle |
| 62 | OBJ_SKELBOOK | book2 | 0-0 |  | THEME_SKELROOM | - | Solid,MissilesPassThrough,Light | 4 | 0 | 96 | Bottom,Middle |
| 63 | OBJ_BOOKCASEL | bcase | 0-0 |  | THEME_LIBRARY | - | Solid,Light | 3 | 0 | 96 | Bottom,Middle |
| 64 | OBJ_BOOKCASER | bcase | 0-0 |  | THEME_LIBRARY | - | Solid,Light | 4 | 0 | 96 | Bottom,Middle |
| 65 | OBJ_BOOKSTAND | book2 | 0-0 |  | THEME_LIBRARY | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 66 | OBJ_BOOKCANDLE | candle2 | 0-0 |  | THEME_LIBRARY | - | Animated,Solid,MissilesPassThrough,Light | 2 | 4 | 96 | - |
| 67 | OBJ_BLOODFTN | bloodfnt | 0-0 |  | THEME_BLOODFOUNTAIN | - | Animated,Solid,MissilesPassThrough,Light | 2 | 10 | 96 | Bottom,Middle |
| 68 | OBJ_DECAP | decap | 13-15 |  | THEME_DECAPITATED | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom |
| 69 | OBJ_TCHEST1 | chest1 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 70 | OBJ_TCHEST2 | chest2 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 71 | OBJ_TCHEST3 | chest3 | 1-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom |
| 72 | OBJ_BLINDBOOK | book1 | 0-0 |  | - | Q_BLIND | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 73 | OBJ_BLOODBOOK | book1 | 0-0 |  | - | Q_BLOOD | Solid,MissilesPassThrough,Light | 4 | 0 | 96 | Bottom,Middle |
| 74 | OBJ_PEDESTAL | pedistl | 0-0 |  | - | Q_BLOOD | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
#### objects rows 75–109 of 109
| # | id | file | lvls | levelType | theme | quest | flags | delay | len | width | region |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 75 | OBJ_L3LDOOR | l3doors | 0-0 | DTYPE_CAVES | - | - | Light,Trap | 1 | 0 | 64 | Bottom,Middle |
| 76 | OBJ_L3RDOOR | l3doors | 0-0 | DTYPE_CAVES | - | - | Light,Trap | 2 | 0 | 64 | Bottom,Middle |
| 77 | OBJ_PURIFYINGFTN | pfountn | 0-0 |  | THEME_PURIFYINGFOUNTAIN | - | Animated,Solid,MissilesPassThrough,Light | 2 | 10 | 128 | Bottom,Middle |
| 78 | OBJ_ARMORSTAND | armstand | 0-0 |  | THEME_ARMORSTAND | - | Solid,Light | 1 | 0 | 96 | Bottom,Middle |
| 79 | OBJ_ARMORSTANDN | armstand | 0-0 |  | THEME_ARMORSTAND | - | Solid,Light | 2 | 0 | 96 | - |
| 80 | OBJ_GOATSHRINE | goatshrn | 0-0 |  | THEME_GOATSHRINE | - | Animated,Solid,MissilesPassThrough,Light | 2 | 10 | 96 | Bottom,Middle |
| 81 | OBJ_CAULDRON | cauldren | 13-15 |  | - | - | Solid,Light | 1 | 0 | 96 | Bottom,Middle |
| 82 | OBJ_MURKYFTN | mfountn | 0-0 |  | THEME_MURKYFOUNTAIN | - | Animated,Solid,MissilesPassThrough,Light | 2 | 10 | 128 | Bottom,Middle |
| 83 | OBJ_TEARFTN | tfountn | 0-0 |  | THEME_TEARFOUNTAIN | - | Animated,Solid,MissilesPassThrough,Light | 2 | 4 | 128 | Bottom,Middle |
| 84 | OBJ_ALTBOY | altboy | 0-0 |  | - | Q_BETRAYER | Solid,MissilesPassThrough,Light | 1 | 0 | 128 | - |
| 85 | OBJ_MCIRCLE1 | mcirl | 0-0 |  | - | Q_BETRAYER | MissilesPassThrough,Light | 1 | 0 | 96 | - |
| 86 | OBJ_MCIRCLE2 | mcirl | 0-0 |  | - | Q_BETRAYER | MissilesPassThrough,Light | 1 | 0 | 96 | - |
| 87 | OBJ_STORYBOOK | bkslbrnt | 0-0 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 88 | OBJ_STORYCANDLE | candle2 | 0-0 |  | - | Q_BETRAYER | Animated,Solid,MissilesPassThrough,Light | 2 | 4 | 96 | - |
| 89 | OBJ_STEELTOME | book1 | 0-0 |  | - | Q_WARLORD | Solid,MissilesPassThrough,Light | 4 | 0 | 96 | Bottom,Middle |
| 90 | OBJ_WARARMOR | armstand | 0-0 |  | - | Q_WARLORD | Solid,Light | 1 | 0 | 96 | Bottom,Middle |
| 91 | OBJ_WARWEAP | weapstnd | 0-0 |  | - | Q_WARLORD | Solid,Light | 1 | 0 | 96 | Bottom,Middle |
| 92 | OBJ_TBCROSS | burncros | 0-0 |  | THEME_BRNCROSS | - | Animated,Solid | 0 | 10 | 160 | - |
| 93 | OBJ_WEAPONRACK | weapstnd | 0-0 |  | THEME_WEAPONRACK | - | Solid,Light | 1 | 0 | 96 | Bottom,Middle |
| 94 | OBJ_WEAPONRACKN | weapstnd | 0-0 |  | THEME_WEAPONRACK | - | Solid,Light | 2 | 0 | 96 | - |
| 95 | OBJ_MUSHPATCH | mushptch | 0-0 |  | - | Q_MUSHROOM | Solid,MissilesPassThrough,Light,Trap | 1 | 0 | 96 | Bottom,Middle |
| 96 | OBJ_LAZSTAND | lzstand | 0-0 |  | - | Q_BETRAYER | Solid,Light | 1 | 0 | 128 | Bottom,Middle |
| 97 | OBJ_SLAINHERO | decap | 9-9 |  | - | - | Solid,MissilesPassThrough,Light | 2 | 0 | 96 | Bottom |
| 98 | OBJ_SIGNCHEST | chest3 | 0-0 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom |
| 99 | OBJ_BOOKSHELFR | bcase | 0-0 |  | - | - | Solid,Light | 2 | 0 | 96 | - |
| 100 | OBJ_POD | l6pod1 | 17-20 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 9 | 96 | Bottom,Middle |
| 101 | OBJ_PODEX | l6pod2 | 17-20 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 10 | 96 | Bottom,Middle |
| 102 | OBJ_URN | urn | 21-24 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 9 | 96 | Bottom,Middle |
| 103 | OBJ_URNEX | urnexpld | 21-24 |  | - | - | Solid,MissilesPassThrough,Light,Breakable | 1 | 10 | 96 | Bottom,Middle |
| 104 | OBJ_L5BOOKS | l5books | 21-24 |  | - | - | Solid,MissilesPassThrough,Light | 1 | 0 | 96 | Bottom,Middle |
| 105 | OBJ_L5CANDLE | l5light | 21-23 |  | - | - | Animated,Solid,MissilesPassThrough,Light | 2 | 4 | 96 | - |
| 106 | OBJ_L5LDOOR | l5door | 0-0 | DTYPE_CRYPT | - | - | Light,Trap | 1 | 0 | 64 | Bottom,Middle |
| 107 | OBJ_L5RDOOR | l5door | 0-0 | DTYPE_CRYPT | - | - | Light,Trap | 2 | 0 | 64 | Bottom,Middle |
| 108 | OBJ_L5LEVER | l5lever | 24-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 1 | 96 | Bottom |
| 109 | OBJ_L5SARC | l5sarco | 21-24 |  | - | - | Solid,MissilesPassThrough,Light,Trap | 1 | 5 | 128 | Bottom,Middle |
Row 1 = `OBJ_L1LIGHT` (file `l1braz`, DTYPE_CATHEDRAL, `Animated,Solid,MissilesPassThrough`,
delay 1, len 26, width 64); row 109 = `OBJ_L5SARC`.
- levelType histogram: =100, DTYPE_CATHEDRAL=3, DTYPE_CATACOMBS=2, DTYPE_CAVES=2, DTYPE_CRYPT=2.
- theme-tagged rows: THEME_ARMORSTAND=2, THEME_BLOODFOUNTAIN=1, THEME_BRNCROSS=1, THEME_DECAPITATED=1, THEME_GOATSHRINE=1, THEME_LIBRARY=4, THEME_MURKYFOUNTAIN=1, THEME_PURIFYINGFOUNTAIN=1, THEME_SHRINE=3, THEME_SKELROOM=5, THEME_TEARFOUNTAIN=1, THEME_TORTURE=6, THEME_WEAPONRACK=2 (consumed by the `Theme_*` placement in V11.3); quest-tagged rows: Q_BETRAYER=5, Q_BLIND=1, Q_BLOOD=2, Q_BUTCHER=12, Q_MUSHROOM=1, Q_PWATER=1, Q_WARLORD=3.
- flag vocabulary (all combos present in the file): ` — distinct atoms: Animated, Breakable, Light, MissilesPassThrough, Solid, Trap.
- shrine graphic rows: `OBJ_SHRINEL`, `OBJ_SHRINER`, `OBJ_GOATSHRINE`.
### V4.2 Object interaction algorithm (`Source/objects.cpp`, 4928 lines)
- Tick/anim driver `HandleObjects`: `Animated` objects advance `_oAnimCnt` against `animDelay`/
  `animLen` from objdat; `Solid` blocks movement, `MissilesPassThrough` skips missile collision,
  `Light` contributes a light radius, `Trap` arms the trap roll, `Breakable` enables `BreakObject`.
- Interact entry: `InteractObject` → `OperateObject(player, object, InteractMode)` switch on `_otype`
  (`_object_id`, objdat.h:42+) with `selectionRegion` cleared after use (see shrine `:2999`).
- **Doors**: `OperateBarricadeDoor`/prison/crypt doors — `SfxID::CryptDoorOpen|CryptDoorClose` vs
  `DoorOpen|DoorClose` by `isCrypt` (`objects.cpp:1766-1775`); lever `SfxID::OperateLever`
  (`:1828`) opening the Uber cow/Na-Krul door at `UberRow/UberCol` (`:1833`).
- **Containers**: chest `ChestOpen` (`:2024`) + `TriggerTrap` (`:2066`); mushroom patch (`:2093`);
  quest container (`:2124`); flame lever (`:2162`); sarcophagus `Sarcophagus` (`:2194`);
  pedestal blood/star `SpellPuddle`/`SpellBloodStar` (`:2229-2241`).
- **Books/lore**: `SfxID::QuestDone` on quest completion (`:1901`); `InitQTextMsg(TEXT_*)` shows the
  V8.1 scroll text via `interfac.cpp` (`scrlltxt` rows).
- **Shrine availability**: `shrineavail[]` (`objects.cpp:183-209`, `ShrineTypeAny|Single|Multi`;
  `Single` = at most one per level (Thaumaturgic, Solar), `Multi` = excluded from multi-pick
  (Spooky, Tainted)); pool size `gbIsHellfire ? NumberOfShrineTypes : 26` (`:1302`).
- **Shrine operation** — `OperateShrine(player, shrine, sType)` (`objects.cpp:2989-3105`):
  `CloseGoldDrop()`; **`DiabloGenerator rng(shrine._oRndSeed)`** so the effect is reproducible per
  level seed (`:2996`); `selectionRegion = None` (`:2998`); `PlaySfxLoc(sType)`; permanent animation
  (`:3001-3002`); then `switch (shrine._oVar1)` over `shrine_type` (`:3004-3080`, Murphys at `:3101`).
- Items dropped by broken objects go through `RndItem`-family with the object's level (`items.cpp`).
- Placement from objdat: `AddObj`/`AddObjPermanently`, `DRLG_*`/theme hooks, `dObject`,
  `ObjMem`/`Objects` container + `ActiveObjects` (`Source/objects.h:32-80`, `MAXOBJECTS 127`).
### V4.3 Complete shrine effect list — all 34 `shrine_type` values (`objects.cpp:65-100`)
Every entry below was read from the handler body at the cited line; `rng` = `DiabloGenerator`
seeded from `_oRndSeed`.
| # | shrine_type (objects.cpp:65+) | handler (objects.cpp) | exact effect |
|---|---|---|---|
| 1 | ShrineMysterious | OperateShrineMysterious (`:2250`) | -1 to Str, Mag, Dex and Vit, then **+6** to one of them chosen by `rng.generateRnd(4)`; `CheckStats`+`CalcPlrInv`; EMSG_SHRINE_MYSTERIOUS. |
| 2 | ShrineHidden | OperateShrineHidden (`:2282`) | If ≥1 equipped item: every equipped, non-indestructible item gets `_iDurability += 10` and `_iMaxDur += 10` (clamped); else a random attribute is rolled down; EMSG_SHRINE_HIDDEN. |
| 3 | ShrineGloomy | OperateShrineGloomy (`:2329`) | Every item in inventory: weapons (Sword/Axe/Bow/Mace/Staff) lose 1 from `_iMaxDam` (never below `_iMinDam`); armour loses 1 `_iMaxAC`; EMSG_SHRINE_GLOOMY. |
| 4 | ShrineWeird | OperateShrineWeird (`:2363`) | If a weapon is equipped (left or right hand, not a shield): `_iMaxDam`-- on inventory weapons; EMSG_SHRINE_WEIRD. |
| 5 | ShrineMagical | OperateShrineMagical (`:2392`) | `AddMissile(player→player, ManaShield, TARGET_MONSTERS, 0, 2*leveltype)` — casts Mana Shield on the player; EMSG_SHRINE_MAGICAL. |
| 6 | ShrineStone | OperateShrineStone (`:2410`) | Every staff in inventory: `_iCharges = _iMaxCharges` (full recharge); EMSG_SHRINE_STONE. |
| 7 | ShrineReligious | OperateShrineReligious (`:2427`) | Every item in inventory: `_iDurability = _iMaxDur` (full repair); EMSG_SHRINE_RELIGIOUS. |
| 8 | ShrineEnchanted | OperateShrineEnchanted (`:2439`) | Counts known spells; if >1 pick one at random via `rng.generateRnd(|SpellsData|)+1` and reduce it, then **+1 spell level** to every other known spell below `MaxSpellLevel`; EMSG_SHRINE_ENCHANTED. |
| 9 | ShrineThaumaturgic | OperateShrineThaumaturgic (`:2487`) | Rerolls `_oRndSeed` of every nearby unopened chest (`object.IsChest() && !canInteractWith()`) → the chests change contents; EMSG_SHRINE_THAUMATURGIC. |
| 10 | ShrineFascinating | OperateShrineCostOfWisdom(Firebolt) (`:2504`, called `:3044`) | `OperateShrineCostOfWisdom(SpellID::Firebolt, EMSG_SHRINE_FASCINATING)`: raises Firebolt to max level, then subtracts `penalty = _pMaxManaBase/10` (10%) from `_pMaxManaBase`/`_pMaxMana` (`:2504-2545`). |
| 11 | ShrineCryptic | OperateShrineCryptic (`:2546`) | `_pMana = _pMaxMana; _pManaBase = _pMaxManaBase` (refill mana); EMSG_SHRINE_CRYPTIC. |
| 12 | ShrineMagicaL2 | OperateShrineMagical — alias (`:2392`, case `:3040`) | Same handler as ShrineMagical (second graphic variant). |
| 13 | ShrineEldritch | OperateShrineEldritch (`:2569`) | Every non-Misc item (inventory + belt): potions `IMISC_HEAL`/`IMISC_MANA` are replaced/re-rolled via their `_iSeed`; EMSG_SHRINE_ELDRITCH. |
| 14 | ShrineEerie | OperateShrineEerie (`:2602`) | `ModifyPlrMag(player, 2)` — +2 Magic; EMSG_SHRINE_EERIE. |
| 15 | ShrineDivine | OperateShrineDivine (`:2621`) | `AddMissile(Phasing)`; then if `currlevel < 4` … full refill of mana **and** HP (`_pMana=_pMaxMana`, `_pHitPoints=_pMaxHP`, bases too); EMSG_SHRINE_DIVINE. |
| 16 | ShrineHoly | OperateShrineHoly (`:2644`) | `AddMissile(player→player, Phasing, TARGET_MONSTERS, 0, 2*leveltype)` — grants the Phasing/teleport effect; EMSG_SHRINE_HOLY. |
| 17 | ShrineSacred | OperateShrineCostOfWisdom(ChargedBolt) (`:2504`, case `:3052`) | `OperateShrineCostOfWisdom(SpellID::ChargedBolt, EMSG_SHRINE_SACRED)`. |
| 18 | ShrineSpiritual | OperateShrineSpiritual (`:2654`) | `MakeGoldStack(goldItem, (5 * leveltype) + rng.generateRnd(10 * leveltype))` added directly to the player's inventory (`_pNumInv++`, `_pGold += _ivalue`); EMSG_SHRINE_SPIRITUAL. |
| 19 | ShrineSpooky | OperateShrineSpooky (`:2673`) | First visit (SPOOKY1 message) simply refills HP+mana (`:2673-2690`); second behaviour EMSG_SHRINE_SPOOKY2. |
| 20 | ShrineAbandoned | OperateShrineAbandoned (`:2692`) | `ModifyPlrDex(player, 2)` — +2 Dexterity; EMSG_SHRINE_ABANDONED. |
| 21 | ShrineCreepy | OperateShrineCreepy (`:2705`) | `ModifyPlrStr(player, 2)` — +2 Strength; EMSG_SHRINE_CREEPY. |
| 22 | ShrineQuiet | OperateShrineQuiet (`:2718`) | `ModifyPlrVit(player, 2)` — +2 Vitality; EMSG_SHRINE_QUIET. |
| 23 | ShrineSecluded | OperateShrineSecluded (`:2731`) | Reveals the whole map: `for x,y in DMAXX/DMAXY: UpdateAutomapExplorer({x,y}, MAP_EXP_SHRINE)`; EMSG_SHRINE_SECLUDED. |
| 24 | ShrineOrnate | OperateShrineCostOfWisdom(HolyBolt) (`:2504`, case `:3067`) | `OperateShrineCostOfWisdom(SpellID::HolyBolt, EMSG_SHRINE_ORNATE)`. |
| 25 | ShrineGlimmering | OperateShrineGlimmering (`:2743`) | Identifies every magic item in inventory (`_iMagical != NORMAL && !_iIdentified ⇒ _iIdentified = true`); EMSG_SHRINE_GLIMMERING. |
| 26 | ShrineTainted | OperateShrineTainted (`:2760`) | Random ± attribute shuffle: `ModifyPlrStr(v1)`, `Mag(v2)`, `Dex(v3)`, `Vit(v4)` from `rng` (TAINTED1 when it would kill, else TAINTED2). |
| 27 | ShrineOily | OperateShrineOily (`:2794`) | Class-scaled stat gain: Warrior +2 Str, Rogue +2 Dex, Sorcerer +2 Mag, Barbarian +2 Vit (twice), Monk +1 Str +1 Dex, Bard +1 Dex +1 Mag — then **a `FireWall` missile is spawned targeting the player** (`TARGET_PLAYERS`, dam `2*currlevel+2`) (`:2824-2839`); EMSG_SHRINE_OILY. |
| 28 | ShrineGlowing | OperateShrineGlowing (`:2841`) | `ModifyPlrMag(player, min(_pExperience/1000, 5))`, then `_pExperience *= 0.95` when exp > 5000 (floor 0); EMSG_SHRINE_GLOWING. |
| 29 | ShrineMendicant | OperateShrineMendicant (`:2861`) | `gold = _pGold / 2` — loses half the player's gold (converted to a dropped pile); EMSG_SHRINE_MENDICANT. |
| 30 | ShrineSparkling | OperateShrineSparkling (`:2880`) | `AddMissile(..., spawnPosition → player, ...)` gold-item spawn variant: drops gold around the shrine; EMSG_SHRINE_SPARKLING. |
| 31 | ShrineTown | OperateShrineTown (`:2907`) | Drops a town-portal scroll at the shrine (`AddMissile(TownPortal ...)` path, `:2907-2924`); EMSG_SHRINE_TOWN. |
| 32 | ShrineShimmering | OperateShrineShimmering (`:2925`) | `_pMana = _pMaxMana; _pManaBase = _pMaxManaBase` (mana refill only); EMSG_SHRINE_SHIMMERING. |
| 33 | ShrineSolar | OperateShrineSolar (`:2938`) | Time-of-day based: `hour >= 20 || hour < 4` ⇒ +2 Vit; `hour >= 18` ⇒ +2 Mag; … four branches with EMSG_SHRINE_SOLAR1..4 (`:2938-2963`). |
| 34 | ShrineMurphys | OperateShrineMurphys (`:2965`, case `:3101`) | For each equipped non-empty item, `rng.flipCoin(3)` ⇒ damage its durability (`DUR_INDESTRUCTIBLE` exempt); EMSG_SHRINE_MURPHYS. |
Only the first 26 are used in Diablo (`objects.cpp:1302`); rows 27-34 (Oily, Glowing, Mendicant,
Sparkling, Town, Shimmering, Solar, Murphy's) are Hellfire additions. `IsShrine()` predicate
(`objects.h:249`).
### V4.4 Book / tome / elixir / rune / oil effects
| miscId (itemdat col 20) | V1.2 rows | effect + citation |
|---|---|---|
- **BOOK** (4 rows: `IDI_BOOK1..4` + the 4 generic `Book of …` rows): `GetBookSpell(item, lvl)`
  (`items.cpp:613-650`): `lvl = max(lvl,1)`; `rv = GenerateRnd(|SpellsData|) + 1`; walk `SpellID`
  from Firebolt, counting only spells with `GetSpellBookLevel(s) != -1 && lvl >= sBookLvl`;
  in **single player** skip `Resurrect→Telekinesis` (`:634-638`) and `HealOther→BloodStar`
  (`:639-643`); `gbIsSpawn` clamps `lvl` to 5 (`:620`). Reading it teaches the spell
  (`AddSpellLevel`/`pMemSpells`, `spells.cpp`) and plays `SfxID::Book`.
  Note the base table's book names are `"Book of "` with an empty suffix (`"Book of "` trailing  space, rows 115, 116, 117, 118) — the spell name is appended at runtime, so the Go table must keep the  literal string verbatim and format it.
- **TO_* tomes (0 rows)**:  — permanent stat/exp/portal effects in `UseItem` (`items.cpp` misc switch).
- **Elixirs (0 rows)**:  — permanent +1 to Strength/Magic/Dexterity/Vitality (`UseItem`).
- **Oils (5 rows: Blacksmith Oil, Blacksmith Oil, Oil of Accuracy, Oil of Sharpness, Oil)** — Hellfire-only (`IsItemAvailable` blocks `i < IDI_OIL || i > IDI_GREYSUIT`
  and the explicit `(i < 83 || i > 86)` oil-range exclusion). Applied through Griswold;
  `CURSOR_OIL` cursor (V12.1).
- **Runes (%d rows: %s)** — Hellfire-only (`(i < 161 || i > 165)` exclusion
  cast the matching `RuneOf*` missile (`missiles.cpp:1248-1273`) and are consumed;
  `Item::isRune()/isRuneOf()` (`items.h:400-420`); cursor `CURSOR_OIL`-family/`RUNE_*` graphics.
- **POTIONS**: 0 rows with `miscId` starting `POTION`; heal/mana/rejuvenation/resurrect use the
  class restore divisors from V10.3 (`itmRestoreLife/itmRestoreMana`) and `SpellID::HealOther`
  divisor `healOtherRestoreLife`; `SfxID::Potion*`, `GetGoldCursor`-style `SetPotionGraphicLS`.
- **STAFF**: 7 rows (`IMISC_STAFF`/`SHORT_STAFF_OF`…): spell from the `spell` column when set,
  otherwise `GetStaffSpell` (`items.cpp:1236+`); charges via `IPL_CHARGES` affix
  (`SaveItemAffix` `:1048-1061`), `_iMaxCharges` roll `staffMin/staffMax` (V7.1),
  `SpellID::StaffRecharge` refills (`spells.cpp`), `CURSOR_RECHARGE`.
---
## V5. TOWN & NPCs — 12 towner rows, dialog matrix, store catalogs, cow content
### V5.1 towners.tsv rows 2-13 (12 rows; loader `tables/townerdat.cpp:100`)
Columns: `type | name | position_x | position_y | direction | animWidth | animPath | animFrames
| animDelay | gossipTexts | animOrder`. Enum `_talker_id` (`Source/towners.h:26-40`): 13 types,
`TOWN_COW` repeated 3× as instances. Interaction/init functions are **compiled in**, not in the
TSV: `TownerTalkFns[]` (`towners.cpp:690-700`) — e.g. `{TOWN_COWFARM, InitCowFarmer,
TalkToCowFarmer}` (`:697`); per-type presence `IsTownerPresent` (`:735-750`); short names for
Lua/mods `TownerShortNames` (`:710-724`).
| # | type | name | pos | dir | animWidth | animFrames | animDelay | animPath | #gossip | gossip ids |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | TOWN_SMITH | Griswold the Blacksmith | (62,63) | SouthWest | 96 | 16 | 3 | towners\smith\smithn | 11 | `TEXT_GRISWOLD2`, `TEXT_GRISWOLD3`, `TEXT_GRISWOLD4`, `TEXT_GRISWOLD5`, `TEXT_GRISWOLD6`, `TEXT_GRISWOLD7`, `TEXT_GRISWOLD8`, `TEXT_GRISWOLD9`, `TEXT_GRISWOLD10`, `TEXT_GRISWOLD12`, `TEXT_GRISWOLD13` |
| 2 | TOWN_HEALER | Pepin the Healer | (55,79) | SouthEast | 96 | 20 | 6 | towners\healer\healer | 9 | `TEXT_PEPIN2`, `TEXT_PEPIN3`, `TEXT_PEPIN4`, `TEXT_PEPIN5`, `TEXT_PEPIN6`, `TEXT_PEPIN7`, `TEXT_PEPIN9`, `TEXT_PEPIN10`, `TEXT_PEPIN11` |
| 3 | TOWN_DEADGUY | Wounded Townsman | (24,32) | North | 96 | 8 | 6 | towners\butch\deadguy | 0 | - |
| 4 | TOWN_TAVERN | Ogden the Tavern owner | (55,62) | SouthWest | 96 | 16 | 3 | towners\twnf\twnfn | 8 | `TEXT_OGDEN2`, `TEXT_OGDEN3`, `TEXT_OGDEN4`, `TEXT_OGDEN5`, `TEXT_OGDEN6`, `TEXT_OGDEN8`, `TEXT_OGDEN9`, `TEXT_OGDEN10` |
| 5 | TOWN_STORY | Cain the Elder | (62,71) | South | 96 | 25 | 3 | towners\strytell\strytell | 9 | `TEXT_STORY2`, `TEXT_STORY3`, `TEXT_STORY4`, `TEXT_STORY5`, `TEXT_STORY6`, `TEXT_STORY7`, `TEXT_STORY9`, `TEXT_STORY10`, `TEXT_STORY11` |
| 6 | TOWN_DRUNK | Farnham the Drunk | (71,84) | South | 96 | 18 | 3 | towners\drunk\twndrunk | 11 | `TEXT_FARNHAM2`, `TEXT_FARNHAM3`, `TEXT_FARNHAM4`, `TEXT_FARNHAM5`, `TEXT_FARNHAM6`, `TEXT_FARNHAM8`, `TEXT_FARNHAM9`, `TEXT_FARNHAM10`, `TEXT_FARNHAM11`, `TEXT_FARNHAM12`, `TEXT_FARNHAM13` |
| 7 | TOWN_WITCH | Adria the Witch | (80,20) | South | 96 | 19 | 6 | towners\townwmn1\witch | 11 | `TEXT_ADRIA2`, `TEXT_ADRIA3`, `TEXT_ADRIA4`, `TEXT_ADRIA5`, `TEXT_ADRIA6`, `TEXT_ADRIA7`, `TEXT_ADRIA8`, `TEXT_ADRIA9`, `TEXT_ADRIA10`, `TEXT_ADRIA12`, `TEXT_ADRIA13` |
| 8 | TOWN_BMAID | Gillian the Barmaid | (43,66) | South | 96 | 18 | 6 | towners\townwmn1\wmnn | 8 | `TEXT_GILLIAN2`, `TEXT_GILLIAN3`, `TEXT_GILLIAN4`, `TEXT_GILLIAN5`, `TEXT_GILLIAN6`, `TEXT_GILLIAN7`, `TEXT_GILLIAN9`, `TEXT_GILLIAN10` |
| 9 | TOWN_PEGBOY | Wirt the Peg-legged boy | (11,53) | South | 96 | 20 | 6 | towners\townboy\pegkid1 | 10 | `TEXT_WIRT2`, `TEXT_WIRT3`, `TEXT_WIRT4`, `TEXT_WIRT5`, `TEXT_WIRT6`, `TEXT_WIRT7`, `TEXT_WIRT8`, `TEXT_WIRT9`, `TEXT_WIRT11`, `TEXT_WIRT12` |
| 10 | TOWN_COW | Cow | (58,16) | SouthWest | 128 | 12 | 3 | (shared CowSprites) | 0 | - |
| 11 | TOWN_COW | Cow | (56,14) | NorthWest | 128 | 12 | 3 | (shared CowSprites) | 0 | - |
| 12 | TOWN_COW | Cow | (59,20) | North | 128 | 12 | 3 | (shared CowSprites) | 0 | - |
### V5.1b animOrder sequences (all 12 rows, verbatim, run-length compressed)
- row 1 `TOWN_SMITH` (Griswold the Blacksmith): **130 entries** — 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×2 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×2 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×2 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×2 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×2 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×2 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×2 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×2 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×2 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×2 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1 0×15 1×1 2×1 3×1
- row 2 `TOWN_HEALER` (Pepin the Healer): **108 entries** — 0×1 1×1 2×2 1×1 0×1 19×1 18×2 19×1 0×1 1×1 2×2 1×1 0×1 19×1 18×2 19×1 0×1 1×1 2×2 1×1 0×1 19×1 18×2 19×1 0×1 1×1 2×2 1×1 0×1 19×1 18×2 19×1 0×1 1×1 2×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1 14×1 13×1 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1 14×1 13×1 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1 16×1 17×1 18×1 19×1
- row 3 `TOWN_DEADGUY` (Wounded Townsman): **0 entries** — 
- row 4 `TOWN_TAVERN` (Ogden the Tavern owner): **111 entries** — 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×2 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×1 1×1 0×1 15×1 14×1 13×2 14×1 15×1 0×1 1×1 2×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1
- row 5 `TOWN_STORY` (Cain the Elder): **59 entries** — 0×2 24×2 23×1 22×1 21×1 20×1 19×1 18×1 17×1 16×1 15×1 14×1 15×1 16×1 17×1 18×1 19×1 20×1 21×1 22×1 23×1 24×3 0×3 24×1 0×1 1×1 2×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×1 13×1 12×1 11×1 10×1 9×1 8×1 7×1 6×1 5×1 4×1 3×1 2×1 1×1 0×1
- row 6 `TOWN_DRUNK` (Farnham the Drunk): **54 entries** — 0×3 1×1 2×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×4 11×1 12×1 13×1 14×1 15×1 16×1 17×2 0×3 17×1 16×1 15×1 14×1 13×1 12×1 11×1 10×1 9×1 10×1 11×1 12×1 13×1 14×1 15×1 16×1 17×1 0×1 1×1 2×1 3×1 4×3 3×1 2×1 1×1
- row 7 `TOWN_WITCH` (Adria the Witch): **141 entries** — 3×3 4×1 5×3 4×1 3×1 14×1 13×1 12×3 13×1 14×1 3×1 4×1 5×3 4×1 3×3 4×1 5×3 4×1 3×1 14×1 13×1 12×3 13×1 14×1 3×1 4×1 5×3 4×1 3×3 4×1 5×3 4×1 3×1 14×1 13×1 12×3 13×1 14×1 3×1 4×1 5×3 4×1 3×1 2×1 1×1 0×1 18×1 17×1 18×1 0×1 1×1 0×1 18×1 17×1 18×1 0×1 1×1 0×1 1×1 2×1 3×1 4×1 5×1 6×1 7×1 8×1 9×1 10×1 11×1 12×1 13×1 14×3 13×1 12×4 13×1 14×3 13×1 12×1 11×3 10×1 9×3 8×1 7×1 8×1 9×2 10×1 11×1 12×1 13×1 14×1 15×1 16×1 17×1 18×1 0×1 1×1 0×1 18×1 17×1 18×1 0×1 1×1 0×1 1×1 2×1
- row 8 `TOWN_BMAID` (Gillian the Barmaid): **0 entries** — 
- row 9 `TOWN_PEGBOY` (Wirt the Peg-legged boy): **0 entries** — 
- row 10 `TOWN_COW` (Cow): **0 entries** — 
- row 11 `TOWN_COW` (Cow): **0 entries** — 
- row 12 `TOWN_COW` (Cow): **0 entries** — 
### V5.2 quest_dialog.tsv rows 2-11 (10 towner rows × 18 quest columns, every cell)
Header: `towner_type, Q_ROCK, Q_MUSHROOM, Q_GARBUD, Q_ZHAR, Q_VEIL, Q_DIABLO, Q_BUTCHER, Q_LTBANNER, Q_BLIND, Q_BLOOD, Q_ANVIL, Q_WARLORD, Q_SKELKING, Q_PWATER, Q_SCHAMB, Q_BETRAYER, Q_GRAVE, Q_TRADER`. Loader `tables/townerdat.cpp:135`.
`TEXT_NONE` ⇒ that NPC has no line for that quest state. Written to the runtime matrix via
`SetTownerQuestDialog` (see `quests.cpp:61-62` forcing `TOWN_WITCH/Q_MUSHROOM = TEXT_MUSH9` and
`TOWN_HEALER/Q_MUSHROOM = TEXT_NONE` at init).
- **row 1 `TOWN_SMITH`** — ROCK: `TEXT_INFRA6`, MUSHROOM: `TEXT_MUSH6`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL5`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH5`, LTBANNER: `TEXT_BANNER6`, BLIND: `TEXT_BLIND5`, BLOOD: `TEXT_BLOOD5`, ANVIL: `TEXT_ANVIL6`, WARLORD: `TEXT_WARLRD5`, SKELKING: `TEXT_KING7`, PWATER: `TEXT_POISON7`, SCHAMB: `TEXT_BONE5`, BETRAYER: `TEXT_VILE9`, GRAVE: `TEXT_GRAVE2`, TRADER: `TEXT_NONE`
- **row 2 `TOWN_HEALER`** — ROCK: `TEXT_INFRA3`, MUSHROOM: `TEXT_NONE`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL3`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH3`, LTBANNER: `TEXT_BANNER4`, BLIND: `TEXT_BLIND3`, BLOOD: `TEXT_BLOOD3`, ANVIL: `TEXT_ANVIL3`, WARLORD: `TEXT_WARLRD3`, SKELKING: `TEXT_KING5`, PWATER: `TEXT_POISON4`, SCHAMB: `TEXT_BONE3`, BETRAYER: `TEXT_VILE7`, GRAVE: `TEXT_GRAVE3`, TRADER: `TEXT_NONE`
- **row 3 `TOWN_DEADGUY`** — ROCK: `TEXT_NONE`, MUSHROOM: `TEXT_NONE`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_NONE`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_NONE`, LTBANNER: `TEXT_NONE`, BLIND: `TEXT_NONE`, BLOOD: `TEXT_NONE`, ANVIL: `TEXT_NONE`, WARLORD: `TEXT_NONE`, SKELKING: `TEXT_NONE`, PWATER: `TEXT_NONE`, SCHAMB: `TEXT_NONE`, BETRAYER: `TEXT_NONE`, GRAVE: `TEXT_NONE`, TRADER: `TEXT_NONE`
- **row 4 `TOWN_TAVERN`** — ROCK: `TEXT_INFRA2`, MUSHROOM: `TEXT_MUSH2`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL2`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH2`, LTBANNER: `TEXT_NONE`, BLIND: `TEXT_BLIND2`, BLOOD: `TEXT_BLOOD2`, ANVIL: `TEXT_ANVIL2`, WARLORD: `TEXT_WARLRD2`, SKELKING: `TEXT_KING3`, PWATER: `TEXT_POISON2`, SCHAMB: `TEXT_BONE2`, BETRAYER: `TEXT_VILE4`, GRAVE: `TEXT_GRAVE5`, TRADER: `TEXT_NONE`
- **row 5 `TOWN_STORY`** — ROCK: `TEXT_INFRA1`, MUSHROOM: `TEXT_MUSH1`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL1`, DIABLO: `TEXT_VILE3`, BUTCHER: `TEXT_BUTCH1`, LTBANNER: `TEXT_BANNER1`, BLIND: `TEXT_BLIND1`, BLOOD: `TEXT_BLOOD1`, ANVIL: `TEXT_ANVIL1`, WARLORD: `TEXT_WARLRD1`, SKELKING: `TEXT_KING1`, PWATER: `TEXT_POISON1`, SCHAMB: `TEXT_BONE1`, BETRAYER: `TEXT_VILE2`, GRAVE: `TEXT_GRAVE6`, TRADER: `TEXT_NONE`
- **row 6 `TOWN_DRUNK`** — ROCK: `TEXT_INFRA8`, MUSHROOM: `TEXT_MUSH7`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL6`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH6`, LTBANNER: `TEXT_BANNER7`, BLIND: `TEXT_BLIND6`, BLOOD: `TEXT_BLOOD6`, ANVIL: `TEXT_ANVIL8`, WARLORD: `TEXT_WARLRD6`, SKELKING: `TEXT_KING8`, PWATER: `TEXT_POISON8`, SCHAMB: `TEXT_BONE6`, BETRAYER: `TEXT_VILE10`, GRAVE: `TEXT_GRAVE7`, TRADER: `TEXT_NONE`
- **row 7 `TOWN_WITCH`** — ROCK: `TEXT_INFRA9`, MUSHROOM: `TEXT_MUSH9`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL7`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH7`, LTBANNER: `TEXT_BANNER8`, BLIND: `TEXT_BLIND7`, BLOOD: `TEXT_BLOOD7`, ANVIL: `TEXT_ANVIL9`, WARLORD: `TEXT_WARLRD7`, SKELKING: `TEXT_KING9`, PWATER: `TEXT_POISON9`, SCHAMB: `TEXT_BONE7`, BETRAYER: `TEXT_VILE11`, GRAVE: `TEXT_GRAVE1`, TRADER: `TEXT_NONE`
- **row 8 `TOWN_BMAID`** — ROCK: `TEXT_INFRA4`, MUSHROOM: `TEXT_MUSH5`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL4`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH4`, LTBANNER: `TEXT_BANNER5`, BLIND: `TEXT_BLIND4`, BLOOD: `TEXT_BLOOD4`, ANVIL: `TEXT_ANVIL4`, WARLORD: `TEXT_WARLRD4`, SKELKING: `TEXT_KING6`, PWATER: `TEXT_POISON6`, SCHAMB: `TEXT_BONE4`, BETRAYER: `TEXT_VILE8`, GRAVE: `TEXT_GRAVE8`, TRADER: `TEXT_NONE`
- **row 9 `TOWN_PEGBOY`** — ROCK: `TEXT_INFRA10`, MUSHROOM: `TEXT_MUSH13`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_VEIL8`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_BUTCH8`, LTBANNER: `TEXT_BANNER9`, BLIND: `TEXT_BLIND8`, BLOOD: `TEXT_BLOOD8`, ANVIL: `TEXT_ANVIL10`, WARLORD: `TEXT_WARLRD8`, SKELKING: `TEXT_KING10`, PWATER: `TEXT_POISON10`, SCHAMB: `TEXT_BONE8`, BETRAYER: `TEXT_VILE12`, GRAVE: `TEXT_GRAVE9`, TRADER: `TEXT_NONE`
- **row 10 `TOWN_COW`** — ROCK: `TEXT_NONE`, MUSHROOM: `TEXT_NONE`, GARBUD: `TEXT_NONE`, ZHAR: `TEXT_NONE`, VEIL: `TEXT_NONE`, DIABLO: `TEXT_NONE`, BUTCHER: `TEXT_NONE`, LTBANNER: `TEXT_NONE`, BLIND: `TEXT_NONE`, BLOOD: `TEXT_NONE`, ANVIL: `TEXT_NONE`, WARLORD: `TEXT_NONE`, SKELKING: `TEXT_NONE`, PWATER: `TEXT_NONE`, SCHAMB: `TEXT_NONE`, BETRAYER: `TEXT_NONE`, GRAVE: `TEXT_NONE`, TRADER: `TEXT_NONE`
### V5.3 Store catalogs and price rules
TalkID set — 24 states (`Source/stores.h:44-70`): `None, Smith, SmithBuy, SmithSell, SmithRepair,
Witch, WitchBuy, WitchSell, WitchRecharge, NoMoney, NoRoom, Confirm, Boy, BoyBuy, Healer,
Storyteller, HealerBuy, StorytellerIdentify, SmithPremiumBuy, Gossip, StorytellerIdentifyShow,
Tavern, Drunk, Barmaid`. Sell snapshot buffers `PlayerItemIndexes[48]`/`PlayerItems[48]`
(`stores.h:70-76`). Vendor→NPC names: Boy→`wirt`, Healer→`pepin`, Storyteller→`cain`,
Tavern→`ogden`, Drunk→`farnham`, Barmaid→`gillian` (`stores.cpp:2346-2356`); extra dialog
options are injected into free even lines 10..last-selectable (`stores.cpp:2387-2404`).
| store | container + caps | generation | pricing / behaviour |
|---|---|---|---|
| **Griswold basic** | `SmithItems` ≤ 19 (`NumSmithBasicItems`) / 24 (`…Hf`) — `stores.h:27-28` | `SpawnSmith(lvl)` `items.cpp:4396-4425`: `iCnt = RandomIntBetween(10, maxItems)`; per item `newItem._iSeed = AdvanceRndSeed(); SetRndSeed(seed); RndSmithItem(player, lvl)` (`:1897-1900` = `RndVendorItem<SmithItemOk,true>`); reject `_iIvalue > MaxVendorValue` 140000 / `MaxVendorValueHf` 200000 (`items.h:36-37`); `_iCreateInfo = lvl \| CF_SMITH`, `_iIdentified = true`; `SortVendor(list, PinnedItemCount=0)` (`:1902-1908`) | buy at `_iIvalue`; sell `_iIvalue/4` (min 1, `stores.cpp:565,582`); repair `SmithRepairItem(item._iIvalue)` restores `_iDurability = _iMaxDur` (`stores.cpp:1837`) |
| **Griswold premium** | `PremiumItems` ≤ `NumSmithItemsHf` (`stores.h:88`); `PremiumItemCount` reset `stores.cpp:2150`; 6 slots (Diablo) / 15 (Hellfire) | `SpawnPremium(player)` `items.cpp:4436-4465`; per slot `SpawnOnePremium(PremiumItems[idx], plvl, player)` with `plvl = (gbIsHellfire ? itemLevelAddHf : itemLevelAdd)[idx] + player.level` (`ReplacePremium` `:4427-4434`); item via `RndPremiumItem(player, plvl/4, plvl)` (`:1932`) + `GetItemBonus(item, plvl/2, plvl, onlygood=true, allowspells=!gbIsHellfire)` (`:1956`) | `SmithPremiumBuy` flow `stores.cpp:489-499, 1400-1410`; refresh timer per slot; Hellfire allows spell charges on premium (`allowspells` false ⇒ no staff spells) |
| **Adria (witch)** | `WitchItems` ≤ 17 / 24 HF (`stores.h:38-39`) | `SpawnWitch` `items.cpp:4467-4528`: pinned `IDI_MANA, IDI_FULLMANA, IDI_PORTAL` (`PinnedItemCount`), HF additionally pins 0..4 of `IDI_BOOK1..4` (`pinnedBookCount = RandomIntLessThan(4)`); others `RndWitchItem<WitchItemOk>` (`:2034`); `maxlvl = 2*lvl` when `GenerateRnd(100) <= 5` **or** the item is a staff, then `GetItemBonus(item, maxlvl/2, maxlvl, true, true)`; value cap `MaxVendorValue(Hf)`; `CF_WITCH`; `SortVendor` after pinned | **recharge** = `WitchRecharge` TalkID (staff charges restored for `_iIvalue`-scaled cost); sells scrolls/books/potions + magic items |
| **Pepin (healer)** | `HealerItems` ≤ 17 / 19 HF (`stores.h:33-34`) | `SpawnHealer(lvl)` `items.cpp:4649-4673`: pinned `IDI_HEAL, IDI_FULLHEAL, IDI_RESURRECT` (`PinnedItemTypes`), MP pins the 3rd (`NumHealerPinnedItemsMp`); others `RndHealerItem<HealerItemOk>` (`:2073+`, misc potions only); `itemCount = RandomIntBetween(10, max)`; `CF_HEALER` | Talk free full heal (`HealPlayer()` `stores.cpp:1021-1030`, sets HP/mana to max, `SfxID::CastHealing`); buying a pinned slot **re-seeds** it (`HealerBuyItem` `:1737-1763`: SP re-seed when `idx < 2`, MP when `idx < 3`; erase beyond pinned) |
| **Wirt (boy)** | exactly one item `BoyItem`; `BoyItemLevel` cooldown | `SpawnBoy(playerLevel)` `items.cpp:4533-4646`: skip when `BoyItemLevel >= lvl/2 && !BoyItem.isEmpty()`; stats = `max(GetMaximumAttributeValue(a), _p<a>)*1.2`; roll `RndBoyItem<PremiumItemOk>` (`:2039`) + `GetItemBonus(item, lvl, 2*lvl, true, true)`; re-roll while `_iIvalue > MaxBoyValueHf` (`items.h:39`) **or** reqs exceed the scaled stats **or** `_iIvalue < mostValuablePlayerItem._iIvalue * 4/5` (`:4594-4627`); class-aware bans (`count < 200`) set `ivalue = INT_MAX` for off-class types — Warrior bans Bow/Staff, Rogue bans Sword/Staff/Axe/Mace/Shield, Sorcerer bans Staff/Axe/Bow/Mace, Monk bans Bow/MediumArmor/Shield/Mace, Bard bans Axe/Mace/Staff, Barbarian bans Bow/Staff (`:4603-4626`); loop cap `count < 250`; `CF_BOY`; `BoyItemLevel = lvl/2` | **50-gold peek fee** `TakePlrsMoney(50)` (`stores.cpp:1706`) then `BoyBuy`; price `= _iIvalue − _iIvalue/4` in Hellfire, `= _iIvalue + _iIvalue/2` in Diablo (`SStartBoyBuy` `:1003-1005`, `BoyBuyEnter` `:1770-1776`); buy via `BoyBuyItem` (`:1725-1732`) |
| **Cain (storyteller)** | identify service only | `TalkID::Storyteller` → `StartStoryteller` (`stores.cpp` ~`:1070+`); `StorytellerIdentify`/`StorytellerIdentifyShow` states | identify: cost scaled from `_iIvalue` (`StorytellerIdentifyItem` `stores.cpp:1797+`); gossip = the quest `TEXT_*` matrix (V5.2) |
| **Ogden / Farnham / Gillian** | gossip only | random `gossip` topic per NPC chosen on town load (V5.1 `gossipTexts` lists) | `StartTavern`/`StartDrunk`/`StartBarmaid` (`stores.cpp:2380-2385`); Farnham's news can open the dark-portal/quest hints |
Shared mechanics: `AddSText`/`AddSLine`/`AddSTextVal` 20-line screen model (`NumStoreLines`),
`ScrollPos`/`CurrentTextLine`/`OldTextLine` paging with `idx = ScrollPos + ((line − PreviousScrollPos)/4)`
(`stores.cpp:1340,1357,1382,1398`); `PlayerCanAfford`/`NoMoney`, `StoreAutoPlace`/`NoRoom`,
`Confirm` yes/no (`StartStore` re-entry at `:504,911-929,1192,1293`); `StoreGoldFit` requires the
price to fit `w*h*MaxGold` plus `RoomForGold()` (`stores.cpp:1414-1425`); `SpawnStores()`
(`:2150-2180`) calls `SpawnSmith/SpawnWitch/SpawnHealer/SpawnBoy/SpawnPremium` (`:2176-2180`);
`RecreateItem`/`InvokeItems` re-instantiate vendor stock from `_iCreateInfo` CF_* bits
(`items.cpp:1990-2130`) — vendor items are **deterministic per seed**, which the port must match.
### V5.4 Cows — complete documentation (town cows + Hellfire cow quest)
#### (a) The three town cows
- Data rows 10-12 of V5.1: `TOWN_COW` at (58,16) SW, (56,14) NW, (59,20) N; `animWidth` 128,
  12 frames, delay 3, **no** `animPath` (shared sheet).
- Shared sprite sheet `CowSprites` (`towners.cpp:26`); `InitCows()` (`:141-175`) loads one frame
  per cow from `(*CowSprites)[entry.direction]` (`:154`) and occupies **4 tiles** by writing
  `-cowId` (negative id, the same convention the renderer uses for moving players) into the three
  extra tiles (`:158-171`) — comment: vanilla allowed standing next to the two top-right cows
  causing clipping, so all cows are treated as 4-tile sprites.
- Interaction `TalkToCow(player, cow)` (`towners.cpp:494-520`): ignored while a cow sound is
  already playing (`CowPlaying != None && effect_is_playing`, `:495-497`); `CowClicks++`;
  sound `Cow1`; **4th click** ⇒ `Cow2` (in spawn `CowClicks` resets to 0, so the sequence
  loops 1→2→1→2…); **8th+ click (non-spawn)** ⇒ reset to 4 and play a hero speech cycling
  `YepThatsACowAlright` → `ImNotThirsty` → `ImNoMilkmaid` (`HeroSpeech`, `CowMsg` wraps at 3,
  `:506-515`); `PlaySfxLoc(CowPlaying, cow.position)` (`:518`). Counter reset on any NPC talk:
  `TownerTalk` sets `CowClicks = CowMsg = 0` (`:205-209`); also reset at town init (`:207-208`).
- Voices come from the `hfvoice` archive (probed with `cowsut1`, `assets.cpp:570-586`); SfxIDs
  `Cow1`/`Cow2` = `sound_effect_enums.h:310-311`, `YepThatsACowAlright` = `:61`; both in V9.1.
- No gameplay effect: cows are decorative towners (`talk` only). `TEXT_COW1`/`TEXT_COW2` exist in
  `textdat.h:266-267`; grep confirms **no** secret cow level exists in this fork — the only
  “secret level” content is the Hellfire Hive/Nest + Crypt (V11) and the cow-quest below.
#### (b) Hellfire cow quest = Farmer's Orchard variant (`bCowQuest`)
- Option: `cowQuest` “Cow Quest”, flags `CantChangeInGame | OnlyHellfire`, default **false**,
  description: “Enable Jersey's quest. Lester the farmer is replaced by the Complete Nut.”
  (`Source/options.cpp:848`); carried on the wire as `sgGameInitInfo.bCowQuest`
  (`game_mode.hpp:27`), set from options at `multi.cpp:612`, demo-serialised
  (`demomode.cpp:126,164,193,228,647`). Chat gate check `control_chat_commands.cpp:159-161`.
- **NPC swap**: `IsTownerPresent` (`towners.cpp:735-750`): `TOWN_FARMER` requires
  `gbIsHellfire && bCowQuest == 0 && Q_FARMER != QUEST_HIVE_DONE`; `TOWN_COWFARM` requires
  `gbIsHellfire && bCowQuest != 0`; `TOWN_GIRL` requires `gbIsHellfire && bTheoQuest &&`
  `_pLvlVisited[17] && Q_GIRL != QUEST_DONE`. Cow farmer short name `nut` (`:722`).
- **Normal path (Lester, `Q_FARMER`)** — `TalkToFarmer` (`towners.cpp:523-576`):
  `NOTAVAIL/INIT`: with `IDI_RUNEBOMB` in inventory ⇒ `TEXT_FARMER2` + `QUEST_ACTIVE`,
  `_qvar1=1`, `_qmsg=TEXT_FARMER1`, `_qlog=true`; else if `!_pLvlVisited[9] && level < 15` ⇒
  taunts chosen by visited levels (`TEXT_FARMER8`, or `5` if L2 visited, `7` if L5, `9` if L7);
  else `TEXT_FARMER1` + ACTIVE + **`SpawnRuneBomb(farmer.pos + (1,0))`**. `ACTIVE` ⇒
  `TEXT_FARMER2` if holding the bomb else `TEXT_FARMER3`. `DONE` ⇒ `TEXT_FARMER4` +
  **`SpawnRewardItem(IDI_AURIC, …)`** (Auric Amulet) + state→`QUEST_HIVE_DONE`.
- **Cow path (Nut, `Q_JERSEY`)** — `TalkToCowFarmer` (`towners.cpp:586-650`): returning the
  **Grey Suit** ⇒ `TEXT_JERSEY7`; handing over the **Brown Suit** ⇒
  `SpawnUnique(UITEM_BOVINE, cowFarmer.pos + SouthEast, quest._qlevel)` + `TEXT_JERSEY8` +
  `QUEST_DONE` + `UpdateCowFarmerAnimAfterQuestComplete()` + `NetSendCmdQuest`; holding
  `IDI_RUNEBOMB` ⇒ `TEXT_JERSEY5` + ACTIVE/`qvar1=1`; first talk ⇒ `TEXT_JERSEY1`→`HIVE_TEASE1`;
  TEASE1 ⇒ `TEXT_JERSEY2`→TEASE2; TEASE2 ⇒ `TEXT_JERSEY3`→`HIVE_ACTIVE`; HIVE_ACTIVE while
  `!_pLvlVisited[9] && level < 15` ⇒ random taunt `PickRandomlyAmong({TEXT_JERSEY9,10,11,12})`;
  otherwise `TEXT_JERSEY4` + `QUEST_ACTIVE` + `_qvar1=1` + `SpawnRuneBomb(nut.pos + (1,0))`.
- **Suit spawns**: when `bCowQuest`, level generation puts `IDI_BROWNSUIT` at tile **(25,25)** on
  **level 20** and `IDI_GREYSUIT` at **(25,25)** on **level 19** (`items.cpp:2455-2460`, inside the
  `!setlevel` block next to `SpawnRock()`/`Q_ANVIL` placement).
- **Reward item**: Bovine Plate = `IDI_LGTFORGE` (V1.2 row 33, unique base `BOVINE`, HeavyArmor,
  durability 40, AC 0-0, req Str 50, `Quest` class) reached via `UITEM_BOVINE`
  (`_unique_items`, `items.h:47+`; `unique_itemdat.tsv` row with base `BOVINE`).
- **Rune bomb → Hive (the “secret” level)**: `IDI_RUNEBOMB` is a `Quest`/`Unequippable` item
  (V1.2 row 40). Holding it and clicking the town Hive tiles opens the Nest:
  `OpensHive(position) ⇒ x ∈ [79,82], y ∈ [61,64]` (`levels/town.cpp:249-254`), checked with the
  held item in `diablo.cpp:323` (`holdItem.IDidx == IDI_RUNEBOMB && OpensHive(cursPosition)`)
  which calls `OpenHive()` (`town.cpp:265-272`): `CMD_OPENHIVE` + `Q_FARMER._qactive = QUEST_DONE`.
  `TownOpenHive()` (`:283-300`) rewrites `dPiece[78..80][60..65]` to open the gate; `IsWarpOpen(DTYPE_NEST)`)
  drives it at level load (`:228-233`); town trigger tile **(80,62)** → level 17 (`trigs.cpp:136-141`);
  `AddOpenNest` missile spawns the explosion at those four tiles (`missiles.cpp:1238-1246`).
  Grave/Crypt analogues: `OpensGrave` x∈[35,38] y∈[20,24] (`:256-261`), `TownCryptList`
  (`trigs.cpp:34`), trigger (36,24)→level 21.
- **Na-Krul reacts**: with `bCowQuest` the Na-Krul greeting stream SFX becomes `NaKrul6` instead of
  `NaKrul4`/`NaKrul5` (`monster.cpp:931-936` and `:4282-4290`) — the joke payoff.
- **Q_JERSEY town hook**: `quests.cpp:618` — when `Q_JERSEY` completes while the player is on
  level 0, town state is refreshed (nut replaces the farmer animation: cel path
  `towners\farmer\cfrmrn2` → `mfrmrn2` on DONE, `towners.cpp:184-188`).
#### (c) Wirt's gambling — see V5.3 Wirt row (50-gold peek fee, single stock item, class-aware
re-roll constraints, HF −25%/Diablo +50% markup, `BoyItemLevel` half-level cooldown).
---
## V6. QUESTS — 24 questdat rows + the state machine
### V6.1 questdat.tsv rows 2-25 (24 rows, complete; loader `tables/questdat.cpp:80`)
Header: `qdlvl, qdmultlvl, qlvlt, bookOrder, qdrnd, qslvl, isSinglePlayerOnly, qdmsg, qlstr` — `qdlvl` SP trigger level,
`qdmultlvl` MP trigger level (`-1` ⇒ unavailable in MP-shared mode), `qlvlt` set-level dungeon
type, `bookOrder` quest-log sort key, `qdrnd` random-pool weight, `qslvl` `_setlevels` id,
`isSinglePlayerOnly`, `qdmsg` intro `_speech_id`, `qlstr` log title.
`quest_id` order is authoritative (`tables/questdat.hpp:74-99`); `MAXQUESTS 24` (`:14`).
| # | quest_id | qdlvl | qdmultlvl | qlvlt | bookOrder | qdrnd | qslvl | SPonly | qdmsg | qlstr |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Q_ROCK | 5 | -1 | - | 5 | 100 | SL_NONE | true | TEXT_INFRA5 | The Magic Rock |
| 2 | Q_MUSHROOM | 9 | -1 | - | 10 | 100 | SL_NONE | true | TEXT_MUSH8 | Black Mushroom |
| 3 | Q_GARBUD | 4 | -1 | - | 3 | 100 | SL_NONE | true | TEXT_GARBUD1 | Gharbad The Weak |
| 4 | Q_ZHAR | 8 | -1 | - | 9 | 100 | SL_NONE | true | TEXT_ZHAR1 | Zhar the Mad |
| 5 | Q_VEIL | 14 | -1 | - | 21 | 100 | SL_NONE | true | TEXT_VEIL9 | Lachdanan |
| 6 | Q_DIABLO | 15 | -1 | - | 23 | 100 | SL_NONE | false | TEXT_VILE3 | Diablo |
| 7 | Q_BUTCHER | 2 | 2 | - | 0 | 100 | SL_NONE | false | TEXT_BUTCH9 | The Butcher |
| 8 | Q_LTBANNER | 4 | -1 | - | 4 | 100 | SL_NONE | true | TEXT_BANNER2 | Ogden's Sign |
| 9 | Q_BLIND | 7 | -1 | - | 8 | 100 | SL_NONE | true | TEXT_BLINDING | Halls of the Blind |
| 10 | Q_BLOOD | 5 | -1 | - | 6 | 100 | SL_NONE | true | TEXT_BLOODY | Valor |
| 11 | Q_ANVIL | 10 | -1 | - | 11 | 100 | SL_NONE | true | TEXT_ANVIL5 | Anvil of Fury |
| 12 | Q_WARLORD | 13 | -1 | - | 20 | 100 | SL_NONE | true | TEXT_BLOODWAR | Warlord of Blood |
| 13 | Q_SKELKING | 3 | 3 | DTYPE_CATHEDRAL | 2 | 100 | SL_SKELKING | false | TEXT_KING2 | The Curse of King Leoric |
| 14 | Q_PWATER | 2 | -1 | DTYPE_CAVES | 1 | 100 | SL_POISONWATER | true | TEXT_POISON3 | Poisoned Water Supply |
| 15 | Q_SCHAMB | 6 | -1 | DTYPE_CATACOMBS | 7 | 100 | SL_BONECHAMB | true | TEXT_BONER | The Chamber of Bone |
| 16 | Q_BETRAYER | 15 | 15 | DTYPE_CATHEDRAL | 22 | 100 | SL_VILEBETRAYER | false | TEXT_VILE1 | Archbishop Lazarus |
| 17 | Q_GRAVE | 17 | 17 | - | 17 | 100 | SL_NONE | false | TEXT_GRAVE7 | Grave Matters |
| 18 | Q_FARMER | 9 | 9 | - | 12 | 100 | SL_NONE | false | TEXT_FARMER1 | Farmer's Orchard |
| 19 | Q_GIRL | 17 | -1 | - | 14 | 100 | SL_NONE | true | TEXT_GIRL2 | Little Girl |
| 20 | Q_TRADER | 19 | -1 | - | 16 | 100 | SL_NONE | true | TEXT_TRADER | Wandering Trader |
| 21 | Q_DEFILER | 17 | 17 | - | 15 | 100 | SL_NONE | false | TEXT_DEFILER1 | The Defiler |
| 22 | Q_NAKRUL | 21 | 21 | - | 19 | 100 | SL_NONE | false | TEXT_NAKRUL1 | Na-Krul |
| 23 | Q_CORNSTN | 21 | -1 | - | 18 | 100 | SL_NONE | true | TEXT_CORNSTN | Cornerstone of the World |
| 24 | Q_JERSEY | 9 | 9 | - | 13 | 100 | SL_NONE | false | TEXT_JERSEY4 | The Jersey's Jersey |
Diablo's 16 = rows 1-16 (Q_ROCK…Q_BETRAYER); Hellfire's 8 = Q_GRAVE, Q_FARMER, Q_GIRL, Q_TRADER,
Q_DEFILER, Q_NAKRUL, Q_CORNSTN, Q_JERSEY. `qdmultlvl == -1` on %d rows ⇒ single-player-only
quests (`isSinglePlayerOnly` true on 16).
15
### V6.2 Quest state machine (`Source/quests.cpp`, 623 lines)
- States — `quest_state` (`tables/questdat.hpp:62+`): `QUEST_NOTAVAIL`, `QUEST_INIT`,
  `QUEST_ACTIVE`, `QUEST_DONE`, and Hellfire hive states `QUEST_HIVE_TEASE1`, `QUEST_HIVE_TEASE2`,
  `QUEST_HIVE_ACTIVE`, `QUEST_HIVE_DONE`; per-quest progress lives in `_qvar1`/`_qvar2` using the
  `QS_*` constants (e.g. `QS_MUSHGIVEN`, `QS_BRAINSPAWNED`, `QS_TOMESPAWNED`).
- **`InitQuests()`** (`quests.cpp:60-110`): forces `TOWN_HEALER/Q_MUSHROOM = TEXT_NONE` and
  `TOWN_WITCH/Q_MUSHROOM = TEXT_MUSH9` (`:61-62`); `WaterDone = 0` (`:64`); per quest copies
  `_qlvltype/_qslvl/_qdmsg` from QuestData, `_qactive = QUEST_NOTAVAIL`, `position = (0,0)`;
  SP (`!UseMultiplayerQuests()`) ⇒ `_qlevel = _qdlvl`, `QUEST_INIT`; MP ⇒ `_qlevel = _qdmultlvl`
  and skipped when `isSinglePlayerOnly` (`:76-85`); then `!MP && randomizeQuests` ⇒
  **`InitialiseQuestPools(DungeonSeeds[15], Quests)`** (`:88-92` — quest order derives from the
  **level-15 seed**, function body `:114-135`); `gbIsSpawn` ⇒ wipe all to NOTAVAIL (`:94-99`);
  tail fixes: `Q_SKELKING._qvar2 = 2` and `Q_ROCK._qvar2 = 2` when unavailable,
  `Q_LTBANNER._qvar1 = 1` always, MP `Q_BETRAYER._qvar1 = 2`, MP mushroom pre-spawns the tome
  (`QS_TOMESPAWNED`) (`:101-110`).
- **`CheckQuests()`** (`:137-201`) per tick: MP betrayer spawns `OBJ_ALTBOY` at
  `SetPiece.pos + (4,6)` when `_qvar1 == 2` (`:142-147`); SP betrayer portal
  `AddMissile(quest.position, RedPortal)` + `_qvar2 = 1` (`:155-165`); after completion a return
  portal at **(35,32)** inside `SL_VILEBETRAYER` when `_qvar2 == 4` (`:167-174`);
  Poison Water auto-completes when `ActiveMonsterCount == 4` in the set level (`:176-187`);
  set-level entry: standing on `quest.position` with `_qslvl != 0` ⇒ `StartNewLvl(WM_DIABSETLVL,`
  `_qslvl)` (`:189-200`).
- **`CheckQuestKill(monster, sendmsg)`** (`:227-287`), returns early when `gbIsSpawn` (`:229`):
  `MT_SKING` ⇒ `Q_SKELKING = DONE` + `HeroSpeech::RestWellLeoricIllFindYourSon` (`:234-240`);
  `MT_CLEAVER` ⇒ `Q_BUTCHER = DONE` + `TheSpiritsOfTheDeadAreNowAvenged` (`:241-246`);
  `UniqueMonsterType::Garbud` ⇒ `Q_GARBUD DONE` + `ImNotImpressed` (`:247-250`);
  `Zhar` ⇒ `Q_ZHAR DONE` + `ImSorryDidIBreakYourConcentration` (`:251-255`);
  `Lazarus` ⇒ `Q_BETRAYER DONE` + `YourMadnessEndsHereBetrayer` **and `Q_DIABLO → ACTIVE`**
  (`:256-262`); `WarlordOfBlood` ⇒ `Q_WARLORD DONE` + `YourReignOfPainHasEnded` (`:263-266`).
  The remaining rewards (banner restore, mushroom→brain chain, holy water, anvil, lachdanan's
  Undead Crown, gharbad trade, chamber-of-bone) are driven by object/towner handlers plus the
  `_qvar*` state machine — every reward item/unique is one of the V1.5 rows or a V1.2 `Quest`-class
  row (`IDI_ROCK, IDI_FUNGALTM, IDI_BRAIN, IDI_BANNER, IDI_ANVIL, IDI_BLOODSTONE, IDI_INFRARING,
  IDI_OPTAMULET, IDI_TRING, IDI_HARCREST, IDI_STEELVEIL, IDI_ARMOFVAL, IDI_GRISWOLD,
  IDI_LAZSTAFF, IDI_NOTE1..4, IDI_AURIC, IDI_RUNEBOMB, IDI_THEODORE, IDI_BROWNSUIT, IDI_GREYSUIT`).
- **`GetMapReturnLevel()`** (`:289-325`) maps set levels back to their parent (`Q_SKELKING`,
  `Q_SCHAMB`, `Q_PWATER`, `Q_BETRAYER`), positions offset by direction (`:326-334`).
- **Resync on load** — `ResyncMPQuests()` (`:349-395`) re-activates SKELKING/BUTCHER/BETRAYER/CRYPT
  (`Q_GRAVE`)/DEFILER/NAKRUL from visited flags; `ResyncQuests()` (`:397-589`) rebuilds objects,
  portals and `_qvar` state per quest (the port must transcribe this switch in full).
- `SetMultiQuest` (`:591+`) is the network state applier (`MultiQuests{qstate,qlog,qvar1,qvar2,qmsg}`).
- Multiplayer differences: quest-pool randomisation disabled (`:88`), betrayer altar auto-spawn
  (`:142`), mushroom tome pre-spawn (`:108-110`), brain as an extra drop (`items.cpp:3446-3453`),
  `UseMultiplayerQuests()` (`questdat.hpp:131`), unique re-drops allowed (`items.cpp:3330`),
  `isSinglePlayerOnly` quests unavailable, `Q_JERSEY`/`Q_FARMER` hive states broadcast with
  `NetSendCmdQuest` on every transition (`towners.cpp:536,558,569,…`).
---
## V7. SPELLS — 36 spelldat rows + formulas
### V7.1 spelldat.tsv rows 2-37 (36 rows, complete; loader `tables/spelldat.cpp:232-256`)
Columns: `id | name | soundId | bookCost10 | staffCost10 | manaCost | flags | bookLevel |
staffLevel | minIntelligence | missiles | manaMultiplier | minMana | staffMin | staffMax`.
A **`Null` spell is prepended** by `AddNullSpell()` (`:20-35`: all costs 0, flags Fire,
sMissiles Null, sStaffMin/Max = 40/80), so `SpellsData.size() == 37` and `GetSpellData(id)` =
`SpellsData[id]` (`spelldat.h:269-272`). `flags` ∈ {Fire, Lightning, Magic} ∪ {Targeted,
AllowedInTown} (`ParseSpellDataFlag` `:48-55`); damage type is the magic-school bit.
`missiles` is parsed into `MissileID sMissiles[2]` with `Null` fill (`:250`) — only `Flash` uses
two (`FlashBottom,FlashTop`, row 4). `bookLevel/staffLevel = -1` ⇒ no book/staff exists
(`GetSpellBookLevel`/`GetSpellStaffLevel`, `spells.cpp:280-345`, with the `gbIsSpawn` skip lists
StoneCurse/Guardian/Golem/Elemental/BloodStar/BoneSpirit (+Apocalypse for staff)). `manaCost 255`
is the sentinel “cost = (max mana >> 6)”.
| # | id | name | sfx | book10 | staff10 | mana | flags | bookLvl | staffLvl | minInt | missiles | manaMul | minMana | staffMin | staffMax |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Firebolt | Firebolt | CastFire | 100 | 5 | 6 | Fire,Targeted | 1 | 1 | 15 | Firebolt | 1 | 3 | 40 | 80 |
| 2 | Healing | Healing | CastHealing | 100 | 5 | 5 | Magic,AllowedInTown | 1 | 1 | 17 | Healing | 3 | 1 | 20 | 40 |
| 3 | Lightning | Lightning | CastLightning | 300 | 15 | 10 | Lightning,Targeted | 4 | 3 | 20 | LightningControl | 1 | 6 | 20 | 60 |
| 4 | Flash | Flash | CastLightning | 750 | 50 | 30 | Lightning | 5 | 4 | 33 | FlashBottom,FlashTop | 2 | 16 | 20 | 40 |
| 5 | Identify | Identify | CastSkill | 0 | 10 | 13 | Magic,AllowedInTown | -1 | -1 | 23 | Identify | 2 | 1 | 8 | 12 |
| 6 | FireWall | Fire Wall | CastFire | 600 | 40 | 28 | Fire,Targeted | 3 | 2 | 27 | FireWallControl | 2 | 16 | 8 | 16 |
| 7 | TownPortal | Town Portal | CastSkill | 300 | 20 | 35 | Magic,Targeted | 3 | 3 | 20 | TownPortal | 3 | 18 | 8 | 12 |
| 8 | StoneCurse | Stone Curse | CastFire | 1200 | 80 | 60 | Magic,Targeted | 6 | 5 | 51 | StoneCurse | 3 | 40 | 8 | 16 |
| 9 | Infravision | Infravision | CastHealing | 0 | 60 | 40 | Magic | -1 | -1 | 36 | Infravision | 5 | 20 | 0 | 0 |
| 10 | Phasing | Phasing | CastFire | 350 | 20 | 12 | Magic | 7 | 6 | 39 | Phasing | 2 | 4 | 40 | 80 |
| 11 | ManaShield | Mana Shield | CastFire | 1600 | 120 | 33 | Magic | 6 | 5 | 25 | ManaShield | 0 | 33 | 4 | 10 |
| 12 | Fireball | Fireball | CastFire | 800 | 30 | 16 | Fire,Targeted | 8 | 7 | 48 | Fireball | 1 | 10 | 40 | 80 |
| 13 | Guardian | Guardian | CastFire | 1400 | 95 | 50 | Fire,Targeted | 9 | 8 | 61 | Guardian | 2 | 30 | 16 | 32 |
| 14 | ChainLightning | Chain Lightning | CastFire | 1100 | 75 | 30 | Lightning | 8 | 7 | 54 | ChainLightning | 1 | 18 | 20 | 60 |
| 15 | FlameWave | Flame Wave | CastFire | 1000 | 65 | 35 | Fire,Targeted | 9 | 8 | 54 | FlameWaveControl | 3 | 20 | 20 | 40 |
| 16 | DoomSerpents | Doom Serpents | CastFire | 0 | 0 | 0 | Lightning | -1 | -1 | 0 |  | 0 | 0 | 40 | 80 |
| 17 | BloodRitual | Blood Ritual | CastFire | 0 | 0 | 0 | Magic | -1 | -1 | 0 |  | 0 | 0 | 40 | 80 |
| 18 | Nova | Nova | CastLightning | 2100 | 130 | 60 | Magic | -1 | 10 | 87 | Nova | 3 | 35 | 16 | 32 |
| 19 | Invisibility | Invisibility | CastFire | 0 | 0 | 0 | Magic | -1 | -1 | 0 |  | 0 | 0 | 40 | 80 |
| 20 | Inferno | Inferno | CastFire | 200 | 10 | 11 | Fire,Targeted | 3 | 2 | 20 | InfernoControl | 1 | 6 | 20 | 40 |
| 21 | Golem | Golem | CastFire | 1800 | 110 | 100 | Fire,Targeted | 11 | 9 | 81 | Golem | 6 | 60 | 16 | 32 |
| 22 | Rage | Rage | CastHealing | 0 | 0 | 15 | Magic | -1 | -1 | 0 | Rage | 1 | 1 | 0 | 0 |
| 23 | Teleport | Teleport | CastSkill | 2000 | 125 | 35 | Magic,Targeted | 14 | 12 | 105 | Teleport | 3 | 15 | 16 | 32 |
| 24 | Apocalypse | Apocalypse | CastFire | 3000 | 200 | 150 | Fire | -1 | 15 | 149 | Apocalypse | 6 | 90 | 8 | 12 |
| 25 | Etherealize | Etherealize | CastFire | 2600 | 160 | 100 | Magic | -1 | -1 | 93 | Etherealize | 0 | 100 | 2 | 6 |
| 26 | ItemRepair | Item Repair | CastSkill | 0 | 0 | 0 | Magic,AllowedInTown | -1 | -1 | 255 | ItemRepair | 0 | 0 | 40 | 80 |
| 27 | StaffRecharge | Staff Recharge | CastSkill | 0 | 0 | 0 | Magic,AllowedInTown | -1 | -1 | 255 | StaffRecharge | 0 | 0 | 40 | 80 |
| 28 | TrapDisarm | Trap Disarm | CastSkill | 0 | 0 | 0 | Magic | -1 | -1 | 255 | TrapDisarm | 0 | 0 | 40 | 80 |
| 29 | Elemental | Elemental | CastFire | 1050 | 70 | 35 | Fire | 8 | 6 | 68 | Elemental | 2 | 20 | 20 | 60 |
| 30 | ChargedBolt | Charged Bolt | CastFire | 100 | 5 | 6 | Lightning,Targeted | 1 | 1 | 25 | ChargedBolt | 1 | 6 | 40 | 80 |
| 31 | HolyBolt | Holy Bolt | CastFire | 100 | 5 | 7 | Magic,Targeted | 1 | 1 | 20 | HolyBolt | 1 | 3 | 40 | 80 |
| 32 | Resurrect | Resurrect | CastHealing | 400 | 25 | 20 | Magic,AllowedInTown | -1 | 5 | 30 | Resurrect | 0 | 20 | 4 | 10 |
| 33 | Telekinesis | Telekinesis | CastFire | 250 | 20 | 15 | Magic | 2 | 2 | 33 | Telekinesis | 2 | 8 | 20 | 40 |
| 34 | HealOther | Heal Other | CastHealing | 100 | 5 | 5 | Magic,AllowedInTown | 1 | 1 | 17 | HealOther | 3 | 1 | 20 | 40 |
| 35 | BloodStar | Blood Star | CastFire | 2750 | 180 | 25 | Magic | 14 | 13 | 70 | BloodStar | 2 | 14 | 20 | 60 |
| 36 | BoneSpirit | Bone Spirit | CastFire | 1150 | 80 | 24 | Magic | 9 | 7 | 34 | BoneSpirit | 1 | 12 | 20 | 60 |
Row 1 = `Firebolt`; row 36 = `%s` (the last Diablo spell, `SpellID::LastDiablo = BoneSpirit`).
Hellfire spells (`Mana`…`RuneOfStone`, `SpellID` `spelldat.h:80-87`) are supplied by the `hf` mod
table: 15 appended rows + 2 modified (`mods/hf/txtdata/spells/spelldat.tsv`, 51 rows) — see V12.10.
### V7.2 Mana / damage formulas and item mechanics
- **Mana cost** — `GetManaAmount(player, sn)` (`Source/spells.cpp:104-140`):
  `sl = max(spellLevel-1, 0)`; `adj = sl * sManaAdj`; Firebolt halves `adj` (`:117-119`);
  Resurrect uses `adj = sl * (sManaCost/8)` (`:120-122`); Healing/HealOther use
  `ma = Healing.sManaCost + 2*level - adj` (`:124`); `sManaCost == 255 ⇒ ma = (_pMaxManaBase >> 6) - adj`
  (`:125-126`); else `ma = sManaCost - adj`; `ma = max(ma,0) << 6` then class multiplier
  `ma = ma * GetClassAttributes(cls).manaCost >> 6` (`:136-138`) — the V10.3 `manaCost` column.
- **Min-mana floor**: `sMinMana`/`minMana` clamps staff/potion-style rolls; `staffMin`/`staffMax`
  bound staff charge counts (`SpawnItem` staff path `items.cpp:1236-1270`, `GetStaffPower`).
- **Damage**: per-spell formulas in `Source/spells.cpp` (`CalculateSpellDamage`) driving the
  `AddMissile(..., midam, spllvl)` parameters; spell level comes from `pSplLvl[]` (stored 37 wide
  in saves, `pack.h:33-88`), learned via books/levels (`LevelUpSpell`).
- **Books** — V4.4 `GetBookSpell` walk; vendor rows `IDI_BOOK1..4` with `bookCost10`.
- **Scrolls** — `IMISC_SCROLL` rows (V1.2, 27 scroll rows: `scroll` count %d) consume the item on
  cast; unusable-scroll validation `IsReadiedSpellValid` (`spells.cpp:26-45`); `CURSOR_*` mapping
  for targeted spells (V12.1).
- **Staves** — `IMISC_STAFF` rows; `_iSpell` from `GetStaffSpell` (`items.cpp:1236+`), charges
  `_iCharges`/`_iMaxCharges` via `IPL_CHARGES` affix (`items.cpp:1048-1061`) or `staffMin/staffMax`;
  each cast decrements charges (player spell path), `SpellID::StaffRecharge` refills,
  `CURSOR_RECHARGE`; vanilla rule `!gbIsHellfire && FlipCoin(4)` ⇒ staff gets plain affixes instead
  of a spell (`items.cpp:1239`).
- **Skill vs spell** — `SpellType` (`spelldat.h:20`): `Skill` (hotkey 0, always available, e.g.
  Trap Disarm/Identify/Repair family), `Spell`, `Book`, `Scroll`, `Staff`, `Charge`, `Invalid`;
  readied-slot bookkeeping `_pRSplType`/`_pRSpell` (player.h:190-389).
- **Town rules** — `AllowedInTown` flag (rows: Healing, Identify, Town Portal, Item Repair, Staff
  Recharge, Resurrect, Heal Other, …) gates casting while `leveltype == DTYPE_TOWN`.
---
## V8. TEXT & MENUS — 338 textdat rows (+139 unused) and every menu tree
### V8.1 textdat.tsv rows 2-339 (338 entries, complete id list; loader `tables/textdat.cpp:106`)
Columns `txtstrid | txtstr | scrlltxt | sfxnr`; `_speech_id` enum `tables/textdat.h:20-448`.
`scrlltxt=true` ⇒ shown as full-screen scroll text (`interfac.cpp`/`DrawCutsceneText`);
`sfxnr` ⇒ SfxID to play with it. Consumers: `InitQTextMsg`/`DrawQText` (`Source/minitext.cpp`),
quest/towner dialog (V5.2), and `GetSpeechText`. **Text bodies stay in the TSV** — the Go port
embeds the file verbatim; only ids/flags/sfx are transcribed here (the bodies are translatable
strings owned by epic-10 i18n).
| file | rows | with sfxnr | scroll text | total chars |
|---|---|---|---|---|
| textdat.tsv | 338 | 338 | 324 | 74963 |
| textdat-unused.tsv | 139 | 139 | 139 | 30145 |
#### textdat rows 1–60 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 1 | TEXT_KING1 | true | Cain1 | 379 | Skeleton King quest |
| 2 | TEXT_KING2 | true | Ogden21 | 997 | Skeleton King quest |
| 3 | TEXT_KING3 | true | Ogden22 | 156 | Skeleton King quest |
| 4 | TEXT_KING4 | true | Ogden23 | 243 | Skeleton King quest |
| 5 | TEXT_KING5 | true | Pepin1 | 282 | Skeleton King quest |
| 6 | TEXT_KING6 | true | Gillian1 | 163 | Skeleton King quest |
| 7 | TEXT_KING7 | true | Griswold1 | 299 | Skeleton King quest |
| 8 | TEXT_KING8 | true | Farnham1 | 121 | Skeleton King quest |
| 9 | TEXT_KING9 | true | Adria1 | 246 | Skeleton King quest |
| 10 | TEXT_KING10 | true | Wirt1 | 228 | Skeleton King quest |
| 11 | TEXT_KING11 | false | LeoricGreeting | 98 | Skeleton King quest |
| 12 | TEXT_BANNER1 | true | Cain2 | 342 | Ogden's Sign |
| 13 | TEXT_BANNER2 | true | Ogden24 | 564 | Ogden's Sign |
| 14 | TEXT_BANNER3 | true | Ogden25 | 341 | Ogden's Sign |
| 15 | TEXT_BANNER4 | true | Pepin2 | 201 | Ogden's Sign |
| 16 | TEXT_BANNER5 | true | Gillian2 | 162 | Ogden's Sign |
| 17 | TEXT_BANNER6 | true | Griswold2 | 178 | Ogden's Sign |
| 18 | TEXT_BANNER7 | true | Farnham2 | 240 | Ogden's Sign |
| 19 | TEXT_BANNER8 | true | Adria2 | 134 | Ogden's Sign |
| 20 | TEXT_BANNER9 | true | Wirt2 | 176 | Ogden's Sign |
| 21 | TEXT_BANNER10 | true | Snotspill1 | 192 | Ogden's Sign |
| 22 | TEXT_BANNER11 | true | Snotspill2 | 56 | Ogden's Sign |
| 23 | TEXT_BANNER12 | true | Snotspill3 | 67 | Ogden's Sign |
| 24 | TEXT_VILE1 | true | Cain36 | 1239 | Archbishop/Lazarus quest |
| 25 | TEXT_VILE2 | true | Cain37 | 124 | Archbishop/Lazarus quest |
| 26 | TEXT_VILE3 | true | Cain38 | 622 | Archbishop/Lazarus quest |
| 27 | TEXT_VILE4 | true | Ogden1 | 284 | Archbishop/Lazarus quest |
| 28 | TEXT_VILE5 | true | Cain38 | 0 | Archbishop/Lazarus quest |
| 29 | TEXT_VILE6 | true | Cain38 | 0 | Archbishop/Lazarus quest |
| 30 | TEXT_VILE7 | true | Pepin3 | 295 | Archbishop/Lazarus quest |
| 31 | TEXT_VILE8 | true | Gillian3 | 222 | Archbishop/Lazarus quest |
| 32 | TEXT_VILE9 | true | Griswold3 | 275 | Archbishop/Lazarus quest |
| 33 | TEXT_VILE10 | true | Farnham3 | 204 | Archbishop/Lazarus quest |
| 34 | TEXT_VILE11 | true | Adria3 | 212 | Archbishop/Lazarus quest |
| 35 | TEXT_VILE12 | true | Wirt3 | 189 | Archbishop/Lazarus quest |
| 36 | TEXT_VILE13 | false | LazarusGreeting | 141 | Archbishop/Lazarus quest |
| 37 | TEXT_VILE14 | false | LazarusGreeting | 0 | Archbishop/Lazarus quest |
| 38 | TEXT_POISON1 | true | Cain4 | 313 | Poisoned Water Supply |
| 39 | TEXT_POISON2 | true | Ogden2 | 244 | Poisoned Water Supply |
| 40 | TEXT_POISON3 | true | Pepin20 | 351 | Poisoned Water Supply |
| 41 | TEXT_POISON4 | true | Pepin21 | 143 | Poisoned Water Supply |
| 42 | TEXT_POISON5 | true | Pepin22 | 280 | Poisoned Water Supply |
| 43 | TEXT_POISON6 | true | Gillian4 | 131 | Poisoned Water Supply |
| 44 | TEXT_POISON7 | true | Griswold4 | 193 | Poisoned Water Supply |
| 45 | TEXT_POISON8 | true | Farnham4 | 16 | Poisoned Water Supply |
| 46 | TEXT_POISON9 | true | Adria4 | 190 | Poisoned Water Supply |
| 47 | TEXT_POISON10 | true | Wirt4 | 140 | Poisoned Water Supply |
| 48 | TEXT_BONE1 | true | Cain7 | 413 | Chamber of Bone |
| 49 | TEXT_BONE2 | true | Ogden5 | 109 | Chamber of Bone |
| 50 | TEXT_BONE3 | true | Pepin5 | 86 | Chamber of Bone |
| 51 | TEXT_BONE4 | true | Gillian6 | 105 | Chamber of Bone |
| 52 | TEXT_BONE5 | true | Griswold7 | 156 | Chamber of Bone |
| 53 | TEXT_BONE6 | true | Farnham7 | 226 | Chamber of Bone |
| 54 | TEXT_BONE7 | true | Adria7 | 147 | Chamber of Bone |
| 55 | TEXT_BONE8 | true | Wirt7 | 202 | Chamber of Bone |
| 56 | TEXT_BUTCH1 | true | Cain10 | 350 | Butcher quest |
| 57 | TEXT_BUTCH2 | true | Ogden8 | 120 | Butcher quest |
| 58 | TEXT_BUTCH3 | true | Pepin8 | 413 | Butcher quest |
| 59 | TEXT_BUTCH4 | true | Gillian8 | 135 | Butcher quest |
| 60 | TEXT_BUTCH5 | true | Griswold10 | 389 | Butcher quest |
#### textdat rows 61–120 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 61 | TEXT_BUTCH6 | true | Farnham10 | 169 | Butcher quest |
| 62 | TEXT_BUTCH7 | true | Adria10 | 203 | Butcher quest |
| 63 | TEXT_BUTCH8 | true | Wirt10 | 255 | Butcher quest |
| 64 | TEXT_BUTCH9 | true | WoundedTownsman | 267 | Butcher quest |
| 65 | TEXT_BUTCH10 | true | ButcherGreeting | 0 | Butcher quest |
| 66 | TEXT_BLIND1 | true | Cain12 | 364 | Halls of the Blind |
| 67 | TEXT_BLIND2 | true | Ogden10 | 230 | Halls of the Blind |
| 68 | TEXT_BLIND3 | true | Pepin10 | 228 | Halls of the Blind |
| 69 | TEXT_BLIND4 | true | Gillian10 | 169 | Halls of the Blind |
| 70 | TEXT_BLIND5 | true | Griswold12 | 154 | Halls of the Blind |
| 71 | TEXT_BLIND6 | true | Farnham12 | 65 | Halls of the Blind |
| 72 | TEXT_BLIND7 | true | Adria12 | 165 | Halls of the Blind |
| 73 | TEXT_BLIND8 | true | Wirt11 | 189 | Halls of the Blind |
| 74 | TEXT_VEIL1 | true | Cain13 | 475 | Lachdanan quest |
| 75 | TEXT_VEIL2 | true | Ogden11 | 131 | Lachdanan quest |
| 76 | TEXT_VEIL3 | true | Pepin11 | 273 | Lachdanan quest |
| 77 | TEXT_VEIL4 | true | Gillian11 | 103 | Lachdanan quest |
| 78 | TEXT_VEIL5 | true | Griswold13 | 261 | Lachdanan quest |
| 79 | TEXT_VEIL6 | true | Farnham13 | 127 | Lachdanan quest |
| 80 | TEXT_VEIL7 | true | Adria13 | 157 | Lachdanan quest |
| 81 | TEXT_VEIL8 | true | Wirt12 | 267 | Lachdanan quest |
| 82 | TEXT_VEIL9 | true | Lachdanan1 | 655 | Lachdanan quest |
| 83 | TEXT_VEIL10 | true | Lachdanan2 | 98 | Lachdanan quest |
| 84 | TEXT_VEIL11 | true | Lachdanan3 | 327 | Lachdanan quest |
| 85 | TEXT_ANVIL1 | true | Cain14 | 531 | Anvil of Fury |
| 86 | TEXT_ANVIL2 | true | Ogden12 | 101 | Anvil of Fury |
| 87 | TEXT_ANVIL3 | true | Pepin12 | 229 | Anvil of Fury |
| 88 | TEXT_ANVIL4 | true | Gillian12 | 276 | Anvil of Fury |
| 89 | TEXT_ANVIL5 | true | Griswold21 | 887 | Anvil of Fury |
| 90 | TEXT_ANVIL6 | true | Griswold22 | 158 | Anvil of Fury |
| 91 | TEXT_ANVIL7 | true | Griswold23 | 211 | Anvil of Fury |
| 92 | TEXT_ANVIL8 | true | Farnham14 | 99 | Anvil of Fury |
| 93 | TEXT_ANVIL9 | true | Adria14 | 275 | Anvil of Fury |
| 94 | TEXT_ANVIL10 | true | Wirt13 | 127 | Anvil of Fury |
| 95 | TEXT_BLOOD1 | true | Cain15 | 870 | Valor/Blood King |
| 96 | TEXT_BLOOD2 | true | Ogden13 | 187 | Valor/Blood King |
| 97 | TEXT_BLOOD3 | true | Pepin13 | 156 | Valor/Blood King |
| 98 | TEXT_BLOOD4 | true | Gillian13 | 139 | Valor/Blood King |
| 99 | TEXT_BLOOD5 | true | Griswold14 | 282 | Valor/Blood King |
| 100 | TEXT_BLOOD6 | true | Farnham15 | 35 | Valor/Blood King |
| 101 | TEXT_BLOOD7 | true | Adria15 | 143 | Valor/Blood King |
| 102 | TEXT_BLOOD8 | true | Wirt14 | 187 | Valor/Blood King |
| 103 | TEXT_WARLRD1 | true | Cain18 | 571 | Warlord of Blood |
| 104 | TEXT_WARLRD2 | true | Ogden16 | 161 | Warlord of Blood |
| 105 | TEXT_WARLRD3 | true | Pepin16 | 98 | Warlord of Blood |
| 106 | TEXT_WARLRD4 | true | Gillian16 | 119 | Warlord of Blood |
| 107 | TEXT_WARLRD5 | true | Griswold17 | 122 | Warlord of Blood |
| 108 | TEXT_WARLRD6 | true | Farnham17 | 168 | Warlord of Blood |
| 109 | TEXT_WARLRD7 | true | Adria18 | 152 | Warlord of Blood |
| 110 | TEXT_WARLRD8 | true | Wirt17 | 145 | Warlord of Blood |
| 111 | TEXT_WARLRD9 | false | Warlord | 85 | Warlord of Blood |
| 112 | TEXT_INFRA1 | true | Cain20 | 349 | Magic Rock quest |
| 113 | TEXT_INFRA2 | true | Ogden18 | 222 | Magic Rock quest |
| 114 | TEXT_INFRA3 | true | Pepin18 | 155 | Magic Rock quest |
| 115 | TEXT_INFRA4 | true | Gillian18 | 256 | Magic Rock quest |
| 116 | TEXT_INFRA5 | true | Griswold24 | 457 | Magic Rock quest |
| 117 | TEXT_INFRA6 | true | Griswold25 | 120 | Magic Rock quest |
| 118 | TEXT_INFRA7 | true | Griswold26 | 195 | Magic Rock quest |
| 119 | TEXT_INFRA8 | true | Farnham19 | 169 | Magic Rock quest |
| 120 | TEXT_INFRA9 | true | Adria20 | 175 | Magic Rock quest |
#### textdat rows 121–180 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 121 | TEXT_INFRA10 | true | Wirt18 | 167 | Magic Rock quest |
| 122 | TEXT_MUSH1 | true | Cain21 | 224 | Black Mushroom quest |
| 123 | TEXT_MUSH2 | true | Ogden19 | 233 | Black Mushroom quest |
| 124 | TEXT_MUSH3 | true | Pepin26 | 353 | Black Mushroom quest |
| 125 | TEXT_MUSH4 | true | Pepin27 | 208 | Black Mushroom quest |
| 126 | TEXT_MUSH5 | true | Gillian19 | 85 | Black Mushroom quest |
| 127 | TEXT_MUSH6 | true | Griswold19 | 215 | Black Mushroom quest |
| 128 | TEXT_MUSH7 | true | Farnham20 | 127 | Black Mushroom quest |
| 129 | TEXT_MUSH8 | true | Adria22 | 201 | Black Mushroom quest |
| 130 | TEXT_MUSH9 | true | Adria23 | 137 | Black Mushroom quest |
| 131 | TEXT_MUSH10 | true | Adria24 | 347 | Black Mushroom quest |
| 132 | TEXT_MUSH11 | true | Adria25 | 264 | Black Mushroom quest |
| 133 | TEXT_MUSH12 | true | Adria26 | 122 | Black Mushroom quest |
| 134 | TEXT_MUSH13 | true | Wirt19 | 96 | Black Mushroom quest |
| 135 | TEXT_DOOM1 | true | Cain22 | 981 | misc lore/dialog/task text |
| 136 | TEXT_DOOM2 | true | Cain23 | 115 | misc lore/dialog/task text |
| 137 | TEXT_DOOM3 | true | Cain24 | 298 | misc lore/dialog/task text |
| 138 | TEXT_DOOM4 | true | Ogden20 | 215 | misc lore/dialog/task text |
| 139 | TEXT_DOOM5 | true | Pepin19 | 218 | misc lore/dialog/task text |
| 140 | TEXT_DOOM6 | true | Gillian20 | 205 | misc lore/dialog/task text |
| 141 | TEXT_DOOM7 | true | Griswold20 | 170 | misc lore/dialog/task text |
| 142 | TEXT_DOOM8 | true | Farnham21 | 116 | misc lore/dialog/task text |
| 143 | TEXT_DOOM9 | true | Adria21 | 217 | misc lore/dialog/task text |
| 144 | TEXT_DOOM10 | true | Wirt20 | 193 | misc lore/dialog/task text |
| 145 | TEXT_GARBUD1 | true | Gharbad1 | 72 | Gharbad quest |
| 146 | TEXT_GARBUD2 | true | Gharbad2 | 118 | Gharbad quest |
| 147 | TEXT_GARBUD3 | true | Gharbad3 | 103 | Gharbad quest |
| 148 | TEXT_GARBUD4 | true | Gharbad4 | 58 | Gharbad quest |
| 149 | TEXT_ZHAR1 | true | Zhar1 | 140 | Zhar quest |
| 150 | TEXT_ZHAR2 | true | Zhar2 | 51 | Zhar quest |
| 151 | TEXT_STORY1 | false | Cain25 | 43 | Cain gossip |
| 152 | TEXT_STORY2 | true | Cain26 | 175 | Cain gossip |
| 153 | TEXT_STORY3 | true | Cain27 | 239 | Cain gossip |
| 154 | TEXT_STORY4 | true | Cain28 | 348 | Cain gossip |
| 155 | TEXT_STORY5 | true | Cain29 | 332 | Cain gossip |
| 156 | TEXT_STORY6 | true | Cain30 | 359 | Cain gossip |
| 157 | TEXT_STORY7 | true | Cain31 | 327 | Cain gossip |
| 158 | TEXT_STORY9 | true | Cain33 | 429 | Cain gossip |
| 159 | TEXT_STORY10 | true | Cain34 | 234 | Cain gossip |
| 160 | TEXT_STORY11 | true | Cain35 | 316 | Cain gossip |
| 161 | TEXT_OGDEN1 | false | Ogden36 | 64 | Ogden gossip |
| 162 | TEXT_OGDEN2 | true | Ogden37 | 268 | Ogden gossip |
| 163 | TEXT_OGDEN3 | true | Ogden38 | 149 | Ogden gossip |
| 164 | TEXT_OGDEN4 | true | Ogden39 | 152 | Ogden gossip |
| 165 | TEXT_OGDEN5 | true | Ogden40 | 217 | Ogden gossip |
| 166 | TEXT_OGDEN6 | true | Ogden41 | 113 | Ogden gossip |
| 167 | TEXT_OGDEN8 | true | Ogden43 | 270 | Ogden gossip |
| 168 | TEXT_OGDEN9 | true | Ogden44 | 196 | Ogden gossip |
| 169 | TEXT_OGDEN10 | true | Ogden45 | 259 | Ogden gossip |
| 170 | TEXT_PEPIN1 | false | Pepin37 | 25 | Pepin gossip |
| 171 | TEXT_PEPIN2 | true | Pepin38 | 233 | Pepin gossip |
| 172 | TEXT_PEPIN3 | true | Pepin39 | 244 | Pepin gossip |
| 173 | TEXT_PEPIN4 | true | Pepin40 | 205 | Pepin gossip |
| 174 | TEXT_PEPIN5 | true | Pepin41 | 194 | Pepin gossip |
| 175 | TEXT_PEPIN6 | true | Pepin42 | 147 | Pepin gossip |
| 176 | TEXT_PEPIN7 | true | Pepin43 | 316 | Pepin gossip |
| 177 | TEXT_PEPIN9 | true | Pepin45 | 240 | Pepin gossip |
| 178 | TEXT_PEPIN10 | true | Pepin46 | 361 | Pepin gossip |
| 179 | TEXT_PEPIN11 | true | Pepin47 | 179 | Pepin gossip |
| 180 | TEXT_GILLIAN1 | false | Gillian31 | 30 | Gillian gossip |
#### textdat rows 181–240 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 181 | TEXT_GILLIAN2 | true | Gillian32 | 117 | Gillian gossip |
| 182 | TEXT_GILLIAN3 | true | Gillian33 | 229 | Gillian gossip |
| 183 | TEXT_GILLIAN4 | true | Gillian34 | 268 | Gillian gossip |
| 184 | TEXT_GILLIAN5 | true | Gillian35 | 152 | Gillian gossip |
| 185 | TEXT_GILLIAN6 | true | Gillian36 | 250 | Gillian gossip |
| 186 | TEXT_GILLIAN7 | true | Gillian37 | 262 | Gillian gossip |
| 187 | TEXT_GILLIAN9 | true | Gillian39 | 383 | Gillian gossip |
| 188 | TEXT_GILLIAN10 | true | Gillian40 | 241 | Gillian gossip |
| 189 | TEXT_GRISWOLD1 | false | Griswold44 | 27 | Griswold gossip |
| 190 | TEXT_GRISWOLD2 | true | Griswold45 | 239 | Griswold gossip |
| 191 | TEXT_GRISWOLD3 | true | Griswold46 | 229 | Griswold gossip |
| 192 | TEXT_GRISWOLD4 | true | Griswold47 | 256 | Griswold gossip |
| 193 | TEXT_GRISWOLD5 | true | Griswold48 | 184 | Griswold gossip |
| 194 | TEXT_GRISWOLD6 | true | Griswold49 | 294 | Griswold gossip |
| 195 | TEXT_GRISWOLD7 | true | Griswold50 | 151 | Griswold gossip |
| 196 | TEXT_GRISWOLD8 | true | Griswold51 | 202 | Griswold gossip |
| 197 | TEXT_GRISWOLD9 | true | Griswold52 | 301 | Griswold gossip |
| 198 | TEXT_GRISWOLD10 | true | Griswold53 | 175 | Griswold gossip |
| 199 | TEXT_GRISWOLD12 | true | Griswold55 | 335 | Griswold gossip |
| 200 | TEXT_GRISWOLD13 | true | Griswold56 | 445 | Griswold gossip |
| 201 | TEXT_FARNHAM1 | false | Farnham27 | 29 | misc lore/dialog/task text |
| 202 | TEXT_FARNHAM2 | true | Farnham28 | 74 | misc lore/dialog/task text |
| 203 | TEXT_FARNHAM3 | true | Farnham29 | 225 | misc lore/dialog/task text |
| 204 | TEXT_FARNHAM4 | true | Farnham30 | 156 | misc lore/dialog/task text |
| 205 | TEXT_FARNHAM5 | true | Farnham31 | 138 | misc lore/dialog/task text |
| 206 | TEXT_FARNHAM6 | true | Farnham32 | 230 | misc lore/dialog/task text |
| 207 | TEXT_FARNHAM8 | true | Farnham34 | 220 | misc lore/dialog/task text |
| 208 | TEXT_FARNHAM9 | true | Farnham35 | 217 | misc lore/dialog/task text |
| 209 | TEXT_FARNHAM10 | true | Farnham23 | 175 | misc lore/dialog/task text |
| 210 | TEXT_FARNHAM11 | true | Farnham24 | 181 | misc lore/dialog/task text |
| 211 | TEXT_FARNHAM12 | true | Farnham25 | 252 | misc lore/dialog/task text |
| 212 | TEXT_FARNHAM13 | true | Farnham26 | 225 | misc lore/dialog/task text |
| 213 | TEXT_ADRIA1 | false | Adria38 | 38 | Adria gossip |
| 214 | TEXT_ADRIA2 | true | Adria39 | 218 | Adria gossip |
| 215 | TEXT_ADRIA3 | true | Adria40 | 407 | Adria gossip |
| 216 | TEXT_ADRIA4 | true | Adria41 | 347 | Adria gossip |
| 217 | TEXT_ADRIA5 | true | Adria42 | 199 | Adria gossip |
| 218 | TEXT_ADRIA6 | true | Adria43 | 190 | Adria gossip |
| 219 | TEXT_ADRIA7 | true | Adria44 | 235 | Adria gossip |
| 220 | TEXT_ADRIA8 | true | Adria45 | 263 | Adria gossip |
| 221 | TEXT_ADRIA9 | true | Adria46 | 276 | Adria gossip |
| 222 | TEXT_ADRIA10 | true | Adria47 | 318 | Adria gossip |
| 223 | TEXT_ADRIA12 | true | Adria49 | 460 | Adria gossip |
| 224 | TEXT_ADRIA13 | true | Adria50 | 512 | Adria gossip |
| 225 | TEXT_WIRT1 | false | Wirt32 | 21 | Wirt gossip |
| 226 | TEXT_WIRT2 | true | Wirt33 | 214 | Wirt gossip |
| 227 | TEXT_WIRT3 | true | Wirt34 | 95 | Wirt gossip |
| 228 | TEXT_WIRT4 | true | Wirt35 | 247 | Wirt gossip |
| 229 | TEXT_WIRT5 | true | Wirt36 | 270 | Wirt gossip |
| 230 | TEXT_WIRT6 | true | Wirt37 | 281 | Wirt gossip |
| 231 | TEXT_WIRT7 | true | Wirt38 | 217 | Wirt gossip |
| 232 | TEXT_WIRT8 | true | Wirt39 | 360 | Wirt gossip |
| 233 | TEXT_WIRT9 | true | Wirt40 | 163 | Wirt gossip |
| 234 | TEXT_WIRT11 | true | Wirt42 | 301 | Wirt gossip |
| 235 | TEXT_WIRT12 | true | Wirt43 | 270 | Wirt gossip |
| 236 | TEXT_BONER | true | Warrior1 | 192 | Chamber of Bone |
| 237 | TEXT_BLOODY | true | Warrior10 | 120 | Valor/Blood King |
| 238 | TEXT_BLINDING | true | Warrior11 | 259 | misc lore/dialog/task text |
| 239 | TEXT_BLOODWAR | true | Warrior12 | 230 | misc lore/dialog/task text |
| 240 | TEXT_MBONER | true | Warrior1 | 192 | misc lore/dialog/task text |
#### textdat rows 241–300 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 241 | TEXT_MBLOODY | true | Warrior10 | 120 | misc lore/dialog/task text |
| 242 | TEXT_MBLINDING | true | Warrior11 | 259 | misc lore/dialog/task text |
| 243 | TEXT_MBLOODWAR | true | Warrior12 | 230 | misc lore/dialog/task text |
| 244 | TEXT_RBONER | true | Warrior1 | 192 | misc lore/dialog/task text |
| 245 | TEXT_RBLOODY | true | Warrior10 | 120 | misc lore/dialog/task text |
| 246 | TEXT_RBLINDING | true | Warrior11 | 259 | misc lore/dialog/task text |
| 247 | TEXT_RBLOODWAR | true | Warrior12 | 230 | misc lore/dialog/task text |
| 248 | TEXT_COW1 | false | Cow1 | 0 | cow text |
| 249 | TEXT_COW2 | false | Cow2 | 0 | cow text |
| 250 | TEXT_BOOK11 | true | Narrator1 | 497 | lore book |
| 251 | TEXT_BOOK12 | true | Narrator2 | 462 | lore book |
| 252 | TEXT_BOOK13 | true | Narrator3 | 1099 | lore book |
| 253 | TEXT_BOOK21 | true | Narrator4 | 497 | lore book |
| 254 | TEXT_BOOK22 | true | Narrator5 | 370 | lore book |
| 255 | TEXT_BOOK23 | true | Narrator6 | 822 | lore book |
| 256 | TEXT_BOOK31 | true | Narrator7 | 435 | lore book |
| 257 | TEXT_BOOK32 | true | Narrator8 | 418 | lore book |
| 258 | TEXT_BOOK33 | true | Narrator9 | 624 | lore book |
| 259 | TEXT_INTRO | true | Ogden0 | 623 | misc lore/dialog/task text |
| 260 | TEXT_HBONER | true | Warrior1 | 192 | misc lore/dialog/task text |
| 261 | TEXT_HBLOODY | true | Warrior10 | 120 | misc lore/dialog/task text |
| 262 | TEXT_HBLINDING | true | Warrior11 | 259 | misc lore/dialog/task text |
| 263 | TEXT_HBLOODWAR | true | Warrior12 | 230 | misc lore/dialog/task text |
| 264 | TEXT_BBONER | true | Warrior1 | 192 | misc lore/dialog/task text |
| 265 | TEXT_BBLOODY | true | Warrior10 | 120 | misc lore/dialog/task text |
| 266 | TEXT_BBLINDING | true | Warrior11 | 259 | misc lore/dialog/task text |
| 267 | TEXT_BBLOODWAR | true | Warrior12 | 230 | misc lore/dialog/task text |
| 268 | TEXT_GRAVE1 | true | Adria19 | 190 | Ghoul Dance (HF crypt) |
| 269 | TEXT_GRAVE2 | true | Griswold18 | 259 | Ghoul Dance (HF crypt) |
| 270 | TEXT_GRAVE3 | true | Pepin17 | 130 | Ghoul Dance (HF crypt) |
| 271 | TEXT_GRAVE4 | true | Adria9 | 208 | Ghoul Dance (HF crypt) |
| 272 | TEXT_GRAVE5 | true | Ogden17 | 169 | Ghoul Dance (HF crypt) |
| 273 | TEXT_GRAVE6 | true | Cain19 | 264 | Ghoul Dance (HF crypt) |
| 274 | TEXT_GRAVE7 | true | Farnham21 | 122 | Ghoul Dance (HF crypt) |
| 275 | TEXT_GRAVE8 | true | Gillian27 | 447 | Ghoul Dance (HF crypt) |
| 276 | TEXT_GRAVE9 | true | Wirt7 | 215 | Ghoul Dance (HF crypt) |
| 277 | TEXT_GRAVE10 | true | Adria9 | 208 | Ghoul Dance (HF crypt) |
| 278 | TEXT_FARMER1 | true | Farmer1 | 494 | Farmer's Orchard (HF hive) |
| 279 | TEXT_FARMER2 | true | Farmer2 | 124 | Farmer's Orchard (HF hive) |
| 280 | TEXT_FARMER3 | true | Farmer3 | 228 | Farmer's Orchard (HF hive) |
| 281 | TEXT_FARMER4 | true | Farmer4 | 259 | Farmer's Orchard (HF hive) |
| 282 | TEXT_FARMER5 | true | Farmer5 | 235 | Farmer's Orchard (HF hive) |
| 283 | TEXT_GIRL1 | true | Celia1 | 31 | Little Girl/Theodore (HF) |
| 284 | TEXT_GIRL2 | true | Celia2 | 290 | Little Girl/Theodore (HF) |
| 285 | TEXT_GIRL3 | true | Celia3 | 109 | Little Girl/Theodore (HF) |
| 286 | TEXT_GIRL4 | true | Celia4 | 199 | Little Girl/Theodore (HF) |
| 287 | TEXT_DEFILER1 | true | Defiler6 | 144 | The Defiler (HF) |
| 288 | TEXT_DEFILER2 | true | Defiler2 | 109 | The Defiler (HF) |
| 289 | TEXT_DEFILER3 | true | Defiler7 | 222 | The Defiler (HF) |
| 290 | TEXT_DEFILER4 | true | Defiler4 | 96 | The Defiler (HF) |
| 291 | TEXT_DEFILER5 | true | Defiler8 | 0 | The Defiler (HF) |
| 292 | TEXT_NAKRUL1 | true | NaKrul1 | 0 | Na-Krul (HF) |
| 293 | TEXT_NAKRUL2 | true | NaKrul2 | 0 | Na-Krul (HF) |
| 294 | TEXT_NAKRUL3 | true | NaKrul3 | 0 | Na-Krul (HF) |
| 295 | TEXT_NAKRUL4 | true | NaKrul4 | 0 | Na-Krul (HF) |
| 296 | TEXT_NAKRUL5 | true | NaKrul5 | 0 | Na-Krul (HF) |
| 297 | TEXT_CORNSTN | true | NarratorHF3 | 762 | misc lore/dialog/task text |
| 298 | TEXT_JERSEY1 | true | CompleteNut1 | 4 | Cow quest (HF) |
| 299 | TEXT_JERSEY2 | true | CompleteNut2 | 12 | Cow quest (HF) |
| 300 | TEXT_JERSEY3 | true | CompleteNut3 | 24 | Cow quest (HF) |
#### textdat rows 301–338 of 338
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 301 | TEXT_JERSEY4 | true | CompleteNut4 | 708 | Cow quest (HF) |
| 302 | TEXT_JERSEY5 | true | CompleteNut5 | 106 | Cow quest (HF) |
| 303 | TEXT_JERSEY6 | true | CompleteNut6 | 98 | Cow quest (HF) |
| 304 | TEXT_JERSEY7 | true | CompleteNut7 | 158 | Cow quest (HF) |
| 305 | TEXT_JERSEY8 | true | CompleteNut8 | 349 | Cow quest (HF) |
| 306 | TEXT_JERSEY9 | true | CompleteNut9 | 100 | Cow quest (HF) |
| 307 | TEXT_TRADER | true | None | 0 | misc lore/dialog/task text |
| 308 | TEXT_FARMER6 | true | Farmer2a | 142 | Farmer's Orchard (HF hive) |
| 309 | TEXT_FARMER7 | true | Farmer6 | 109 | Farmer's Orchard (HF hive) |
| 310 | TEXT_FARMER8 | true | Farmer7 | 152 | Farmer's Orchard (HF hive) |
| 311 | TEXT_FARMER9 | true | Farmer8 | 146 | Farmer's Orchard (HF hive) |
| 312 | TEXT_FARMER10 | true | Farmer9 | 180 | Farmer's Orchard (HF hive) |
| 313 | TEXT_JERSEY10 | true | CompleteNut10 | 77 | Cow quest (HF) |
| 314 | TEXT_JERSEY11 | true | CompleteNut11 | 142 | Cow quest (HF) |
| 315 | TEXT_JERSEY12 | true | CompleteNut12 | 191 | Cow quest (HF) |
| 316 | TEXT_JERSEY13 | true | CompleteNut4a | 773 | Cow quest (HF) |
| 317 | TEXT_SKLJRN | true | None | 0 | misc lore/dialog/task text |
| 318 | TEXT_BOOK4 | true | NarratorHF6 | 181 | lore book |
| 319 | TEXT_BOOK5 | true | NarratorHF7 | 270 | lore book |
| 320 | TEXT_BOOK6 | true | NarratorHF8 | 289 | lore book |
| 321 | TEXT_BOOK7 | true | NarratorHF5 | 260 | lore book |
| 322 | TEXT_BOOK8 | true | NarratorHF9 | 240 | lore book |
| 323 | TEXT_BOOK9 | true | NarratorHF4 | 626 | lore book |
| 324 | TEXT_BOOKA | true | Warrior54 | 19 | lore book |
| 325 | TEXT_BOOKB | true | Warrior55 | 17 | lore book |
| 326 | TEXT_BOOKC | true | Warrior56 | 27 | lore book |
| 327 | TEXT_OBOOKA | true | Warrior54 | 19 | misc lore/dialog/task text |
| 328 | TEXT_OBOOKB | true | Warrior55 | 17 | misc lore/dialog/task text |
| 329 | TEXT_OBOOKC | true | Warrior56 | 27 | misc lore/dialog/task text |
| 330 | TEXT_MBOOKA | true | Warrior54 | 19 | misc lore/dialog/task text |
| 331 | TEXT_MBOOKB | true | Warrior55 | 17 | misc lore/dialog/task text |
| 332 | TEXT_MBOOKC | true | Warrior56 | 27 | misc lore/dialog/task text |
| 333 | TEXT_RBOOKA | true | Warrior54 | 19 | lore book |
| 334 | TEXT_RBOOKB | true | Warrior55 | 17 | lore book |
| 335 | TEXT_RBOOKC | true | Warrior56 | 27 | lore book |
| 336 | TEXT_BBOOKA | true | Warrior54 | 19 | misc lore/dialog/task text |
| 337 | TEXT_BBOOKB | true | Warrior55 | 17 | misc lore/dialog/task text |
| 338 | TEXT_BBOOKC | true | Warrior56 | 27 | misc lore/dialog/task text |
Row 1 = `TEXT_KING1` (King Leoric lore, scroll text); row 338 = `TEXT_BBOOKC`.
### V8.2 textdat-unused.tsv rows 2-140 (139 entries, complete ids)
#### unused rows 1–60 of 139
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 1 | TEXT_DEADGUY | true | WoundedTownsmanOld | 167 | misc lore/dialog/task text |
| 2 | TEXT_FARNHAM14 | true | Farnham5 | 160 | misc lore/dialog/task text |
| 3 | TEXT_FARNHAM15 | true | Farnham6 | 7 | misc lore/dialog/task text |
| 4 | TEXT_FARNHAM16 | true | Farnham8 | 66 | misc lore/dialog/task text |
| 5 | TEXT_FARNHAM17 | true | Farnham9 | 65 | misc lore/dialog/task text |
| 6 | TEXT_FARNHAM18 | true | Farnham11 | 161 | misc lore/dialog/task text |
| 7 | TEXT_FARNHAM19 | true | Farnham16 | 124 | misc lore/dialog/task text |
| 8 | TEXT_FARNHAM20 | true | Farnham18 | 41 | misc lore/dialog/task text |
| 9 | TEXT_FARNHAM21 | true | Farnham22 | 202 | misc lore/dialog/task text |
| 10 | TEXT_FARNHAM22 | true | Farnham33 | 203 | misc lore/dialog/task text |
| 11 | TEXT_GILLIAN11 | true | Gillian5 | 184 | Gillian gossip |
| 12 | TEXT_GILLIAN12 | true | Gillian7 | 155 | Gillian gossip |
| 13 | TEXT_GILLIAN13 | true | Gillian9 | 81 | Gillian gossip |
| 14 | TEXT_GILLIAN14 | true | Gillian14 | 70 | Gillian gossip |
| 15 | TEXT_GILLIAN15 | true | Gillian15 | 399 | Gillian gossip |
| 16 | TEXT_GILLIAN16 | true | Gillian17 | 241 | Gillian gossip |
| 17 | TEXT_GILLIAN17 | true | Gillian21 | 470 | Gillian gossip |
| 18 | TEXT_GILLIAN18 | true | Gillian22 | 163 | Gillian gossip |
| 19 | TEXT_GILLIAN19 | true | Gillian23 | 243 | Gillian gossip |
| 20 | TEXT_GILLIAN20 | true | Gillian24 | 452 | Gillian gossip |
| 21 | TEXT_GILLIAN21 | true | Gillian25 | 135 | Gillian gossip |
| 22 | TEXT_GILLIAN22 | true | Gillian26 | 257 | Gillian gossip |
| 23 | TEXT_GILLIAN23 | true | Gillian28 | 241 | Gillian gossip |
| 24 | TEXT_GILLIAN24 | true | Gillian29 | 319 | Gillian gossip |
| 25 | TEXT_GILLIAN25 | true | Gillian30 | 117 | Gillian gossip |
| 26 | TEXT_GILLIAN26 | true | Gillian38 | 194 | Gillian gossip |
| 27 | TEXT_PEPIN12 | true | Pepin4 | 208 | Pepin gossip |
| 28 | TEXT_PEPIN13 | true | Pepin6 | 155 | Pepin gossip |
| 29 | TEXT_PEPIN14 | true | Pepin7 | 175 | Pepin gossip |
| 30 | TEXT_PEPIN15 | true | Pepin9 | 185 | Pepin gossip |
| 31 | TEXT_PEPIN16 | true | Pepin14 | 308 | Pepin gossip |
| 32 | TEXT_PEPIN17 | true | Pepin15 | 370 | Pepin gossip |
| 33 | TEXT_PEPIN18 | true | Pepin23 | 524 | Pepin gossip |
| 34 | TEXT_PEPIN19 | true | Pepin24 | 237 | Pepin gossip |
| 35 | TEXT_PEPIN20 | true | Pepin25 | 110 | Pepin gossip |
| 36 | TEXT_PEPIN21 | true | Pepin28 | 265 | Pepin gossip |
| 37 | TEXT_PEPIN22 | true | Pepin29 | 144 | Pepin gossip |
| 38 | TEXT_PEPIN23 | true | Pepin30 | 120 | Pepin gossip |
| 39 | TEXT_PEPIN24 | true | Pepin31 | 273 | Pepin gossip |
| 40 | TEXT_PEPIN25 | true | Pepin32 | 129 | Pepin gossip |
| 41 | TEXT_PEPIN26 | true | Pepin33 | 172 | Pepin gossip |
| 42 | TEXT_PEPIN27 | true | Pepin34 | 246 | Pepin gossip |
| 43 | TEXT_PEPIN28 | true | Pepin35 | 148 | Pepin gossip |
| 44 | TEXT_PEPIN29 | true | Pepin36 | 128 | Pepin gossip |
| 45 | TEXT_PEPIN30 | true | Pepin44 | 265 | Pepin gossip |
| 46 | TEXT_GRISWOLD14 | true | Griswold5 | 156 | Griswold gossip |
| 47 | TEXT_GRISWOLD15 | true | Griswold6 | 183 | Griswold gossip |
| 48 | TEXT_GRISWOLD16 | true | Griswold8 | 186 | Griswold gossip |
| 49 | TEXT_GRISWOLD17 | true | Griswold9 | 314 | Griswold gossip |
| 50 | TEXT_GRISWOLD18 | true | Griswold11 | 306 | Griswold gossip |
| 51 | TEXT_GRISWOLD19 | true | Griswold15 | 258 | Griswold gossip |
| 52 | TEXT_GRISWOLD20 | true | Griswold16 | 337 | Griswold gossip |
| 53 | TEXT_GRISWOLD21 | true | Griswold27 | 183 | Griswold gossip |
| 54 | TEXT_GRISWOLD22 | true | Griswold28 | 89 | Griswold gossip |
| 55 | TEXT_GRISWOLD23 | true | Griswold29 | 77 | Griswold gossip |
| 56 | TEXT_GRISWOLD24 | true | Griswold30 | 167 | Griswold gossip |
| 57 | TEXT_GRISWOLD25 | true | Griswold31 | 45 | Griswold gossip |
| 58 | TEXT_GRISWOLD26 | true | Griswold32 | 92 | Griswold gossip |
| 59 | TEXT_GRISWOLD27 | true | Griswold33 | 169 | Griswold gossip |
| 60 | TEXT_GRISWOLD28 | true | Griswold34 | 83 | Griswold gossip |
#### unused rows 61–120 of 139
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 61 | TEXT_GRISWOLD29 | true | Griswold35 | 48 | Griswold gossip |
| 62 | TEXT_GRISWOLD30 | true | Griswold36 | 177 | Griswold gossip |
| 63 | TEXT_GRISWOLD31 | true | Griswold37 | 95 | Griswold gossip |
| 64 | TEXT_GRISWOLD32 | true | Griswold38 | 101 | Griswold gossip |
| 65 | TEXT_GRISWOLD33 | true | Griswold39 | 169 | Griswold gossip |
| 66 | TEXT_GRISWOLD34 | true | Griswold40 | 129 | Griswold gossip |
| 67 | TEXT_GRISWOLD35 | true | Griswold41 | 94 | Griswold gossip |
| 68 | TEXT_GRISWOLD36 | true | Griswold42 | 130 | Griswold gossip |
| 69 | TEXT_GRISWOLD37 | true | Griswold43 | 123 | Griswold gossip |
| 70 |  | true | Wirt5 | 222 | misc lore/dialog/task text |
| 71 |  | true | Wirt6 | 153 | misc lore/dialog/task text |
| 72 |  | true | Wirt8 | 163 | misc lore/dialog/task text |
| 73 |  | true | Wirt9 | 124 | misc lore/dialog/task text |
| 74 |  | true | Wirt15 | 154 | misc lore/dialog/task text |
| 75 |  | true | Wirt16 | 264 | misc lore/dialog/task text |
| 76 |  | true | Wirt21 | 97 | misc lore/dialog/task text |
| 77 |  | true | Wirt22 | 141 | misc lore/dialog/task text |
| 78 |  | true | Wirt23 | 180 | misc lore/dialog/task text |
| 79 |  | true | Wirt24 | 61 | misc lore/dialog/task text |
| 80 |  | true | Wirt25 | 213 | misc lore/dialog/task text |
| 81 |  | true | Wirt26 | 154 | misc lore/dialog/task text |
| 82 |  | true | Wirt27 | 133 | misc lore/dialog/task text |
| 83 |  | true | Wirt28 | 236 | misc lore/dialog/task text |
| 84 |  | true | Wirt29 | 265 | misc lore/dialog/task text |
| 85 |  | true | Wirt30 | 151 | misc lore/dialog/task text |
| 86 |  | true | Wirt31 | 146 | misc lore/dialog/task text |
| 87 |  | true | Wirt41 | 276 | misc lore/dialog/task text |
| 88 |  | true | Tremain0 | 36 | misc lore/dialog/task text |
| 89 |  | true | Tremain1 | 1016 | misc lore/dialog/task text |
| 90 |  | true | Tremain2 | 105 | misc lore/dialog/task text |
| 91 |  | true | Tremain3 | 277 | misc lore/dialog/task text |
| 92 |  | true | Tremain4 | 308 | misc lore/dialog/task text |
| 93 |  | true | Tremain5 | 1078 | misc lore/dialog/task text |
| 94 |  | true | Tremain6 | 236 | misc lore/dialog/task text |
| 95 |  | true | Tremain7 | 244 | misc lore/dialog/task text |
| 96 |  | true | Cain0 | 243 | misc lore/dialog/task text |
| 97 |  | true | Cain3 | 305 | misc lore/dialog/task text |
| 98 |  | true | Cain5 | 275 | misc lore/dialog/task text |
| 99 |  | true | Cain6 | 552 | misc lore/dialog/task text |
| 100 |  | true | Cain8 | 732 | misc lore/dialog/task text |
| 101 |  | true | Cain9 | 1016 | misc lore/dialog/task text |
| 102 |  | true | Cain11 | 396 | misc lore/dialog/task text |
| 103 |  | true | Cain16 | 634 | misc lore/dialog/task text |
| 104 |  | true | Cain17 | 563 | misc lore/dialog/task text |
| 105 |  | true | Ogden3 | 143 | misc lore/dialog/task text |
| 106 |  | true | Ogden4 | 227 | misc lore/dialog/task text |
| 107 |  | true | Ogden6 | 146 | misc lore/dialog/task text |
| 108 |  | true | Ogden7 | 88 | misc lore/dialog/task text |
| 109 |  | true | Ogden9 | 186 | misc lore/dialog/task text |
| 110 |  | true | Ogden14 | 247 | misc lore/dialog/task text |
| 111 |  | true | Ogden15 | 190 | misc lore/dialog/task text |
| 112 |  | true | Ogden26 | 314 | misc lore/dialog/task text |
| 113 |  | true | Ogden27 | 171 | misc lore/dialog/task text |
| 114 |  | true | Ogden28 | 251 | misc lore/dialog/task text |
| 115 |  | true | Ogden29 | 199 | misc lore/dialog/task text |
| 116 |  | true | Ogden30 | 349 | misc lore/dialog/task text |
| 117 |  | true | Ogden31 | 78 | misc lore/dialog/task text |
| 118 |  | true | Ogden32 | 86 | misc lore/dialog/task text |
| 119 |  | true | Ogden33 | 664 | misc lore/dialog/task text |
| 120 |  | true | Ogden34 | 188 | misc lore/dialog/task text |
#### unused rows 121–139 of 139
| # | txtstrid | scrlltxt | sfxnr | chars | usage class |
|---|---|---|---|---|---|
| 121 |  | true | Ogden35 | 91 | misc lore/dialog/task text |
| 122 |  | true | Adria5 | 141 | misc lore/dialog/task text |
| 123 |  | true | Adria6 | 200 | misc lore/dialog/task text |
| 124 |  | true | Adria8 | 245 | misc lore/dialog/task text |
| 125 |  | true | Adria11 | 123 | misc lore/dialog/task text |
| 126 |  | true | Adria16 | 100 | misc lore/dialog/task text |
| 127 |  | true | Adria17 | 209 | misc lore/dialog/task text |
| 128 |  | true | Adria27 | 181 | misc lore/dialog/task text |
| 129 |  | true | Adria28 | 66 | misc lore/dialog/task text |
| 130 |  | true | Adria29 | 62 | misc lore/dialog/task text |
| 131 |  | true | Adria30 | 176 | misc lore/dialog/task text |
| 132 |  | true | Adria31 | 68 | misc lore/dialog/task text |
| 133 |  | true | Adria32 | 146 | misc lore/dialog/task text |
| 134 |  | true | Adria33 | 188 | misc lore/dialog/task text |
| 135 |  | true | Adria34 | 65 | misc lore/dialog/task text |
| 136 |  | true | Adria35 | 155 | misc lore/dialog/task text |
| 137 |  | true | Adria36 | 112 | misc lore/dialog/task text |
| 138 |  | true | Adria37 | 84 | misc lore/dialog/task text |
| 139 |  | true | Adria48 | 267 | misc lore/dialog/task text |
### V8.3 Menu trees — every item, labels verbatim from the C++
#### Front-end flow (`Source/DiabloUI/`, `UiInitialize` `diabloui.cpp:689`)
```
title.cpp ──any input / attract timeout──▶ mainmenu.cpp
mainmenu.cpp:53-58   [Single Player | Multi Player | Settings | Support | Show Credits | Exit]
  Single Player ▶ hero/selhero.cpp ─▶ hero list / New Hero ▶ Choose Class ▶ Enter Name ▶ game
  Multi Player  ▶ multi/selconn.cpp (connection) ▶ multi/selgame.cpp (create/join) ▶ password
                    ▶ selstart.cpp [Enter Hellfire | Switch to Diablo] ▶ game
  Settings      ▶ settingsmenu.cpp (options tree) ▶ key/pad rebind submenus
  Support       ▶ credits.cpp TextDialog(SupportLines)   [support_lines.cpp:8-24] — 8 lines
  Credits       ▶ credits.cpp CreditLines               [credits_lines.cpp:8+, 559-line file]
  Exit          ▶ quit / save prompt
shared modals: selok.cpp (OK, :86), selyesno.cpp, dialogs.cpp, progress.cpp, button.cpp,
               scrollbar.cpp, text_input.cpp
```
#### Main menu items (`DiabloUI/mainmenu.cpp:53-58`, `_mainmenu_selections`)
- `Single Player` → MAINMENU_SINGLE_PLAYER (`:53`)
- `Multi Player` → MAINMENU_MULTIPLAYER (`:54`)
- `Settings` → MAINMENU_SETTINGS (`:55`)
- `Support` → MAINMENU_SHOW_SUPPORT (`:56`)
- `Show Credits` → MAINMENU_SHOW_CREDITS (`:57`)
- `Exit Diablo` / `Exit Hellfire` → MAINMENU_EXIT_DIABLO (`:58`, label by `gbIsHellfire`, compiled out under `NOEXIT`)
- background art: spawn ⇒ `ui_art\swmmenu`, else `ui_art\mainmenu`(+`mainmenuw.clx` widescreen) (`:60-66`)
#### Hero select (`DiabloUI/hero/selhero.cpp`)
- hero list: each saved hero (`:542`) + `New Hero` (`:546`); header `Single Player Characters` (`:235`) /
  `New Single Player Hero`/`New Multi Player Hero` (`:212,:302`).
- save-exists branch: `Load Game` / `New Game` (`:224-225`), `Save File Exists` title (`:221`),
  `OK`/`Cancel` buttons (`:229,:232`).
- class pick: `Choose Class` (`:172`) listing `playerData.className` per class (`:192`) — gated by
  `classAvailable`/`classAllowed`; shareware lock message for Rogue/Sorcerer (`:294`).
- name entry: `Enter Name` (`:308`), 15-char edit (`:311`, `_uiheroinfo.name[16]`), OK/Cancel (`:314,:317`),
  invalid-name dialog (`:340`).
- stat panel labels: `Level:` `Strength:` `Magic:` `Dexterity:` `Vitality:` `Savegame:` (`:513-519`).
- delete: `UiSelYesNoDialog` (`selyesno.cpp`).
#### Multi player (`DiabloUI/multi/`)
- `selconn.cpp:28-31` `ConnectionNames = { ZeroTier, "Client-Server (TCP)", Offline }` — ZeroTier entry
  added only when `!DISABLE_ZERO_TIER` (`:70-77`); texts `Multi Player Game` (`:72`),
  `Requirements:` (`:78`), `no gateway needed` (`:84`), `Select Connection` (`:90`),
  `Change Gateway` (hidden, `:93`), `OK`/`Cancel` (`:98,:101`); descriptions per provider
  (`:126-134`), `Players Supported: %d` (`:139`).
- `selgame.cpp`: `Description:` (`:144`), `Select Action` (`:150`) with items
  `Create Game`/`Create Public Game`/`Join Game` (`:153-156`), `Select Difficulty` (`:337`) +
  Normal/Nightmare/Hell descriptions (`:401-410`), spawn/full mismatch messages (`:104-106`),
  `Difficulty: %s` (`:249`); password entry via `text_input.cpp`; `selstart.cpp:49-50`
  `Enter Hellfire` / `Switch to Diablo`.
#### Settings (`DiabloUI/settingsmenu.cpp`)
- `Settings` title (`:392`); one row per `Options` tree entry (Volume I §11 lists the tree);
- rebind submenus: `Bound key:` (`:447`), `Press any key to change.` (`:496`), `Unbind key` (`:498`),
  `Bound button combo:` (`:502`), `Unbind button combo` (`:511`), `Previous Menu` (`:555`);
  pad prompt `Press gamepad buttons to change. (N)` (`:170`).
#### In-game pause menu (`Source/gamemenu.cpp`, engine `Source/gmenu.cpp`)
- Single player `sgSingleMenu` (`:49-57`): `Options`, `Save Game`, `Load Game`, `Exit to Main Menu`, `Quit Game`.
- Multiplayer `sgMultiMenu` (`:61-67`): `Options`, `Exit to Main Menu`, `Quit Game`.
- Options `sgOptionsMenu` (`:70-78`): Music slider, Sound slider, `Gamma` slider, `Speed` slider,
  `Previous Menu`; slider names `MusicToggleNames` (`:79-82`) / `SoundToggleNames` (`:83-86`);
  tick-rate slider 20..50 (`:274-277`); `GamemenuUpdateSingle` (`:92`).
- Engine: `gmenu_draw` (`gmenu.cpp:185`), `gmenu_draw_menu_items` (`:251`), `TMenuItem{dwFlags,pszStr,fnMenu}`
  with `GMENU_ENABLED`/`GMENU_SLIDER`, slider API `gmenu.h:101,106`.
#### Control panel (`Source/control/control_panel.cpp`)
- Buttons — `PanBtnStr[8]` (`:111-120`): `Character Information` ('c'), `Quests log` ('q'),
  `Automap` (Tab), `Main Menu` (Esc), `Inventory` ('i'), `Spell book` ('b'), `Send Message` (Enter),
  `` (player-attack); hotkeys `PanBtnHotKey[8]` (`:123-124`); `panel_button_id` enum (`:131-141`);
  SP shows 6 buttons, MP 8 (`TotalSpMainPanelButtons`/`TotalMpMainPanelButtons`, `:126-127`).
- Rectangles: `LevelButtonRect {40,-39,41,22}` (`:97`), `BeltRect {205,5}` 8 slots (`:100-101`),
  `SpellButtonRect {565,64,56,56}` (`:103`), `PanelPaddingHeight 16` (`:105`).
- Flasks (`control_flasks.cpp`), gold (`control_gold.cpp`), info box (`control_infobox.cpp`),
  chat + chat commands (`control_chat*.cpp`).
- Spell book (`panels/spell_book.cpp`, `spell_list.cpp`, `spell_icons.cpp`): two-page list,
  auto-spell selector, per-row name/level/mana from V7; `SpellbookTab` state (`:38`),
  `SpellbookFlag`/`SpellSelectFlag` (`:40,:44`).
#### Other in-game screens
- Character sheet (`panels/charpanel.cpp`): attributes, to-hit/damage/AC/resistances, stat-point
  buttons with `CapStatPointsToAdd` (`control_panel.cpp:154+`), shift-click assign-all (help row).
- Quest log (`panels/quest_log.cpp`): `_qlog` entries ordered by `bookOrder` (V6.1).
- Party panel (`panels/partypanel.cpp`), inventory (`inv.cpp` 40-cell grid + 8 belt), console (`panels/console.cpp`).
- Help (`Source/help.cpp`) — `HelpText[]` (`:28-99`), 48 entries, `$`-prefixed headers:
  Section headers present in `HelpText[]`: `$Keyboard Shortcuts:`, `$Movement:`, `$Combat:`,
  `$Auto-map:`, `$Info Panels/Storage (belt)`, `$Multicolor Icons`, `$Spell Hotkeys`,
  `$Spell Books` (`help.cpp:28-99`; the list is 48 items incl. blanks; layout constants
  `PaddingTop 32`, `PaddingLeft 32`, `PanelHeight 297`, `ContentTextWidth 565` (`:101-106`),
  slider `DrawHelpSlider` (`:143`), open/refresh (`:167`, `:191`).
- Diablo messages/toasts: `diablo_msg.cpp:45` `InitDiabloMsg`, `:125-130` 3500 ms default; every
  `EMSG_*` used by the shrine list (V4.3) and quest text. Automap / inter-level cutscenes
  (`interfac.cpp`, driven by `scrlltxt` rows in V8.1), visual store (`qol/visual_store.cpp`),
  progress UI (`interfac.cpp` + `DiabloUI/progress.cpp`).
- Store screens: 24 TalkID screens (`stores.cpp` `Start*/SStart*` pairs, `:930-1330` + `:2340-2560`) —
  the Wirt/Healer examples in V5.3 show the exact `AddSText` line layout; every store screen uses
  the 20-line `TextLine[]` model with a scrollbar flag (`IsTextFullSize`/`HasScrollbar`).
---
## V9. AUDIO — 814 effects rows (+15 unused), music tracks, movie list
### V9.1 sound/effects.tsv rows 2-815 (814 rows, complete; loader `effects.cpp:141-165`)
Columns `id | flags | path`. Loader: `reader.advance(); // Skip the first column (effect ID)`
(`:151`) — `SfxID` order (`sound_effect_enums.h:117`) is defined by **row order**, and `flags`
is an enum list parsed by `ParseSfxFlag` (`:132-140`: Stream/Misc/Ui/Monk/Rogue/Warrior/Sorcerer).
`sgSFX` holds one `TSFX` per row; `LoadEffectsData` `:141-165`; playback `PlaySFX` (`:208`),
`PlaySfxLoc(psfx, position, randomizeByCategory=true)` (`:216-232` — Walk/ShootBow/CastSpell/Swing
reset `start_tc`), `effect_is_playing` (`:186-197`), `stream_stop` (`:199-206`), streaming
track = `sgpStreamSFX`; cues volume window `SfxID::AccessibilityWeapon…AccessibilityInteract`
(`effects.cpp:100-103`).
**Randomisation pairs** — `RndSFX` (`effects.cpp:106-128`): `PickRandomlyAmong({psfx, psfx+1})`
for Warrior69/Sorceror69/Rogue69/Monk69/Swing/SpellAcid/OperateShrine; `PickRandomlyAmong({psfx,
psfx+1, psfx+2})` for Warrior14/15/16/Warrior2/Rogue14/Sorceror14/Monk14. Entity hooks are the
`SfxID::` call sites across `objects.cpp` (door/chest/sarcophagus/shrine — see V4.2),
`monster.cpp` (greetings/deaths), `player.cpp` (swing/hit/death per class), `spells.cpp`
(`missiles` castSound/hitSound columns, V3.1), `towners.cpp` (cow/moo), `stores.cpp`, `help.cpp`,
`gamemenu.cpp`, `cursor.cpp`, `diablo.cpp` (level-in/quest).
Flag histogram: `Stream`=390, `Misc`=107, `Warrior`=62, `Sorcerer`=56, `Rogue`=56, `Stream,Sorcerer`=47, `Stream,Rogue`=47, `Stream,Warrior`=47, `Ui`=2 (total 814).
#### flag = `Misc` — 107 rows
| # | id | path |
|---|---|---|
| 1 | Walk | sfx\misc\walk1.wav |
| 2 | ShootBow | sfx\misc\bfire.wav |
| 3 | CastSpell | sfx\misc\tmag.wav |
| 4 | Swing | sfx\misc\swing.wav |
| 5 | Swing2 | sfx\misc\swing2.wav |
| 6 | WarriorDeath | sfx\misc\dead.wav |
| 8 | BarrelExpload | sfx\items\barlfire.wav |
| 9 | BarrelBreak | sfx\items\barrel.wav |
| 10 | ChestOpen | sfx\items\chest.wav |
| 11 | DoorClose | sfx\items\doorclos.wav |
| 12 | DoorOpen | sfx\items\dooropen.wav |
| 13 | ItemAnvilFlip | sfx\items\flipanvl.wav |
| 14 | ItemAxeFlip | sfx\items\flipaxe.wav |
| 15 | ItemBloodStoneFlip | sfx\items\flipblst.wav |
| 16 | ItemBodyPartFlip | sfx\items\flipbody.wav |
| 17 | ItemBookFlip | sfx\items\flipbook.wav |
| 18 | ItemBowFlip | sfx\items\flipbow.wav |
| 19 | ItemCapFlip | sfx\items\flipcap.wav |
| 20 | ItemArmorFlip | sfx\items\flipharm.wav |
| 21 | ItemLeatherFlip | sfx\items\fliplarm.wav |
| 22 | ItemMushroomFlip | sfx\items\flipmush.wav |
| 23 | ItemPotionFlip | sfx\items\flippot.wav |
| 24 | ItemRingFlip | sfx\items\flipring.wav |
| 25 | ItemRockFlip | sfx\items\fliprock.wav |
| 26 | ItemScrollFlip | sfx\items\flipscrl.wav |
| 27 | ItemShieldFlip | sfx\items\flipshld.wav |
| 28 | ItemSignFlip | sfx\items\flipsign.wav |
| 29 | ItemStaffFlip | sfx\items\flipstaf.wav |
| 30 | ItemSwordFlip | sfx\items\flipswor.wav |
| 31 | ItemGold | sfx\items\gold.wav |
| 32 | ItemAnvil | sfx\items\invanvl.wav |
| 33 | ItemAxe | sfx\items\invaxe.wav |
| 34 | ItemBloodStone | sfx\items\invblst.wav |
| 35 | ItemBodyPart | sfx\items\invbody.wav |
| 36 | ItemBook | sfx\items\invbook.wav |
| 37 | ItemBow | sfx\items\invbow.wav |
| 38 | ItemCap | sfx\items\invcap.wav |
| 39 | GrabItem | sfx\items\invgrab.wav |
| 40 | ItemArmor | sfx\items\invharm.wav |
| 41 | ItemLeather | sfx\items\invlarm.wav |
| 42 | ItemMushroom | sfx\items\invmush.wav |
| 43 | ItemPotion | sfx\items\invpot.wav |
| 44 | ItemRing | sfx\items\invring.wav |
| 45 | ItemRock | sfx\items\invrock.wav |
| 46 | ItemScroll | sfx\items\invscrol.wav |
| 47 | ItemShield | sfx\items\invshiel.wav |
| 48 | ItemSign | sfx\items\invsign.wav |
| 49 | ItemStaff | sfx\items\invstaf.wav |
| 50 | ItemSword | sfx\items\invsword.wav |
| 51 | OperateLever | sfx\items\lever.wav |
| 52 | OperateShrine | sfx\items\magic.wav |
| 53 | OperateShrine1 | sfx\items\magic1.wav |
| 54 | ReadBook | sfx\items\readbook.wav |
| 55 | Sarcophagus | sfx\items\sarc.wav |
| 58 | TriggerTrap | sfx\items\trap.wav |
| 59 | CastFire | sfx\misc\cast2.wav |
| 60 | CastLightning | sfx\misc\cast4.wav |
| 61 | CastSkill | sfx\misc\cast6.wav |
| 62 | SpellEnd | sfx\misc\cast7.wav |
| 63 | CastHealing | sfx\misc\cast8.wav |
| 64 | SpellRepair | sfx\misc\repair.wav |
| 65 | SpellAcid | sfx\misc\acids1.wav |
| 66 | SpellAcid1 | sfx\misc\acids2.wav |
| 67 | SpellApocalypse | sfx\misc\apoc.wav |
| 68 | SpellBloodStar | sfx\misc\blodstar.wav |
| 69 | SpellBloodStarHit | sfx\misc\blsimpt.wav |
| 70 | SpellBoneSpirit | sfx\misc\bonesp.wav |
| 71 | SpellBoneSpiritHit | sfx\misc\bsimpct.wav |
| 72 | OperateCaldron | sfx\misc\caldron.wav |
| 73 | SpellChargedBolt | sfx\misc\cbolt.wav |
| 74 | SpellDoomSerpents | sfx\misc\dserp.wav |
| 75 | SpellLightningHit | sfx\misc\elecimp1.wav |
| 76 | SpellElemental | sfx\misc\elementl.wav |
| 77 | SpellEtherealize | sfx\misc\ethereal.wav |
| 78 | SpellFirebolt | sfx\misc\fbolt1.wav |
| 79 | SpellFireHit | sfx\misc\firimp2.wav |
| 80 | SpellFlameWave | sfx\misc\flamwave.wav |
| 81 | OperateFountain | sfx\misc\fountain.wav |
| 82 | SpellGolem | sfx\misc\golum.wav |
| 83 | OperateGoatShrine | sfx\misc\gshrine.wav |
| 84 | SpellGuardian | sfx\misc\guard.wav |
| 85 | SpellHolyBolt | sfx\misc\holybolt.wav |
| 86 | SpellInfravision | sfx\misc\infravis.wav |
| 87 | SpellInvisibility | sfx\misc\invisibl.wav |
| 88 | SpellLightning | sfx\misc\lning1.wav |
| 89 | SpellManaShield | sfx\misc\mshield.wav |
| 90 | SpellNova | sfx\misc\nova.wav |
| 91 | SpellPuddle | sfx\misc\puddle.wav |
| 92 | SpellResurrect | sfx\misc\resur.wav |
| 93 | SpellStoneCurse | sfx\misc\scurimp.wav |
| 94 | SpellPortal | sfx\misc\sentinel.wav |
| 95 | SpellInferno | sfx\misc\spoutstr.wav |
| 96 | SpellTrapDisarm | sfx\misc\trapdis.wav |
| 97 | SpellTeleport | sfx\misc\teleport.wav |
| 98 | SpellFireWall | sfx\misc\wallloop.wav |
| 193 | Cow1 | sfx\towners\cow1.wav |
| 194 | Cow2 | sfx\towners\cow2.wav |
| 805 | AccessibilityWeapon | sfx\items\titlemov.wav |
| 806 | AccessibilityArmor | sfx\items\titlemov.wav |
| 807 | AccessibilityGold | sfx\items\titlemov.wav |
| 808 | AccessibilityPotion | sfx\items\titlemov.wav |
| 809 | AccessibilityScroll | sfx\items\titlemov.wav |
| 810 | AccessibilityChest | sfx\items\titlemov.wav |
| 811 | AccessibilityDoor | sfx\items\titlemov.wav |
| 812 | AccessibilityStairs | sfx\items\titlemov.wav |
| 813 | AccessibilityMonster | sfx\items\titlemov.wav |
| 814 | AccessibilityInteract | sfx\items\titlemov.wav |
#### flag = `Stream` — 390 rows
| # | id | path |
|---|---|---|
| 7 | QuestDone | sfx\misc\questdon.wav |
| 99 | Gillian1 | sfx\towners\bmaid01.wav |
| 100 | Gillian2 | sfx\towners\bmaid02.wav |
| 101 | Gillian3 | sfx\towners\bmaid03.wav |
| 102 | Gillian4 | sfx\towners\bmaid04.wav |
| 103 | Gillian5 | sfx\towners\bmaid05.wav |
| 104 | Gillian6 | sfx\towners\bmaid06.wav |
| 105 | Gillian7 | sfx\towners\bmaid07.wav |
| 106 | Gillian8 | sfx\towners\bmaid08.wav |
| 107 | Gillian9 | sfx\towners\bmaid09.wav |
| 108 | Gillian10 | sfx\towners\bmaid10.wav |
| 109 | Gillian11 | sfx\towners\bmaid11.wav |
| 110 | Gillian12 | sfx\towners\bmaid12.wav |
| 111 | Gillian13 | sfx\towners\bmaid13.wav |
| 112 | Gillian14 | sfx\towners\bmaid14.wav |
| 113 | Gillian15 | sfx\towners\bmaid15.wav |
| 114 | Gillian16 | sfx\towners\bmaid16.wav |
| 115 | Gillian17 | sfx\towners\bmaid17.wav |
| 116 | Gillian18 | sfx\towners\bmaid18.wav |
| 117 | Gillian19 | sfx\towners\bmaid19.wav |
| 118 | Gillian20 | sfx\towners\bmaid20.wav |
| 119 | Gillian21 | sfx\towners\bmaid21.wav |
| 120 | Gillian22 | sfx\towners\bmaid22.wav |
| 121 | Gillian23 | sfx\towners\bmaid23.wav |
| 122 | Gillian24 | sfx\towners\bmaid24.wav |
| 123 | Gillian25 | sfx\towners\bmaid25.wav |
| 124 | Gillian26 | sfx\towners\bmaid26.wav |
| 125 | Gillian27 | sfx\towners\bmaid27.wav |
| 126 | Gillian28 | sfx\towners\bmaid28.wav |
| 127 | Gillian29 | sfx\towners\bmaid29.wav |
| 128 | Gillian30 | sfx\towners\bmaid30.wav |
| 129 | Gillian31 | sfx\towners\bmaid31.wav |
| 130 | Gillian32 | sfx\towners\bmaid32.wav |
| 131 | Gillian33 | sfx\towners\bmaid33.wav |
| 132 | Gillian34 | sfx\towners\bmaid34.wav |
| 133 | Gillian35 | sfx\towners\bmaid35.wav |
| 134 | Gillian36 | sfx\towners\bmaid36.wav |
| 135 | Gillian37 | sfx\towners\bmaid37.wav |
| 136 | Gillian38 | sfx\towners\bmaid38.wav |
| 137 | Gillian39 | sfx\towners\bmaid39.wav |
| 138 | Gillian40 | sfx\towners\bmaid40.wav |
| 139 | Griswold1 | sfx\towners\bsmith01.wav |
| 140 | Griswold2 | sfx\towners\bsmith02.wav |
| 141 | Griswold3 | sfx\towners\bsmith03.wav |
| 142 | Griswold4 | sfx\towners\bsmith04.wav |
| 143 | Griswold5 | sfx\towners\bsmith05.wav |
| 144 | Griswold6 | sfx\towners\bsmith06.wav |
| 145 | Griswold7 | sfx\towners\bsmith07.wav |
| 146 | Griswold8 | sfx\towners\bsmith08.wav |
| 147 | Griswold9 | sfx\towners\bsmith09.wav |
| 148 | Griswold10 | sfx\towners\bsmith10.wav |
| 149 | Griswold12 | sfx\towners\bsmith12.wav |
| 150 | Griswold13 | sfx\towners\bsmith13.wav |
| 151 | Griswold14 | sfx\towners\bsmith14.wav |
| 152 | Griswold15 | sfx\towners\bsmith15.wav |
| 153 | Griswold16 | sfx\towners\bsmith16.wav |
| 154 | Griswold17 | sfx\towners\bsmith17.wav |
| 155 | Griswold18 | sfx\towners\bsmith18.wav |
| 156 | Griswold19 | sfx\towners\bsmith19.wav |
| 157 | Griswold20 | sfx\towners\bsmith20.wav |
| 158 | Griswold21 | sfx\towners\bsmith21.wav |
| 159 | Griswold22 | sfx\towners\bsmith22.wav |
| 160 | Griswold23 | sfx\towners\bsmith23.wav |
| 161 | Griswold24 | sfx\towners\bsmith24.wav |
| 162 | Griswold25 | sfx\towners\bsmith25.wav |
| 163 | Griswold26 | sfx\towners\bsmith26.wav |
| 164 | Griswold27 | sfx\towners\bsmith27.wav |
| 165 | Griswold28 | sfx\towners\bsmith28.wav |
| 166 | Griswold29 | sfx\towners\bsmith29.wav |
| 167 | Griswold30 | sfx\towners\bsmith30.wav |
| 168 | Griswold31 | sfx\towners\bsmith31.wav |
| 169 | Griswold32 | sfx\towners\bsmith32.wav |
| 170 | Griswold33 | sfx\towners\bsmith33.wav |
| 171 | Griswold34 | sfx\towners\bsmith34.wav |
| 172 | Griswold35 | sfx\towners\bsmith35.wav |
| 173 | Griswold36 | sfx\towners\bsmith36.wav |
| 174 | Griswold37 | sfx\towners\bsmith37.wav |
| 175 | Griswold38 | sfx\towners\bsmith38.wav |
| 176 | Griswold39 | sfx\towners\bsmith39.wav |
| 177 | Griswold40 | sfx\towners\bsmith40.wav |
| 178 | Griswold41 | sfx\towners\bsmith41.wav |
| 179 | Griswold42 | sfx\towners\bsmith42.wav |
| 180 | Griswold43 | sfx\towners\bsmith43.wav |
| 181 | Griswold44 | sfx\towners\bsmith44.wav |
| 182 | Griswold45 | sfx\towners\bsmith45.wav |
| 183 | Griswold46 | sfx\towners\bsmith46.wav |
| 184 | Griswold47 | sfx\towners\bsmith47.wav |
| 185 | Griswold48 | sfx\towners\bsmith48.wav |
| 186 | Griswold49 | sfx\towners\bsmith49.wav |
| 187 | Griswold50 | sfx\towners\bsmith50.wav |
| 188 | Griswold51 | sfx\towners\bsmith51.wav |
| 189 | Griswold52 | sfx\towners\bsmith52.wav |
| 190 | Griswold53 | sfx\towners\bsmith53.wav |
| 191 | Griswold55 | sfx\towners\bsmith55.wav |
| 192 | Griswold56 | sfx\towners\bsmith56.wav |
| 195 | WoundedTownsmanOld | sfx\towners\deadguy2.wav |
| 196 | Farnham1 | sfx\towners\drunk01.wav |
| 197 | Farnham2 | sfx\towners\drunk02.wav |
| 198 | Farnham3 | sfx\towners\drunk03.wav |
| 199 | Farnham4 | sfx\towners\drunk04.wav |
| 200 | Farnham5 | sfx\towners\drunk05.wav |
| 201 | Farnham6 | sfx\towners\drunk06.wav |
| 202 | Farnham7 | sfx\towners\drunk07.wav |
| 203 | Farnham8 | sfx\towners\drunk08.wav |
| 204 | Farnham9 | sfx\towners\drunk09.wav |
| 205 | Farnham10 | sfx\towners\drunk10.wav |
| 206 | Farnham11 | sfx\towners\drunk11.wav |
| 207 | Farnham12 | sfx\towners\drunk12.wav |
| 208 | Farnham13 | sfx\towners\drunk13.wav |
| 209 | Farnham14 | sfx\towners\drunk14.wav |
| 210 | Farnham15 | sfx\towners\drunk15.wav |
| 211 | Farnham16 | sfx\towners\drunk16.wav |
| 212 | Farnham17 | sfx\towners\drunk17.wav |
| 213 | Farnham18 | sfx\towners\drunk18.wav |
| 214 | Farnham19 | sfx\towners\drunk19.wav |
| 215 | Farnham20 | sfx\towners\drunk20.wav |
| 216 | Farnham21 | sfx\towners\drunk21.wav |
| 217 | Farnham22 | sfx\towners\drunk22.wav |
| 218 | Farnham23 | sfx\towners\drunk23.wav |
| 219 | Farnham24 | sfx\towners\drunk24.wav |
| 220 | Farnham25 | sfx\towners\drunk25.wav |
| 221 | Farnham26 | sfx\towners\drunk26.wav |
| 222 | Farnham27 | sfx\towners\drunk27.wav |
| 223 | Farnham28 | sfx\towners\drunk28.wav |
| 224 | Farnham29 | sfx\towners\drunk29.wav |
| 225 | Farnham30 | sfx\towners\drunk30.wav |
| 226 | Farnham31 | sfx\towners\drunk31.wav |
| 227 | Farnham32 | sfx\towners\drunk32.wav |
| 228 | Farnham33 | sfx\towners\drunk33.wav |
| 229 | Farnham34 | sfx\towners\drunk34.wav |
| 230 | Farnham35 | sfx\towners\drunk35.wav |
| 231 | Pepin1 | sfx\towners\healer01.wav |
| 232 | Pepin2 | sfx\towners\healer02.wav |
| 233 | Pepin3 | sfx\towners\healer03.wav |
| 234 | Pepin4 | sfx\towners\healer04.wav |
| 235 | Pepin5 | sfx\towners\healer05.wav |
| 236 | Pepin6 | sfx\towners\healer06.wav |
| 237 | Pepin7 | sfx\towners\healer07.wav |
| 238 | Pepin8 | sfx\towners\healer08.wav |
| 239 | Pepin9 | sfx\towners\healer09.wav |
| 240 | Pepin10 | sfx\towners\healer10.wav |
| 241 | Pepin11 | sfx\towners\healer11.wav |
| 242 | Pepin12 | sfx\towners\healer12.wav |
| 243 | Pepin13 | sfx\towners\healer13.wav |
| 244 | Pepin14 | sfx\towners\healer14.wav |
| 245 | Pepin15 | sfx\towners\healer15.wav |
| 246 | Pepin16 | sfx\towners\healer16.wav |
| 247 | Pepin17 | sfx\towners\healer17.wav |
| 248 | Pepin18 | sfx\towners\healer18.wav |
| 249 | Pepin19 | sfx\towners\healer19.wav |
| 250 | Pepin20 | sfx\towners\healer20.wav |
| 251 | Pepin21 | sfx\towners\healer21.wav |
| 252 | Pepin22 | sfx\towners\healer22.wav |
| 253 | Pepin23 | sfx\towners\healer23.wav |
| 254 | Pepin24 | sfx\towners\healer24.wav |
| 255 | Pepin25 | sfx\towners\healer25.wav |
| 256 | Pepin26 | sfx\towners\healer26.wav |
| 257 | Pepin27 | sfx\towners\healer27.wav |
| 258 | Pepin28 | sfx\towners\healer28.wav |
| 259 | Pepin29 | sfx\towners\healer29.wav |
| 260 | Pepin30 | sfx\towners\healer30.wav |
| 261 | Pepin31 | sfx\towners\healer31.wav |
| 262 | Pepin32 | sfx\towners\healer32.wav |
| 263 | Pepin33 | sfx\towners\healer33.wav |
| 264 | Pepin34 | sfx\towners\healer34.wav |
| 265 | Pepin35 | sfx\towners\healer35.wav |
| 266 | Pepin36 | sfx\towners\healer36.wav |
| 267 | Pepin37 | sfx\towners\healer37.wav |
| 268 | Pepin38 | sfx\towners\healer38.wav |
| 269 | Pepin39 | sfx\towners\healer39.wav |
| 270 | Pepin40 | sfx\towners\healer40.wav |
| 271 | Pepin41 | sfx\towners\healer41.wav |
| 272 | Pepin42 | sfx\towners\healer42.wav |
| 273 | Pepin43 | sfx\towners\healer43.wav |
| 274 | Pepin44 | sfx\towners\healer44.wav |
| 275 | Pepin45 | sfx\towners\healer45.wav |
| 276 | Pepin46 | sfx\towners\healer46.wav |
| 277 | Pepin47 | sfx\towners\healer47.wav |
| 278 | Wirt1 | sfx\towners\pegboy01.wav |
| 279 | Wirt2 | sfx\towners\pegboy02.wav |
| 280 | Wirt3 | sfx\towners\pegboy03.wav |
| 281 | Wirt4 | sfx\towners\pegboy04.wav |
| 282 | Wirt5 | sfx\towners\pegboy05.wav |
| 283 | Wirt6 | sfx\towners\pegboy06.wav |
| 284 | Wirt7 | sfx\towners\pegboy07.wav |
| 285 | Wirt8 | sfx\towners\pegboy08.wav |
| 286 | Wirt9 | sfx\towners\pegboy09.wav |
| 287 | Wirt10 | sfx\towners\pegboy10.wav |
| 288 | Wirt11 | sfx\towners\pegboy11.wav |
| 289 | Wirt12 | sfx\towners\pegboy12.wav |
| 290 | Wirt13 | sfx\towners\pegboy13.wav |
| 291 | Wirt14 | sfx\towners\pegboy14.wav |
| 292 | Wirt15 | sfx\towners\pegboy15.wav |
| 293 | Wirt16 | sfx\towners\pegboy16.wav |
| 294 | Wirt17 | sfx\towners\pegboy17.wav |
| 295 | Wirt18 | sfx\towners\pegboy18.wav |
| 296 | Wirt19 | sfx\towners\pegboy19.wav |
| 297 | Wirt20 | sfx\towners\pegboy20.wav |
| 298 | Wirt21 | sfx\towners\pegboy21.wav |
| 299 | Wirt22 | sfx\towners\pegboy22.wav |
| 300 | Wirt23 | sfx\towners\pegboy23.wav |
| 301 | Wirt24 | sfx\towners\pegboy24.wav |
| 302 | Wirt25 | sfx\towners\pegboy25.wav |
| 303 | Wirt26 | sfx\towners\pegboy26.wav |
| 304 | Wirt27 | sfx\towners\pegboy27.wav |
| 305 | Wirt28 | sfx\towners\pegboy28.wav |
| 306 | Wirt29 | sfx\towners\pegboy29.wav |
| 307 | Wirt30 | sfx\towners\pegboy30.wav |
| 308 | Wirt31 | sfx\towners\pegboy31.wav |
| 309 | Wirt32 | sfx\towners\pegboy32.wav |
| 310 | Wirt33 | sfx\towners\pegboy33.wav |
| 311 | Wirt34 | sfx\towners\pegboy34.wav |
| 312 | Wirt35 | sfx\towners\pegboy35.wav |
| 313 | Wirt36 | sfx\towners\pegboy36.wav |
| 314 | Wirt37 | sfx\towners\pegboy37.wav |
| 315 | Wirt38 | sfx\towners\pegboy38.wav |
| 316 | Wirt39 | sfx\towners\pegboy39.wav |
| 317 | Wirt40 | sfx\towners\pegboy40.wav |
| 318 | Wirt41 | sfx\towners\pegboy41.wav |
| 319 | Wirt42 | sfx\towners\pegboy42.wav |
| 320 | Wirt43 | sfx\towners\pegboy43.wav |
| 321 | Tremain0 | sfx\towners\priest00.wav |
| 322 | Tremain1 | sfx\towners\priest01.wav |
| 323 | Tremain2 | sfx\towners\priest02.wav |
| 324 | Tremain3 | sfx\towners\priest03.wav |
| 325 | Tremain4 | sfx\towners\priest04.wav |
| 326 | Tremain5 | sfx\towners\priest05.wav |
| 327 | Tremain6 | sfx\towners\priest06.wav |
| 328 | Tremain7 | sfx\towners\priest07.wav |
| 329 | Cain0 | sfx\towners\storyt00.wav |
| 330 | Cain1 | sfx\towners\storyt01.wav |
| 331 | Cain2 | sfx\towners\storyt02.wav |
| 332 | Cain3 | sfx\towners\storyt03.wav |
| 333 | Cain4 | sfx\towners\storyt04.wav |
| 334 | Cain5 | sfx\towners\storyt05.wav |
| 335 | Cain6 | sfx\towners\storyt06.wav |
| 336 | Cain7 | sfx\towners\storyt07.wav |
| 337 | Cain8 | sfx\towners\storyt08.wav |
| 338 | Cain9 | sfx\towners\storyt09.wav |
| 339 | Cain10 | sfx\towners\storyt10.wav |
| 340 | Cain11 | sfx\towners\storyt11.wav |
| 341 | Cain12 | sfx\towners\storyt12.wav |
| 342 | Cain13 | sfx\towners\storyt13.wav |
| 343 | Cain14 | sfx\towners\storyt14.wav |
| 344 | Cain15 | sfx\towners\storyt15.wav |
| 345 | Cain16 | sfx\towners\storyt16.wav |
| 346 | Cain17 | sfx\towners\storyt17.wav |
| 347 | Cain18 | sfx\towners\storyt18.wav |
| 348 | Cain19 | sfx\towners\storyt19.wav |
| 349 | Cain20 | sfx\towners\storyt20.wav |
| 350 | Cain21 | sfx\towners\storyt21.wav |
| 351 | Cain22 | sfx\towners\storyt22.wav |
| 352 | Cain23 | sfx\towners\storyt23.wav |
| 353 | Cain24 | sfx\towners\storyt24.wav |
| 354 | Cain25 | sfx\towners\storyt25.wav |
| 355 | Cain26 | sfx\towners\storyt26.wav |
| 356 | Cain27 | sfx\towners\storyt27.wav |
| 357 | Cain28 | sfx\towners\storyt28.wav |
| 358 | Cain29 | sfx\towners\storyt29.wav |
| 359 | Cain30 | sfx\towners\storyt30.wav |
| 360 | Cain31 | sfx\towners\storyt31.wav |
| 361 | Cain33 | sfx\towners\storyt33.wav |
| 362 | Cain34 | sfx\towners\storyt34.wav |
| 363 | Cain35 | sfx\towners\storyt35.wav |
| 364 | Cain36 | sfx\towners\storyt36.wav |
| 365 | Cain37 | sfx\towners\storyt37.wav |
| 366 | Cain38 | sfx\towners\storyt38.wav |
| 367 | Ogden0 | sfx\towners\tavown00.wav |
| 368 | Ogden1 | sfx\towners\tavown01.wav |
| 369 | Ogden2 | sfx\towners\tavown02.wav |
| 370 | Ogden3 | sfx\towners\tavown03.wav |
| 371 | Ogden4 | sfx\towners\tavown04.wav |
| 372 | Ogden5 | sfx\towners\tavown05.wav |
| 373 | Ogden6 | sfx\towners\tavown06.wav |
| 374 | Ogden7 | sfx\towners\tavown07.wav |
| 375 | Ogden8 | sfx\towners\tavown08.wav |
| 376 | Ogden9 | sfx\towners\tavown09.wav |
| 377 | Ogden10 | sfx\towners\tavown10.wav |
| 378 | Ogden11 | sfx\towners\tavown11.wav |
| 379 | Ogden12 | sfx\towners\tavown12.wav |
| 380 | Ogden13 | sfx\towners\tavown13.wav |
| 381 | Ogden14 | sfx\towners\tavown14.wav |
| 382 | Ogden15 | sfx\towners\tavown15.wav |
| 383 | Ogden16 | sfx\towners\tavown16.wav |
| 384 | Ogden17 | sfx\towners\tavown17.wav |
| 385 | Ogden18 | sfx\towners\tavown18.wav |
| 386 | Ogden19 | sfx\towners\tavown19.wav |
| 387 | Ogden20 | sfx\towners\tavown20.wav |
| 388 | Ogden21 | sfx\towners\tavown21.wav |
| 389 | Ogden22 | sfx\towners\tavown22.wav |
| 390 | Ogden23 | sfx\towners\tavown23.wav |
| 391 | Ogden24 | sfx\towners\tavown24.wav |
| 392 | Ogden25 | sfx\towners\tavown25.wav |
| 393 | Ogden26 | sfx\towners\tavown26.wav |
| 394 | Ogden27 | sfx\towners\tavown27.wav |
| 395 | Ogden28 | sfx\towners\tavown28.wav |
| 396 | Ogden29 | sfx\towners\tavown29.wav |
| 397 | Ogden30 | sfx\towners\tavown30.wav |
| 398 | Ogden31 | sfx\towners\tavown31.wav |
| 399 | Ogden32 | sfx\towners\tavown32.wav |
| 400 | Ogden33 | sfx\towners\tavown33.wav |
| 401 | Ogden34 | sfx\towners\tavown34.wav |
| 402 | Ogden35 | sfx\towners\tavown35.wav |
| 403 | Ogden36 | sfx\towners\tavown36.wav |
| 404 | Ogden37 | sfx\towners\tavown37.wav |
| 405 | Ogden38 | sfx\towners\tavown38.wav |
| 406 | Ogden39 | sfx\towners\tavown39.wav |
| 407 | Ogden40 | sfx\towners\tavown40.wav |
| 408 | Ogden41 | sfx\towners\tavown41.wav |
| 409 | Ogden43 | sfx\towners\tavown43.wav |
| 410 | Ogden44 | sfx\towners\tavown44.wav |
| 411 | Ogden45 | sfx\towners\tavown45.wav |
| 412 | Adria1 | sfx\towners\witch01.wav |
| 413 | Adria2 | sfx\towners\witch02.wav |
| 414 | Adria3 | sfx\towners\witch03.wav |
| 415 | Adria4 | sfx\towners\witch04.wav |
| 416 | Adria5 | sfx\towners\witch05.wav |
| 417 | Adria6 | sfx\towners\witch06.wav |
| 418 | Adria7 | sfx\towners\witch07.wav |
| 419 | Adria8 | sfx\towners\witch08.wav |
| 420 | Adria9 | sfx\towners\witch09.wav |
| 421 | Adria10 | sfx\towners\witch10.wav |
| 422 | Adria11 | sfx\towners\witch11.wav |
| 423 | Adria12 | sfx\towners\witch12.wav |
| 424 | Adria13 | sfx\towners\witch13.wav |
| 425 | Adria14 | sfx\towners\witch14.wav |
| 426 | Adria15 | sfx\towners\witch15.wav |
| 427 | Adria16 | sfx\towners\witch16.wav |
| 428 | Adria17 | sfx\towners\witch17.wav |
| 429 | Adria18 | sfx\towners\witch18.wav |
| 430 | Adria19 | sfx\towners\witch19.wav |
| 431 | Adria20 | sfx\towners\witch20.wav |
| 432 | Adria21 | sfx\towners\witch21.wav |
| 433 | Adria22 | sfx\towners\witch22.wav |
| 434 | Adria23 | sfx\towners\witch23.wav |
| 435 | Adria24 | sfx\towners\witch24.wav |
| 436 | Adria25 | sfx\towners\witch25.wav |
| 437 | Adria26 | sfx\towners\witch26.wav |
| 438 | Adria27 | sfx\towners\witch27.wav |
| 439 | Adria28 | sfx\towners\witch28.wav |
| 440 | Adria29 | sfx\towners\witch29.wav |
| 441 | Adria30 | sfx\towners\witch30.wav |
| 442 | Adria31 | sfx\towners\witch31.wav |
| 443 | Adria32 | sfx\towners\witch32.wav |
| 444 | Adria33 | sfx\towners\witch33.wav |
| 445 | Adria34 | sfx\towners\witch34.wav |
| 446 | Adria35 | sfx\towners\witch35.wav |
| 447 | Adria36 | sfx\towners\witch36.wav |
| 448 | Adria37 | sfx\towners\witch37.wav |
| 449 | Adria38 | sfx\towners\witch38.wav |
| 450 | Adria39 | sfx\towners\witch39.wav |
| 451 | Adria40 | sfx\towners\witch40.wav |
| 452 | Adria41 | sfx\towners\witch41.wav |
| 453 | Adria42 | sfx\towners\witch42.wav |
| 454 | Adria43 | sfx\towners\witch43.wav |
| 455 | Adria44 | sfx\towners\witch44.wav |
| 456 | Adria45 | sfx\towners\witch45.wav |
| 457 | Adria46 | sfx\towners\witch46.wav |
| 458 | Adria47 | sfx\towners\witch47.wav |
| 459 | Adria48 | sfx\towners\witch48.wav |
| 460 | Adria49 | sfx\towners\witch49.wav |
| 461 | Adria50 | sfx\towners\witch50.wav |
| 462 | WoundedTownsman | sfx\towners\wound01.wav |
| 778 | Narrator1 | sfx\narrator\nar01.wav |
| 779 | Narrator2 | sfx\narrator\nar02.wav |
| 780 | Narrator3 | sfx\narrator\nar03.wav |
| 781 | Narrator4 | sfx\narrator\nar04.wav |
| 782 | Narrator5 | sfx\narrator\nar05.wav |
| 783 | Narrator6 | sfx\narrator\nar06.wav |
| 784 | Narrator7 | sfx\narrator\nar07.wav |
| 785 | Narrator8 | sfx\narrator\nar08.wav |
| 786 | Narrator9 | sfx\narrator\nar09.wav |
| 787 | DiabloGreeting | sfx\misc\lvl16int.wav |
| 788 | ButcherGreeting | sfx\monsters\butcher.wav |
| 789 | Gharbad1 | sfx\monsters\garbud01.wav |
| 790 | Gharbad2 | sfx\monsters\garbud02.wav |
| 791 | Gharbad3 | sfx\monsters\garbud03.wav |
| 792 | Gharbad4 | sfx\monsters\garbud04.wav |
| 793 | Lachdanan1 | sfx\monsters\lach01.wav |
| 794 | Lachdanan2 | sfx\monsters\lach02.wav |
| 795 | Lachdanan3 | sfx\monsters\lach03.wav |
| 796 | LazarusGreeting | sfx\monsters\laz01.wav |
| 797 | LeoricGreeting | sfx\monsters\sking01.wav |
| 798 | Snotspill1 | sfx\monsters\snot01.wav |
| 799 | Snotspill2 | sfx\monsters\snot02.wav |
| 800 | Snotspill3 | sfx\monsters\snot03.wav |
| 801 | Warlord | sfx\monsters\warlrd01.wav |
| 802 | Zhar1 | sfx\monsters\zhar01.wav |
| 803 | Zhar2 | sfx\monsters\zhar02.wav |
| 804 | DiabloDeath | sfx\monsters\diablod.wav |
#### flag = `Ui` — 2 rows
| # | id | path |
|---|---|---|
| 56 | MenuMove | sfx\items\titlemov.wav |
| 57 | MenuSelect | sfx\items\titlslct.wav |
#### flag = `Stream,Sorcerer` — 47 rows
| # | id | path |
|---|---|---|
| 463 | Sorceror1 | sfx\sorceror\mage01.wav |
| 464 | Sorceror2 | sfx\sorceror\mage02.wav |
| 465 | Sorceror3 | sfx\sorceror\mage03.wav |
| 466 | Sorceror4 | sfx\sorceror\mage04.wav |
| 467 | Sorceror5 | sfx\sorceror\mage05.wav |
| 468 | Sorceror6 | sfx\sorceror\mage06.wav |
| 469 | Sorceror7 | sfx\sorceror\mage07.wav |
| 470 | Sorceror8 | sfx\sorceror\mage08.wav |
| 471 | Sorceror9 | sfx\sorceror\mage09.wav |
| 472 | Sorceror10 | sfx\sorceror\mage10.wav |
| 473 | Sorceror11 | sfx\sorceror\mage11.wav |
| 474 | Sorceror12 | sfx\sorceror\mage12.wav |
| 513 | Sorceror51 | sfx\sorceror\mage51.wav |
| 514 | Sorceror52 | sfx\sorceror\mage52.wav |
| 515 | Sorceror53 | sfx\sorceror\mage53.wav |
| 516 | Sorceror54 | sfx\sorceror\mage54.wav |
| 517 | Sorceror55 | sfx\sorceror\mage55.wav |
| 518 | Sorceror56 | sfx\sorceror\mage56.wav |
| 520 | Sorceror58 | sfx\sorceror\mage58.wav |
| 521 | Sorceror59 | sfx\sorceror\mage59.wav |
| 522 | Sorceror60 | sfx\sorceror\mage60.wav |
| 523 | Sorceror61 | sfx\sorceror\mage61.wav |
| 524 | Sorceror62 | sfx\sorceror\mage62.wav |
| 525 | Sorceror63 | sfx\sorceror\mage63.wav |
| 543 | Sorceror80 | sfx\sorceror\mage80.wav |
| 544 | Sorceror81 | sfx\sorceror\mage81.wav |
| 545 | Sorceror82 | sfx\sorceror\mage82.wav |
| 546 | Sorceror83 | sfx\sorceror\mage83.wav |
| 547 | Sorceror84 | sfx\sorceror\mage84.wav |
| 548 | Sorceror85 | sfx\sorceror\mage85.wav |
| 549 | Sorceror86 | sfx\sorceror\mage86.wav |
| 550 | Sorceror87 | sfx\sorceror\mage87.wav |
| 551 | Sorceror88 | sfx\sorceror\mage88.wav |
| 552 | Sorceror89 | sfx\sorceror\mage89.wav |
| 553 | Sorceror90 | sfx\sorceror\mage90.wav |
| 554 | Sorceror91 | sfx\sorceror\mage91.wav |
| 555 | Sorceror92 | sfx\sorceror\mage92.wav |
| 556 | Sorceror93 | sfx\sorceror\mage93.wav |
| 557 | Sorceror94 | sfx\sorceror\mage94.wav |
| 558 | Sorceror95 | sfx\sorceror\mage95.wav |
| 559 | Sorceror96 | sfx\sorceror\mage96.wav |
| 560 | Sorceror97 | sfx\sorceror\mage97.wav |
| 561 | Sorceror98 | sfx\sorceror\mage98.wav |
| 562 | Sorceror99 | sfx\sorceror\mage99.wav |
| 563 | Sorceror100 | sfx\sorceror\mage100.wav |
| 564 | Sorceror101 | sfx\sorceror\mage101.wav |
| 565 | Sorceror102 | sfx\sorceror\mage102.wav |
#### flag = `Sorcerer` — 56 rows
| # | id | path |
|---|---|---|
| 475 | Sorceror13 | sfx\sorceror\mage13.wav |
| 476 | Sorceror14 | sfx\sorceror\mage14.wav |
| 477 | Sorceror15 | sfx\sorceror\mage15.wav |
| 478 | Sorceror16 | sfx\sorceror\mage16.wav |
| 479 | Sorceror17 | sfx\sorceror\mage17.wav |
| 480 | Sorceror18 | sfx\sorceror\mage18.wav |
| 481 | Sorceror19 | sfx\sorceror\mage19.wav |
| 482 | Sorceror20 | sfx\sorceror\mage20.wav |
| 483 | Sorceror21 | sfx\sorceror\mage21.wav |
| 484 | Sorceror22 | sfx\sorceror\mage22.wav |
| 485 | Sorceror23 | sfx\sorceror\mage23.wav |
| 486 | Sorceror24 | sfx\sorceror\mage24.wav |
| 487 | Sorceror25 | sfx\sorceror\mage25.wav |
| 488 | Sorceror26 | sfx\sorceror\mage26.wav |
| 489 | Sorceror27 | sfx\sorceror\mage27.wav |
| 490 | Sorceror28 | sfx\sorceror\mage28.wav |
| 491 | Sorceror29 | sfx\sorceror\mage29.wav |
| 492 | Sorceror30 | sfx\sorceror\mage30.wav |
| 493 | Sorceror31 | sfx\sorceror\mage31.wav |
| 494 | Sorceror32 | sfx\sorceror\mage32.wav |
| 495 | Sorceror33 | sfx\sorceror\mage33.wav |
| 496 | Sorceror34 | sfx\sorceror\mage34.wav |
| 497 | Sorceror35 | sfx\sorceror\mage35.wav |
| 498 | Sorceror36 | sfx\sorceror\mage36.wav |
| 499 | Sorceror37 | sfx\sorceror\mage37.wav |
| 500 | Sorceror38 | sfx\sorceror\mage38.wav |
| 501 | Sorceror39 | sfx\sorceror\mage39.wav |
| 502 | Sorceror40 | sfx\sorceror\mage40.wav |
| 503 | Sorceror41 | sfx\sorceror\mage41.wav |
| 504 | Sorceror42 | sfx\sorceror\mage42.wav |
| 505 | Sorceror43 | sfx\sorceror\mage43.wav |
| 506 | Sorceror44 | sfx\sorceror\mage44.wav |
| 507 | Sorceror45 | sfx\sorceror\mage45.wav |
| 508 | Sorceror46 | sfx\sorceror\mage46.wav |
| 509 | Sorceror47 | sfx\sorceror\mage47.wav |
| 510 | Sorceror48 | sfx\sorceror\mage48.wav |
| 511 | Sorceror49 | sfx\sorceror\mage49.wav |
| 512 | Sorceror50 | sfx\sorceror\mage50.wav |
| 519 | Sorceror57 | sfx\sorceror\mage57.wav |
| 526 | Sorceror64 | sfx\sorceror\mage64.wav |
| 527 | Sorceror65 | sfx\sorceror\mage65.wav |
| 528 | Sorceror66 | sfx\sorceror\mage66.wav |
| 529 | Sorceror67 | sfx\sorceror\mage67.wav |
| 530 | Sorceror68 | sfx\sorceror\mage68.wav |
| 531 | Sorceror69 | sfx\sorceror\mage69.wav |
| 532 | Sorceror69b | sfx\sorceror\mage69b.wav |
| 533 | Sorceror70 | sfx\sorceror\mage70.wav |
| 534 | Sorceror71 | sfx\sorceror\mage71.wav |
| 535 | Sorceror72 | sfx\sorceror\mage72.wav |
| 536 | Sorceror73 | sfx\sorceror\mage73.wav |
| 537 | Sorceror74 | sfx\sorceror\mage74.wav |
| 538 | Sorceror75 | sfx\sorceror\mage75.wav |
| 539 | Sorceror76 | sfx\sorceror\mage76.wav |
| 540 | Sorceror77 | sfx\sorceror\mage77.wav |
| 541 | Sorceror78 | sfx\sorceror\mage78.wav |
| 542 | Sorceror79 | sfx\sorceror\mage79.wav |
#### flag = `Stream,Rogue` — 47 rows
| # | id | path |
|---|---|---|
| 566 | Rogue1 | sfx\rogue\rogue01.wav |
| 567 | Rogue2 | sfx\rogue\rogue02.wav |
| 568 | Rogue3 | sfx\rogue\rogue03.wav |
| 569 | Rogue4 | sfx\rogue\rogue04.wav |
| 570 | Rogue5 | sfx\rogue\rogue05.wav |
| 571 | Rogue6 | sfx\rogue\rogue06.wav |
| 572 | Rogue7 | sfx\rogue\rogue07.wav |
| 573 | Rogue8 | sfx\rogue\rogue08.wav |
| 574 | Rogue9 | sfx\rogue\rogue09.wav |
| 575 | Rogue10 | sfx\rogue\rogue10.wav |
| 576 | Rogue11 | sfx\rogue\rogue11.wav |
| 577 | Rogue12 | sfx\rogue\rogue12.wav |
| 616 | Rogue51 | sfx\rogue\rogue51.wav |
| 617 | Rogue52 | sfx\rogue\rogue52.wav |
| 618 | Rogue53 | sfx\rogue\rogue53.wav |
| 619 | Rogue54 | sfx\rogue\rogue54.wav |
| 620 | Rogue55 | sfx\rogue\rogue55.wav |
| 621 | Rogue56 | sfx\rogue\rogue56.wav |
| 623 | Rogue58 | sfx\rogue\rogue58.wav |
| 624 | Rogue59 | sfx\rogue\rogue59.wav |
| 625 | Rogue60 | sfx\rogue\rogue60.wav |
| 626 | Rogue61 | sfx\rogue\rogue61.wav |
| 627 | Rogue62 | sfx\rogue\rogue62.wav |
| 628 | Rogue63 | sfx\rogue\rogue63.wav |
| 646 | Rogue80 | sfx\rogue\rogue80.wav |
| 647 | Rogue81 | sfx\rogue\rogue81.wav |
| 648 | Rogue82 | sfx\rogue\rogue82.wav |
| 649 | Rogue83 | sfx\rogue\rogue83.wav |
| 650 | Rogue84 | sfx\rogue\rogue84.wav |
| 651 | Rogue85 | sfx\rogue\rogue85.wav |
| 652 | Rogue86 | sfx\rogue\rogue86.wav |
| 653 | Rogue87 | sfx\rogue\rogue87.wav |
| 654 | Rogue88 | sfx\rogue\rogue88.wav |
| 655 | Rogue89 | sfx\rogue\rogue89.wav |
| 656 | Rogue90 | sfx\rogue\rogue90.wav |
| 657 | Rogue91 | sfx\rogue\rogue91.wav |
| 658 | Rogue92 | sfx\rogue\rogue92.wav |
| 659 | Rogue93 | sfx\rogue\rogue93.wav |
| 660 | Rogue94 | sfx\rogue\rogue94.wav |
| 661 | Rogue95 | sfx\rogue\rogue95.wav |
| 662 | Rogue96 | sfx\rogue\rogue96.wav |
| 663 | Rogue97 | sfx\rogue\rogue97.wav |
| 664 | Rogue98 | sfx\rogue\rogue98.wav |
| 665 | Rogue99 | sfx\rogue\rogue99.wav |
| 666 | Rogue100 | sfx\rogue\rogue100.wav |
| 667 | Rogue101 | sfx\rogue\rogue101.wav |
| 668 | Rogue102 | sfx\rogue\rogue102.wav |
#### flag = `Rogue` — 56 rows
| # | id | path |
|---|---|---|
| 578 | Rogue13 | sfx\rogue\rogue13.wav |
| 579 | Rogue14 | sfx\rogue\rogue14.wav |
| 580 | Rogue15 | sfx\rogue\rogue15.wav |
| 581 | Rogue16 | sfx\rogue\rogue16.wav |
| 582 | Rogue17 | sfx\rogue\rogue17.wav |
| 583 | Rogue18 | sfx\rogue\rogue18.wav |
| 584 | Rogue19 | sfx\rogue\rogue19.wav |
| 585 | Rogue20 | sfx\rogue\rogue20.wav |
| 586 | Rogue21 | sfx\rogue\rogue21.wav |
| 587 | Rogue22 | sfx\rogue\rogue22.wav |
| 588 | Rogue23 | sfx\rogue\rogue23.wav |
| 589 | Rogue24 | sfx\rogue\rogue24.wav |
| 590 | Rogue25 | sfx\rogue\rogue25.wav |
| 591 | Rogue26 | sfx\rogue\rogue26.wav |
| 592 | Rogue27 | sfx\rogue\rogue27.wav |
| 593 | Rogue28 | sfx\rogue\rogue28.wav |
| 594 | Rogue29 | sfx\rogue\rogue29.wav |
| 595 | Rogue30 | sfx\rogue\rogue30.wav |
| 596 | Rogue31 | sfx\rogue\rogue31.wav |
| 597 | Rogue32 | sfx\rogue\rogue32.wav |
| 598 | Rogue33 | sfx\rogue\rogue33.wav |
| 599 | Rogue34 | sfx\rogue\rogue34.wav |
| 600 | Rogue35 | sfx\rogue\rogue35.wav |
| 601 | Rogue36 | sfx\rogue\rogue36.wav |
| 602 | Rogue37 | sfx\rogue\rogue37.wav |
| 603 | Rogue38 | sfx\rogue\rogue38.wav |
| 604 | Rogue39 | sfx\rogue\rogue39.wav |
| 605 | Rogue40 | sfx\rogue\rogue40.wav |
| 606 | Rogue41 | sfx\rogue\rogue41.wav |
| 607 | Rogue42 | sfx\rogue\rogue42.wav |
| 608 | Rogue43 | sfx\rogue\rogue43.wav |
| 609 | Rogue44 | sfx\rogue\rogue44.wav |
| 610 | Rogue45 | sfx\rogue\rogue45.wav |
| 611 | Rogue46 | sfx\rogue\rogue46.wav |
| 612 | Rogue47 | sfx\rogue\rogue47.wav |
| 613 | Rogue48 | sfx\rogue\rogue48.wav |
| 614 | Rogue49 | sfx\rogue\rogue49.wav |
| 615 | Rogue50 | sfx\rogue\rogue50.wav |
| 622 | Rogue57 | sfx\rogue\rogue57.wav |
| 629 | Rogue64 | sfx\rogue\rogue64.wav |
| 630 | Rogue65 | sfx\rogue\rogue65.wav |
| 631 | Rogue66 | sfx\rogue\rogue66.wav |
| 632 | Rogue67 | sfx\rogue\rogue67.wav |
| 633 | Rogue68 | sfx\rogue\rogue68.wav |
| 634 | Rogue69 | sfx\rogue\rogue69.wav |
| 635 | Rogue69b | sfx\rogue\rogue69b.wav |
| 636 | Rogue70 | sfx\rogue\rogue70.wav |
| 637 | Rogue71 | sfx\rogue\rogue71.wav |
| 638 | Rogue72 | sfx\rogue\rogue72.wav |
| 639 | Rogue73 | sfx\rogue\rogue73.wav |
| 640 | Rogue74 | sfx\rogue\rogue74.wav |
| 641 | Rogue75 | sfx\rogue\rogue75.wav |
| 642 | Rogue76 | sfx\rogue\rogue76.wav |
| 643 | Rogue77 | sfx\rogue\rogue77.wav |
| 644 | Rogue78 | sfx\rogue\rogue78.wav |
| 645 | Rogue79 | sfx\rogue\rogue79.wav |
#### flag = `Stream,Warrior` — 47 rows
| # | id | path |
|---|---|---|
| 669 | Warrior1 | sfx\warrior\warior01.wav |
| 670 | Warrior2 | sfx\warrior\warior02.wav |
| 671 | Warrior3 | sfx\warrior\warior03.wav |
| 672 | Warrior4 | sfx\warrior\warior04.wav |
| 673 | Warrior5 | sfx\warrior\warior05.wav |
| 674 | Warrior6 | sfx\warrior\warior06.wav |
| 675 | Warrior7 | sfx\warrior\warior07.wav |
| 676 | Warrior8 | sfx\warrior\warior08.wav |
| 677 | Warrior9 | sfx\warrior\warior09.wav |
| 678 | Warrior10 | sfx\warrior\warior10.wav |
| 679 | Warrior11 | sfx\warrior\warior11.wav |
| 680 | Warrior12 | sfx\warrior\warior12.wav |
| 725 | Warrior51 | sfx\warrior\warior51.wav |
| 726 | Warrior52 | sfx\warrior\warior52.wav |
| 727 | Warrior53 | sfx\warrior\warior53.wav |
| 728 | Warrior54 | sfx\warrior\warior54.wav |
| 729 | Warrior55 | sfx\warrior\warior55.wav |
| 730 | Warrior56 | sfx\warrior\warior56.wav |
| 732 | Warrior58 | sfx\warrior\warior58.wav |
| 733 | Warrior59 | sfx\warrior\warior59.wav |
| 734 | Warrior60 | sfx\warrior\warior60.wav |
| 735 | Warrior61 | sfx\warrior\warior61.wav |
| 736 | Warrior62 | sfx\warrior\warior62.wav |
| 737 | Warrior63 | sfx\warrior\warior63.wav |
| 755 | Warrior80 | sfx\warrior\warior80.wav |
| 756 | Warrior81 | sfx\warrior\warior81.wav |
| 757 | Warrior82 | sfx\warrior\warior82.wav |
| 758 | Warrior83 | sfx\warrior\warior83.wav |
| 759 | Warrior84 | sfx\warrior\warior84.wav |
| 760 | Warrior85 | sfx\warrior\warior85.wav |
| 761 | Warrior86 | sfx\warrior\warior86.wav |
| 762 | Warrior87 | sfx\warrior\warior87.wav |
| 763 | Warrior88 | sfx\warrior\warior88.wav |
| 764 | Warrior89 | sfx\warrior\warior89.wav |
| 765 | Warrior90 | sfx\warrior\warior90.wav |
| 766 | Warrior91 | sfx\warrior\warior91.wav |
| 767 | Warrior92 | sfx\warrior\warior92.wav |
| 768 | Warrior93 | sfx\warrior\warior93.wav |
| 769 | Warrior94 | sfx\warrior\warior94.wav |
| 770 | Warrior95 | sfx\warrior\warior95.wav |
| 771 | Warrior96b | sfx\warrior\wario96b.wav |
| 772 | Warrior97 | sfx\warrior\wario97.wav |
| 773 | Warrior98 | sfx\warrior\wario98.wav |
| 774 | Warrior99 | sfx\warrior\warior99.wav |
| 775 | Warrior100 | sfx\warrior\wario100.wav |
| 776 | Warrior101 | sfx\warrior\wario101.wav |
| 777 | Warrior102 | sfx\warrior\wario102.wav |
#### flag = `Warrior` — 62 rows
| # | id | path |
|---|---|---|
| 681 | Warrior13 | sfx\warrior\warior13.wav |
| 682 | Warrior14 | sfx\warrior\warior14.wav |
| 683 | Warrior14b | sfx\warrior\wario14b.wav |
| 684 | Warrior14c | sfx\warrior\wario14c.wav |
| 685 | Warrior15 | sfx\warrior\warior15.wav |
| 686 | Warrior15b | sfx\warrior\wario15b.wav |
| 687 | Warrior15c | sfx\warrior\wario15c.wav |
| 688 | Warrior16 | sfx\warrior\warior16.wav |
| 689 | Warrior16b | sfx\warrior\wario16b.wav |
| 690 | Warrior16c | sfx\warrior\wario16c.wav |
| 691 | Warrior17 | sfx\warrior\warior17.wav |
| 692 | Warrior18 | sfx\warrior\warior18.wav |
| 693 | Warrior19 | sfx\warrior\warior19.wav |
| 694 | Warrior20 | sfx\warrior\warior20.wav |
| 695 | Warrior21 | sfx\warrior\warior21.wav |
| 696 | Warrior22 | sfx\warrior\warior22.wav |
| 697 | Warrior23 | sfx\warrior\warior23.wav |
| 698 | Warrior24 | sfx\warrior\warior24.wav |
| 699 | Warrior25 | sfx\warrior\warior25.wav |
| 700 | Warrior26 | sfx\warrior\warior26.wav |
| 701 | Warrior27 | sfx\warrior\warior27.wav |
| 702 | Warrior28 | sfx\warrior\warior28.wav |
| 703 | Warrior29 | sfx\warrior\warior29.wav |
| 704 | Warrior30 | sfx\warrior\warior30.wav |
| 705 | Warrior31 | sfx\warrior\warior31.wav |
| 706 | Warrior32 | sfx\warrior\warior32.wav |
| 707 | Warrior33 | sfx\warrior\warior33.wav |
| 708 | Warrior34 | sfx\warrior\warior34.wav |
| 709 | Warrior35 | sfx\warrior\warior35.wav |
| 710 | Warrior36 | sfx\warrior\warior36.wav |
| 711 | Warrior37 | sfx\warrior\warior37.wav |
| 712 | Warrior38 | sfx\warrior\warior38.wav |
| 713 | Warrior39 | sfx\warrior\warior39.wav |
| 714 | Warrior40 | sfx\warrior\warior40.wav |
| 715 | Warrior41 | sfx\warrior\warior41.wav |
| 716 | Warrior42 | sfx\warrior\warior42.wav |
| 717 | Warrior43 | sfx\warrior\warior43.wav |
| 718 | Warrior44 | sfx\warrior\warior44.wav |
| 719 | Warrior45 | sfx\warrior\warior45.wav |
| 720 | Warrior46 | sfx\warrior\warior46.wav |
| 721 | Warrior47 | sfx\warrior\warior47.wav |
| 722 | Warrior48 | sfx\warrior\warior48.wav |
| 723 | Warrior49 | sfx\warrior\warior49.wav |
| 724 | Warrior50 | sfx\warrior\warior50.wav |
| 731 | Warrior57 | sfx\warrior\warior57.wav |
| 738 | Warrior64 | sfx\warrior\warior64.wav |
| 739 | Warrior65 | sfx\warrior\warior65.wav |
| 740 | Warrior66 | sfx\warrior\warior66.wav |
| 741 | Warrior67 | sfx\warrior\warior67.wav |
| 742 | Warrior68 | sfx\warrior\warior68.wav |
| 743 | Warrior69 | sfx\warrior\warior69.wav |
| 744 | Warrior69b | sfx\warrior\wario69b.wav |
| 745 | Warrior70 | sfx\warrior\warior70.wav |
| 746 | Warrior71 | sfx\warrior\warior71.wav |
| 747 | Warrior72 | sfx\warrior\warior72.wav |
| 748 | Warrior73 | sfx\warrior\warior73.wav |
| 749 | Warrior74 | sfx\warrior\warior74.wav |
| 750 | Warrior75 | sfx\warrior\warior75.wav |
| 751 | Warrior76 | sfx\warrior\warior76.wav |
| 752 | Warrior77 | sfx\warrior\warior77.wav |
| 753 | Warrior78 | sfx\warrior\warior78.wav |
| 754 | Warrior79 | sfx\warrior\warior79.wav |
Row 1 = `Walk` (`sfx\misc\walk1.wav`, Misc); row 814 = `AccessibilityInteract` (`sfx\items\titlemov.wav`).
### V9.2 sound/effects-unused.tsv rows 2-16 (15 rows, complete)
| # | id | flags | path |
|---|---|---|---|
| 1 | BrutalHit | Misc | sfx\items\bhit.wav |
| 2 | BrutalHit1 | Misc | sfx\items\bhit1.wav |
| 3 | SpellGuardianHit | Misc | sfx\misc\grdlanch.wav |
| 4 | Griswold11 | Stream | sfx\towners\bsmith11.wav |
| 5 | Griswold54 | Stream | sfx\towners\bsmith54.wav |
| 6 | Cain32 | Stream | sfx\towners\storyt32.wav |
| 7 | Ogden42 | Stream | sfx\towners\tavown42.wav |
| 8 | Warrior95b | Stream,Warrior | sfx\warrior\wario95b.wav |
| 9 | Warrior95c | Stream,Warrior | sfx\warrior\wario95c.wav |
| 10 | Warrior95d | Stream,Warrior | sfx\warrior\wario95d.wav |
| 11 | Warrior95e | Stream,Warrior | sfx\warrior\wario95e.wav |
| 12 | Warrior95f | Stream,Warrior | sfx\warrior\wario95f.wav |
| 13 | Izual | Stream | sfx\monsters\izual01.wav |
| 14 | LazarusGreetingShort | Stream | sfx\monsters\laz02.wav |
| 15 | Warlock | Stream | sfx\monsters\wlock01.wav |
### V9.3 Music tracks per level / menu
- `_music_id` (`Source/engine/sound.h:27-36`): `TMUSIC_TOWN, TMUSIC_CATHEDRAL, TMUSIC_CATACOMBS,`)
  `TMUSIC_CAVES, TMUSIC_HELL, TMUSIC_NEST, TMUSIC_CRYPT, TMUSIC_INTRO, NUM_MUSIC`.
- Level→track `GetLevelMusic(dungeon_type)` (`Source/engine/sound.cpp:310-329`): DTYPE_TOWN→TOWN,
  CATHEDRAL→CATHEDRAL, CATACOMBS→CATACOMBS, CAVES→CAVES, HELL→HELL, NEST→NEST, CRYPT→CRYPT,
  default→INTRO. Called at `diablo.cpp:3404`; applied by `LoadGameLevelStopMusic`/`StartMusic`
  (`diablo.cpp:3074-3085`, invoked `:3482`). Headless stub returns TOWN (`sound_stubs.cpp:26`).
- Menu/attract cycle `NextTrack()` (`Source/menu.cpp:36-58`): spawn ⇒ INTRO;
  INTRO→CATACOMBS→CAVES→HELL→(NEST if HF)→(CRYPT if HF)→INTRO; `RefreshMusic` (`:60-62`).
- Files: `music\dlvlf.wav|mp3` etc.; mp3-over-wav preference `sound.cpp:61-64`; track tables
  `sound.cpp:143-146`; Hellfire tracks come from the `hfmusic` archive (probe `dlvlf.wav/mp3`,
  `assets.cpp:570-586`); default rate 22050 Hz (`options.cpp:55,520`); Aulib init `sound.cpp:281`.
### V9.4 Movie list (`Source/movie.cpp` 141 lines, `movie.h`)
- API: `play_movie(pszMovie, user_can_close)` (`movie.h:16`, `movie.cpp:33` — demo mode returns
  early `:34-36`), `PlayInGameMovie(pszMovie)` (`movie.h:22`, fades to black then plays);
  SMK/CINK decode via `storm/storm_svid.h`; cursor hidden; skip on input when `userCanClose`.
- **Every movie referenced in the tree** (grep of `.smk`):
  `gendata\logo.smk` (splash, `diablo.cpp:1320`), `gendata\Hellfire.smk` (`:1326`, `menu.cpp:98`),
  `gendata\diablo1.smk` (intro, `:1328`, `menu.cpp:100`), `gendata\fprst3.smk` (in-game,
  `monster.cpp:2892` — first time the player opens the Skeleton King coffin path),
  `gendata\diabvic1.smk` / `diabvic2.smk` / `diabvic3.smk` (Diablo victory per difficulty,
  `monster.cpp:4093/4097/4100`), `gendata\diabend.smk` (Diablo death/endgame, `:4103`),
  `gendata\loopdend.smk` (Hellfire loop ending, `:4113`).
- Port stance (Volume I H11): stills + text fallback by default, SMK decode optional.
---
## V10. CLASSES & XP — 50 XP rows, classdat, and 6 × 5 per-class tables
### V10.1 Experience.tsv rows 2-51 (50 rows, complete; loader `tables/playerdat.cpp:89-127`)
Columns `Level | Experience` (cumulative). `ExperienceData` stores `levelThresholds[level-1]`;
gaps are filled with `UINT32_MAX` to keep `ValidatePlayer()` from zeroing XP (`:60-68`).
`GetNextExperienceThresholdForLevel(level)` (`:403`).
(`Δ` = XP required between this level and the next.)
#### XP rows 1–25 of 50
| # | Level | cumulative XP | Δ to next |
|---|---|---|---|
| 1 | 1 | 2000 | 2620 |
| 2 | 2 | 4620 | 3420 |
| 3 | 3 | 8040 | 4449 |
| 4 | 4 | 12489 | 5769 |
| 5 | 5 | 18258 | 7454 |
| 6 | 6 | 25712 | 9597 |
| 7 | 7 | 35309 | 12313 |
| 8 | 8 | 47622 | 15742 |
| 9 | 9 | 63364 | 20055 |
| 10 | 10 | 83419 | 25460 |
| 11 | 11 | 108879 | 32207 |
| 12 | 12 | 141086 | 40597 |
| 13 | 13 | 181683 | 49392 |
| 14 | 14 | 231075 | 82581 |
| 15 | 15 | 313656 | 110411 |
| 16 | 16 | 424067 | 147123 |
| 17 | 17 | 571190 | 195379 |
| 18 | 18 | 766569 | 258585 |
| 19 | 19 | 1025154 | 341073 |
| 20 | 20 | 1366227 | 448341 |
| 21 | 21 | 1814568 | 587327 |
| 22 | 22 | 2401895 | 766756 |
| 23 | 23 | 3168651 | 997549 |
| 24 | 24 | 4166200 | 1293323 |
| 25 | 25 | 5459523 | 1670973 |
#### XP rows 26–50 of 50
| # | Level | cumulative XP | Δ to next |
|---|---|---|---|
| 26 | 26 | 7130496 | 2151378 |
| 27 | 27 | 9281874 | 2760218 |
| 28 | 28 | 12042092 | 3528939 |
| 29 | 29 | 15571031 | 4495869 |
| 30 | 30 | 20066900 | 5707505 |
| 31 | 31 | 25774405 | 7219994 |
| 32 | 32 | 32994399 | 9100803 |
| 33 | 33 | 42095202 | 11430609 |
| 34 | 34 | 53525811 | 14305407 |
| 35 | 35 | 67831218 | 17838843 |
| 36 | 36 | 85670061 | 22164762 |
| 37 | 37 | 107834823 | 27439976 |
| 38 | 38 | 135274799 | 33847210 |
| 39 | 39 | 169122009 | 41598222 |
| 40 | 40 | 210720231 | 50937022 |
| 41 | 41 | 261657253 | 62143167 |
| 42 | 42 | 323800420 | 75535020 |
| 43 | 43 | 399335440 | 91472909 |
| 44 | 44 | 490808349 | 110362065 |
| 45 | 45 | 601170414 | 132655203 |
| 46 | 46 | 733825617 | 158854605 |
| 47 | 47 | 892680222 | 190228390 |
| 48 | 48 | 1082908612 | 227798497 |
| 49 | 49 | 1310707109 | 272788700 |
| 50 | 50 | 1583495809 | - |
- Max character level 50; `AddExperience` (`Source/player.cpp:2424-2462`):
  `clampedExp = clamp(experience * (1 + levelDelta/10.0), 0, UINT32_MAX)` (`:2434`);
  **MP anti-power-leveling**: `clampedExp = min(clampedExp, nextThreshold/20, 200*level)` (`:2440`);
  cap to the max-level threshold before adding (`:2445-2448`); level-up loop `while (!_pLevel
  is max && exp >= next) NextPlrLevel()` (`:2455-2458`); `AddPlrMonstExper(lvl, exp, pmask)`
  splits equally among masked players (`:2464-2476`).
### V10.2 classdat.tsv rows 2-7 (6 classes, complete; loader `tables/playerdat.cpp:345`)
| # | className | folderName | portrait | inv |
|---|---|---|---|---|
| 1 | Warrior | warrior | 0 | inv |
| 2 | Rogue | rogue | 1 | inv_rog |
| 3 | Sorcerer | sorcerer | 2 | inv_sor |
| 4 | Monk | monk | 2 | inv_sor |
| 5 | Bard | bard | 1 | inv_rog |
| 6 | Barbarian | barbarian | 0 | inv |
Order matches `HeroClass` (`Source/player.h`): Warrior, Rogue, Sorcerer, Monk, Bard, Barbarian.
Monk/Bard/Barbarian are Hellfire/mod-supplied (`mods/hf/txtdata/classes/classdat.tsv` rewrites 2
rows); Bard items are gated by the `testBard` option (`IsItemAvailable` tail, items.cpp).
### V10.3 per-class attributes.tsv — 27 variables × 6 classes (column sets identical, md5-verified; loader `tables/playerdat.cpp:171`)
| variable | warrior | rogue | sorcerer | monk | bard | barbarian |
|---|---|---|---|---|---|---|
| classFlags | CriticalStrike | TrapSense |  |  | DualWield | CriticalStrike,IronSkin,NaturalResistance |
| baseStr | 30 | 20 | 15 | 25 | 20 | 40 |
| baseMag | 10 | 15 | 35 | 15 | 20 | 0 |
| baseDex | 20 | 30 | 15 | 25 | 25 | 20 |
| baseVit | 25 | 20 | 20 | 20 | 20 | 25 |
| maxStr | 250 | 55 | 45 | 150 | 120 | 255 |
| maxMag | 50 | 70 | 250 | 80 | 120 | 0 |
| maxDex | 60 | 250 | 85 | 150 | 120 | 55 |
| maxVit | 100 | 80 | 80 | 80 | 100 | 150 |
| blockBonus | 30 | 20 | 10 | 25 | 25 | 30 |
| adjLife | 18 | 23 | 9 | 23 | 23 | 18 |
| adjMana | -1 | 5.5 | -2 | 5.5 | 3 | 0 |
| lvlLife | 2 | 2 | 1 | 2 | 2 | 2 |
| lvlMana | 1 | 2 | 2 | 2 | 2 | 0 |
| chrLife | 2 | 1 | 1 | 1 | 1 | 2 |
| chrMana | 1 | 1 | 2 | 1 | 1.5 | 1 |
| itmLife | 2 | 1.5 | 1 | 1.5 | 1.5 | 2.5 |
| itmMana | 1 | 1.5 | 2 | 1.5 | 1.75 | 1 |
| manaCost | 1 | 0.75 | 1 | 0.75 | 0.75 | 1 |
| itmRestoreLife | 2 | 1.5 | 1 | 1.5 | 1.5 | 2 |
| itmRestoreMana | 1 | 1.5 | 2 | 1.5 | 1.5 | 1 |
| splRestoreLife | 2 | 1.5 | 1 | 2 | 1.5 | 2 |
| splRestoreMana | 1 | 1.5 | 2 | 2 | 1.5 | 1 |
| healOtherRestoreLife | 2 | 1.5 | 1 | 3 | 1.5 | 2 |
| baseMagicToHit | 50 | 50 | 70 | 50 | 60 | 50 |
| baseMeleeToHit | 70 | 50 | 50 | 50 | 50 | 50 |
| baseRangedToHit | 60 | 70 | 50 | 50 | 60 | 50 |
Semantics: `base*` starting attributes; `max*` caps; `blockBonus` shield block %;
`adjLife/adjMana` life/mana per attribute point; `lvl*` per level (`NextPlrLevel`);
`chr*` per character point; `itm*` per item; `manaCost` class mana multiplier applied in
`GetManaAmount` (`spells.cpp:136-138`); `*Restore*` divisors for potions/spells/Heal Other;
`base*ToHit` bases (`tables/playerdat.cpp:203`, `GetClassAttributes` `:379`, used by
player.h:575-666 to-hit formulas). `classFlags`: Warrior/Barbarian `CriticalStrike`,
Rogue `TrapSense`, Bard `DualWield`, Barbarian also `IronSkin,NaturalResistance`,
Sorcerer/Monk none.
### V10.4 per-class starting_loadout.tsv — 9 variables × 6 classes (loader `tables/playerdat.cpp:209`)
| variable | warrior | rogue | sorcerer | monk | bard | barbarian |
|---|---|---|---|---|---|---|
| skill | ItemRepair | TrapDisarm | StaffRecharge | Search | Identify | Rage |
| spell | Null | Null | Firebolt | Null | Null | Null |
| spellLevel | 0 | 0 | 2 | 0 | 0 | 0 |
| item0 | IDI_WARRIOR | IDI_ROGUE | IDI_SORCERER_DIABLO | IDI_SHORTSTAFF | IDI_BARDSWORD | IDI_BARBARIAN |
| item1 | IDI_WARRSHLD | IDI_HEAL | IDI_MANA | IDI_HEAL | IDI_BARDDAGGER | IDI_WARRSHLD |
| item2 | IDI_WARRCLUB | IDI_HEAL | IDI_MANA | IDI_HEAL | IDI_HEAL | IDI_HEAL |
| item3 | IDI_HEAL | IDI_NONE | IDI_NONE | IDI_NONE | IDI_HEAL | IDI_HEAL |
| item4 | IDI_HEAL | IDI_NONE | IDI_NONE | IDI_NONE | IDI_NONE | IDI_NONE |
| gold | 100 | 100 | 100 | 100 | 100 | 100 |
Consumed by `CreatePlayer`/`InitPlayer` (`player.cpp:2300` region); `skill` = the level-1 skill
(all six start with `ItemRepair`), `item0..4` are `_item_indexes` from V1.2.
### V10.5 per-class animations.tsv — 27 variables × 6 classes (identical names; loader `tables/playerdat.cpp:251`)
| variable | warrior | rogue | sorcerer | monk | bard | barbarian |
|---|---|---|---|---|---|---|
| unarmedFrames | 16 | 18 | 20 | 12 | 18 | 16 |
| unarmedActionFrame | 9 | 10 | 12 | 7 | 10 | 9 |
| unarmedShieldFrames | 16 | 18 | 16 | 12 | 18 | 16 |
| unarmedShieldActionFrame | 9 | 10 | 9 | 7 | 10 | 9 |
| swordFrames | 16 | 18 | 16 | 16 | 18 | 16 |
| swordActionFrame | 9 | 10 | 12 | 12 | 10 | 9 |
| swordShieldFrames | 16 | 18 | 16 | 16 | 18 | 16 |
| swordShieldActionFrame | 9 | 10 | 12 | 12 | 10 | 9 |
| bowFrames | 16 | 12 | 20 | 20 | 12 | 16 |
| bowActionFrame | 11 | 7 | 16 | 14 | 11 | 11 |
| axeFrames | 20 | 22 | 24 | 23 | 22 | 20 |
| axeActionFrame | 10 | 13 | 16 | 14 | 13 | 8 |
| maceFrames | 16 | 18 | 16 | 16 | 18 | 16 |
| maceActionFrame | 9 | 10 | 12 | 12 | 10 | 8 |
| maceShieldFrames | 16 | 18 | 16 | 16 | 18 | 16 |
| maceShieldActionFrame | 9 | 10 | 12 | 12 | 10 | 8 |
| staffFrames | 16 | 16 | 16 | 13 | 16 | 16 |
| staffActionFrame | 11 | 11 | 12 | 8 | 11 | 11 |
| idleFrames | 10 | 8 | 8 | 8 | 8 | 10 |
| walkingFrames | 8 | 8 | 8 | 8 | 8 | 8 |
| blockingFrames | 2 | 4 | 6 | 3 | 4 | 2 |
| deathFrames | 20 | 20 | 20 | 20 | 20 | 20 |
| castingFrames | 20 | 16 | 12 | 18 | 16 | 20 |
| recoveryFrames | 6 | 7 | 8 | 6 | 7 | 6 |
| townIdleFrames | 20 | 20 | 20 | 20 | 20 | 20 |
| townWalkingFrames | 8 | 8 | 8 | 8 | 8 | 8 |
| castingActionFrame | 14 | 12 | 8 | 13 | 12 | 14 |
### V10.6 per-class sprites.tsv — 13 variables × 6 classes (identical names; loader `tables/playerdat.cpp:227`)
| variable | warrior | rogue | sorcerer | monk | bard | barbarian |
|---|---|---|---|---|---|---|
| classPath | warrior | rogue | sorceror | monk | rogue | warrior |
| classChar | w | r | s | m | b | c |
| trn | warrior | rogue | sorcerer | monk | bard | barbarian |
| stand | 96 | 96 | 96 | 112 | 96 | 96 |
| walk | 96 | 96 | 96 | 112 | 96 | 96 |
| attack | 128 | 128 | 128 | 130 | 128 | 128 |
| bow | 96 | 128 | 128 | 130 | 128 | 96 |
| swHit | 96 | 96 | 96 | 98 | 96 | 96 |
| block | 96 | 96 | 96 | 98 | 96 | 96 |
| lightning | 96 | 96 | 128 | 114 | 96 | 96 |
| fire | 96 | 96 | 128 | 114 | 96 | 96 |
| magic | 96 | 96 | 128 | 114 | 96 | 96 |
| death | 128 | 128 | 128 | 160 | 128 | 128 |
### V10.7 per-class sounds.tsv — 103 event rows × 6 classes (event names identical across
classes, md5-verified; loader `tables/playerdat.cpp:289`)
Schema: `speech | sfx` — `sfx` names the V9.1 row (per-class voice). Example row 1:
`ChamberOfBoneLore → Warrior1 / Rogue1 / Sorceror1 / Monk1 / …`.
Event grouping by name prefix (all 103):
| prefix | count | first | last |
|---|---|---|---|
| That | 5 | ThatDidntDoAnything | ThatWontWorkYet |
| The | 5 | TheEnchantmentIsGone | TheSanctityOfThisPlaceHasBeenFouled |
| Im | 4 | ImNotThirsty | ImSorryDidIBreakYourConcentration |
| Horazons | 3 | HorazonsSanctumLore | HorazonsJournal |
| ICantC | 3 | ICantCarryAnymore | ICantCastThatYet |
| Its | 3 | ItsTooHeavy | ItsHotDownHere |
| This | 3 | ThisMustBeWhatGriswoldWanted | ThisBladeMustBeDestroyed |
| Your | 3 | YourDeathWillBeAvenged | YourReignOfPainHasEnded |
| Argh | 2 | Argh | ArghClang |
| ICantL | 2 | ICantLiftThis | ICantLiftThisYet |
| ICantM | 2 | ICantMoveThis | ICantMoveThisYet |
| ICantO | 2 | ICantOpenThis | ICantOpenThisYet |
| IDontN | 2 | IDontNeedThat | IDontNeedToDoThat |
| INeedT | 2 | INeedToGetThisToLachdanan | INeedToGetThisToGriswold |
| Id | 2 | IdNeverUseThis | IdHaveToEquipThat |
| Ive | 2 | IveGotABadFeelingAboutThis | IveNeverBeenHereBefore |
| Maybe | 2 | MaybeItsLockedFromTheInside | MaybeTheresAnotherWay |
| No | 2 | NoWay | No |
| Not | 2 | NotAChance | NotEnoughMana |
| Oh | 2 | OhTooEasy | Oh |
| Rest | 2 | RestInPeaceMyFriend | RestWellLeoricIllFindYourSon |
| Aaaaargh | 1 | Aaaaargh | Aaaaargh |
| Ah | 1 | Ah | Ah |
| Auugh | 1 | AuughUh | AuughUh |
| Back | 1 | BackToTheGrave | BackToTheGrave |
| Chamber | 1 | ChamberOfBoneLore | ChamberOfBoneLore |
| Die | 1 | Die | Die |
| Efficio | 1 | EfficioObitusUtInimicus | EfficioObitusUtInimicus |
| Golem | 1 | GolemSpellLore | GolemSpellLore |
| Got | 1 | GotMilk | GotMilk |
| Grimspike | 1 | GrimspikeLieutenantOfBelialLore | GrimspikeLieutenantOfBelialLore |
| Halls | 1 | HallsOfTheBlindLore | HallsOfTheBlindLore |
| Heavy | 1 | HeavyBreathing | HeavyBreathing |
| Hmm | 1 | Hmm | Hmm |
| ICanAl | 1 | ICanAlreadyDoThat | ICanAlreadyDoThat |
| ICanSe | 1 | ICanSeeWhyTheyFearThisWeapon | ICanSeeWhyTheyFearThisWeapon |
| ICantD | 1 | ICantDoThat | ICantDoThat |
| ICantG | 1 | ICantGetThereFromHere | ICantGetThereFromHere |
| ICantU | 1 | ICantUseThisYet | ICantUseThisYet |
| ICould | 1 | ICouldBlowUpTheWholeVillage | ICouldBlowUpTheWholeVillage |
| IDontH | 1 | IDontHaveASpellReady | IDontHaveASpellReady |
| IDontW | 1 | IDontWantThat | IDontWantThat |
| IHaveN | 1 | IHaveNoRoom | IHaveNoRoom |
| IMustB | 1 | IMustBeGettingClose | IMustBeGettingClose |
| In | 1 | InSpirituSanctum | InSpirituSanctum |
| Just | 1 | JustWhatIWasLookingFor | JustWhatIWasLookingFor |
| Looks | 1 | LooksLikeItsRustedShut | LooksLikeItsRustedShut |
| May | 1 | MayTheSpiritOfArkaineProtectMe | MayTheSpiritOfArkaineProtectMe |
| Mm | 1 | MmHmm | MmHmm |
| Morta | 1 | MortaVespaGaieaInnuminoEvegeenJatanLuaGraton | MortaVespaGaieaInnuminoEvegeenJatanLuaGraton |
| Now | 1 | NowThatsOneBigMushroom | NowThatsOneBigMushroom |
| Oof | 1 | OofAh | OofAh |
| Phew | 1 | Phew | Phew |
| Praedictum | 1 | PraedictumOtium | PraedictumOtium |
| Return | 1 | ReturnToHeavenWarriorOfLight | ReturnToHeavenWarriorOfLight |
| Thank | 1 | ThankTheLight | ThankTheLight |
| Time | 1 | TimeToDie | TimeToDie |
| Too | 1 | TooUghHeavy | TooUghHeavy |
| Uh | 1 | UhHuh | UhHuh |
| Valor | 1 | ValorLore | ValorLore |
| Vengeance | 1 | VengeanceIsMine | VengeanceIsMine |
| Warlord | 1 | WarlordOfBloodLore | WarlordOfBloodLore |
| What | 1 | WhatWasThat | WhatWasThat |
| Where | 1 | WhereWouldIPutThis | WhereWouldIPutThis |
| Wow | 1 | Wow | Wow |
| Yeah | 1 | Yeah | Yeah |
| Yep | 1 | YepThatsACowAlright | YepThatsACowAlright |
| Yes | 1 | Yes | Yes |
| Youll | 1 | YoullLureNoMoreMenToTheirDeaths | YoullLureNoMoreMenToTheirDeaths |
Complete 103-event list (identical for every class):
| # | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| barbarian | ChamberOfBoneLore | HorazonsSanctumLore | GolemSpellLore | HorazonsCreatureOfFlameLore | MortaVespaGaieaInnuminoEvegeenJatanLuaGraton | GrimspikeLieutenantOfBelialLore |
| # | 7 | 8 | 9 | 10 | 11 | 12 |
|---|---|---|---|---|---|---|
| barbarian | HorazonsJournal | YourDeathWillBeAvenged | RestInPeaceMyFriend | ValorLore | HallsOfTheBlindLore | WarlordOfBloodLore |
| # | 13 | 14 | 15 | 16 | 17 | 18 |
|---|---|---|---|---|---|---|
| barbarian | ICantUseThisYet | ICantCarryAnymore | IHaveNoRoom | WhereWouldIPutThis | NoWay | NotAChance |
| # | 19 | 20 | 21 | 22 | 23 | 24 |
|---|---|---|---|---|---|---|
| barbarian | IdNeverUseThis | IdHaveToEquipThat | ICantMoveThis | ICantMoveThisYet | ICantOpenThis | ICantOpenThisYet |
| # | 25 | 26 | 27 | 28 | 29 | 30 |
|---|---|---|---|---|---|---|
| barbarian | ICantLiftThis | ICantLiftThisYet | ICantCastThatHere | ICantCastThatYet | ThatDidntDoAnything | ICanAlreadyDoThat |
| # | 31 | 32 | 33 | 34 | 35 | 36 |
|---|---|---|---|---|---|---|
| barbarian | IDontNeedThat | IDontNeedToDoThat | IDontWantThat | IDontHaveASpellReady | NotEnoughMana | ThatWouldKillMe |
| # | 37 | 38 | 39 | 40 | 41 | 42 |
|---|---|---|---|---|---|---|
| barbarian | ICantDoThat | No | Yes | ThatWontWork | ThatWontWorkHere | ThatWontWorkYet |
| # | 43 | 44 | 45 | 46 | 47 | 48 |
|---|---|---|---|---|---|---|
| barbarian | ICantGetThereFromHere | ItsTooHeavy | ItsTooBig | JustWhatIWasLookingFor | IveGotABadFeelingAboutThis | GotMilk |
| # | 49 | 50 | 51 | 52 | 53 | 54 |
|---|---|---|---|---|---|---|
| barbarian | ImNotThirsty | ImNoMilkmaid | ICouldBlowUpTheWholeVillage | YepThatsACowAlright | TooUghHeavy | InSpirituSanctum |
| # | 55 | 56 | 57 | 58 | 59 | 60 |
|---|---|---|---|---|---|---|
| barbarian | PraedictumOtium | EfficioObitusUtInimicus | TheEnchantmentIsGone | OhTooEasy | BackToTheGrave | TimeToDie |
| # | 61 | 62 | 63 | 64 | 65 | 66 |
|---|---|---|---|---|---|---|
| barbarian | ImNotImpressed | ImSorryDidIBreakYourConcentration | VengeanceIsMine | Die | Yeah | Ah |
| # | 67 | 68 | 69 | 70 | 71 | 72 |
|---|---|---|---|---|---|---|
| barbarian | Phew | Argh | ArghClang | Aaaaargh | OofAh | HeavyBreathing |
| # | 73 | 74 | 75 | 76 | 77 | 78 |
|---|---|---|---|---|---|---|
| barbarian | Oh | Wow | ThankTheLight | WhatWasThat | MmHmm | Hmm |
| # | 79 | 80 | 81 | 82 | 83 | 84 |
|---|---|---|---|---|---|---|
| barbarian | UhHuh | TheSpiritsOfTheDeadAreNowAvenged | TheTownIsSafeFromTheseFoulSpawn | RestWellLeoricIllFindYourSon | YourMadnessEndsHereBetrayer | YoullLureNoMoreMenToTheirDeaths |
| # | 85 | 86 | 87 | 88 | 89 | 90 |
|---|---|---|---|---|---|---|
| barbarian | ReturnToHeavenWarriorOfLight | ICanSeeWhyTheyFearThisWeapon | ThisMustBeWhatGriswoldWanted | INeedToGetThisToLachdanan | INeedToGetThisToGriswold | IveNeverBeenHereBefore |
| # | 91 | 92 | 93 | 94 | 95 | 96 |
|---|---|---|---|---|---|---|
| barbarian | MayTheSpiritOfArkaineProtectMe | ThisIsAPlaceOfGreatPower | ThisBladeMustBeDestroyed | YourReignOfPainHasEnded | NowThatsOneBigMushroom | TheSmellOfDeathSurroundsMe |
| # | 97 | 98 | 99 | 100 | 101 | 102 |
|---|---|---|---|---|---|---|
| barbarian | TheSanctityOfThisPlaceHasBeenFouled | ItsHotDownHere | IMustBeGettingClose | MaybeItsLockedFromTheInside | LooksLikeItsRustedShut | MaybeTheresAnotherWay |
| # | 103 |
|---|---|
| barbarian | AuughUh |
Consumer: `HeroSpeech` enum (`Source/sound_effect_enums.h`), `player.Say`/`SaySpecific`
(`player.cpp`, used by quests V6.2 and cows V5.4).
---
## V11. LEVELS — dungeon types, triggers, themes, set maps, unique features
### V11.1 Dungeon types / level bands (`Source/levels/gendung_defs.hpp:16-25`,
`GetLevelType(int level)` `levels/gendung.cpp:326-344`)
| level range | dungeon_type | generator | music | notes |
|---|---|---|---|---|
| 0 | DTYPE_TOWN | `CreateTown` (town.cpp; `gendung.cpp:348-350`) | TMUSIC_TOWN | warps + cows + NPCs (V5); warp gating `IsWarpOpen` `town.cpp:210-234` |
| 1–4 | DTYPE_CATHEDRAL | `CreateL5Dungeon` (`drlg_l1.cpp:1298-1312`) | TMUSIC_CATHEDRAL | Butcher/Skeleton King/Betrayer set pieces, 4 theme rooms |
| 5–8 | DTYPE_CATACOMBS | `CreateL2Dungeon` (`drlg_l2.cpp:2830`, pre-load `:2839`) | TMUSIC_CATACOMBS | RoomMin/Max `:1780-1786` |
| 9–12 | DTYPE_CAVES | `CreateL3Dungeon` (`drlg_l3.cpp:2185`) | TMUSIC_CAVES | river/lava |
| 13–16 | DTYPE_HELL | `CreateL4Dungeon` (`drlg_l4.cpp:1228`) | TMUSIC_HELL | Diablo on 16; Uber rooms |
| 17–20 | DTYPE_NEST (Hellfire) | drlg hive variant | TMUSIC_NEST | `AddOpenNest`/RedPortal, cow suits on 19/20 |
| 21–24 | DTYPE_CRYPT (Hellfire) | `CreateL5Dungeon` crypt branch (`drlg_l1.cpp:1214-1234`) | TMUSIC_CRYPT | lamps, protected minisets |
| set levels | SL_* (V11.4) | `LoadSetMap` (`setmaps.cpp:112-165`) | inherits | quest pieces |
`GetLevelType` returns `DTYPE_NONE` above 24. `lvl_entry` enum (`gendung_defs.hpp:28-37`):
`ENTRY_MAIN, ENTRY_PREV, ENTRY_SETLVL, ENTRY_RTNLVL, ENTRY_LOAD, ENTRY_WARPLVL, ENTRY_TWARPDN,
ENTRY_TWARPUP`. `_difficulty` (`:39-44`): `DIFF_NORMAL, DIFF_NIGHTMARE, DIFF_HELL`.
`giNumberOfLevels` (`loadsave.cpp:51`) = 17 (Diablo) / 24 (Hellfire). `CreateDungeon` dispatch
(`gendung.cpp:346+`): TOWN→CreateTown, CATHEDRAL+CRYPT→CreateL5Dungeon, CATACOMBS→CreateL2, etc.
### V11.2 Level entrance/exit triggers (`Source/levels/trigs.cpp`)
Piece-id lists (`trigs.cpp:31-59`): `TownDownList {715,714,718,719,720,722,723,724,725,726}`,
`TownWarp1List {1170,1171,1172,1173,1174,1175,1176,1177,1178,1180,1182,1184}`,
`TownCryptList {1330,1331,1332,1333,1334,1335,1336,1337}`, `TownHiveList {1306,1307,1308,1309}`,
`L1UpList {126,128,129,130,131,132,134,136,137,138,139}`, `L1DownList {105,106,107,108,109,111,113,114,117}`,
`L2UpList {265,266}`, `L2DownList {268,269,270,271}`, `L2TWarpUpList {557,558}`,
`L3UpList {169..182}`, plus the L3 down/town lists and L4 lists in the same block (`:48-59`).
Containers: `trigs[MAXTRIGGERS]`, `numtrigs` (`:27-28`), `TWarpFrom` (`:29`), `trigflag`.
| init fn (line) | rule (dPiece value → message) |
|---|---|
| `InitTownTriggers` (`:109-149`) | fixed cathedral stairs at **(25,29)** → `WM_DIABNEXTLVL`; then conditional warps, each `WM_DIABTOWNWARP`: Catacombs **(49,21)** `_tlvl=5`, Caves **(17,69)** `9`, Hell **(41,80)** `13`, Nest **(80,62)** `17`, Crypt **(36,24)** `21` — gated by `IsWarpOpen(...)` |
| `InitL1Triggers` (`:152-169`) | scan all tiles: `dPiece==128` → `WM_DIABPREVLVL`; `dPiece==114` → `WM_DIABNEXTLVL` |
| `InitL2Triggers` (`:172-199`) | `==266` → PREV (skipped when it is the Chamber-of-Bone quest tile); `==558` → `WM_DIABTWARPUP` `_tlvl=0`; `==270` → NEXT |
| `InitL3Triggers` (`:200-226`) | `==170` → PREV; `==167` → NEXT; `==548` → TWARPUP |
| `InitL4Triggers` (`:227-263`) | `==82` → PREV; `==421` → TWARPUP (`_tlvl=0`); `==119` → NEXT; second pass `==369` → NEXT **only when `Q_BETRAYER == QUEST_DONE`** |
| `InitHiveTriggers` (`:265-291`) | `==65` → PREV; `==62` → NEXT; `==79` → TWARPUP |
| `InitCryptTriggers` (`:292-315`) | `==183` → TWARPUP (`_tlvl=0`); `==157` → PREV; `==125` → NEXT |
| `InitPlayerTriggers` / set-level variants | `InitSKingTriggers`, `InitSChambTriggers`, `InitPWaterTriggers`, `InitVileTriggers` (declared in `trigs.h`, used by `setmaps.cpp` per case) |
### V11.3 Theme list with effects (`Source/levels/themes.cpp`, `theme_id` `tables/objdat.h:18-37`)
| # | theme_id | init fn (themes.cpp) | effect |
|---|---|---|---|
| 1 | THEME_BARREL | Theme_Barrel (`:377`) | barrel cluster in the themed room (objdat rows with theme tag) |
| 2 | THEME_SHRINE | Theme_Shrine (`:400`) | spawns one random shrine object (`OBJ_SHRINEL/R` family) |
| 3 | THEME_MONSTPIT | Theme_MonstPit (`:422`) | monster pit: extra packed monsters via `AddMonsterType`/`PlaceGroup` |
| 4 | THEME_SKELROOM | Theme_SkelRoom (`:465`) | skeleton room (bone piles + skeleton pack) |
| 5 | THEME_TREASURE | Theme_Treasure (`:513`) | treasure room: chests/gold placement |
| 6 | THEME_LIBRARY | Theme_Library (`:554`) | library: bookshelves + book spawns |
| 7 | THEME_TORTURE | Theme_Torture (`:596`) | torture rack objects |
| 8 | THEME_BLOODFOUNTAIN | Theme_BloodFountain (`:619`) | blood fountain object |
| 9 | THEME_DECAPITATED | Theme_Decap (`:633`) | decapitated/statue objects |
| 10 | THEME_PURIFYINGFOUNTAIN | Theme_PurifyingFountain (`:657`) | purifying fountain object |
| 11 | THEME_ARMORSTAND | Theme_ArmorStand (`:671`) | armor stand (armour drops) |
| 12 | THEME_GOATSHRINE | Theme_GoatShrine (`:700`) | `OBJ_GOATSHRINE` (Gharbad link) |
| 13 | THEME_CAULDRON | Theme_Cauldron (`:718`) | `OBJ_CAULDRON` object |
| 14 | THEME_MURKYFOUNTAIN | Theme_MurkyFountain (`:732`) | murky fountain object |
| 15 | THEME_TEARFOUNTAIN | Theme_TearFountain (`:746`) | fountain of tears object |
| 16 | THEME_BRNCROSS | Theme_BrnCross (`:760`) | burning cross object |
| 17 | THEME_WEAPONRACK | Theme_WeaponRack (`:785`) | weapon rack (weapon drops) |
Placement: `InitThemes()` (`themes.cpp:349-449`) — per level type a `themeCount` and a `ttval`
(room/transparency id); special rooms pick from `ThemeGood[GenerateRnd(4)]` retrying
`SpecialThemeFit(i, j)` else `GenerateRnd(17)` (`:880-886`). `HoldThemeRooms()` (`:891-916`) marks
rooms `DungeonFlag::Populated` (skips level 16 + NEST/CRYPT; non-CATHEDRAL uses
`DRLG_HoldThemeRooms`). `CreateThemeRooms()` (`:918-980`) runs the switch above and calls
`UpdateL4Trans()` for HELL. `THEME_NONE = -1` ⇒ `app_fatal` (`:977`).
### V11.4 Set map list (`Source/levels/setmaps.cpp:112-165`; enum `_setlevels` `gendung_defs.hpp:47-60`)
| _setlevels | files loaded (verbatim) | entry tile | palette/transparency | objects/triggers |
|---|---|---|---|---|
| SL_NONE | — | — | — | normal level |
| SL_SKELKING | `levels\l1data\sklkng1.dun` (pre) + `sklkng2.dun` | {83,44} | transparency `sklkngt.dun`, palette `l1data\l1_2.pal` | `AddSKingObjs()`, `InitSKingTriggers()` |
| SL_BONECHAMB | `l2data\bonecha2.dun` (pre) + `bonecha1.dun` | {70,40} | `bonechat.dun`, `l2_2.pal` | `AddSChamObjs()`, `InitSChambTriggers()` |
| SL_MAZE | hive maze | — | — | (Hellfire nest maze) |
| SL_POISONWATER | `l3data\foulwatr.dun` | {31,83} | `LoadPWaterPalette()`/`UpdatePWaterPalette()` (`quests.cpp:326-347`) | `AddFoulWaterObjs()` etc. |
| SL_VILEBETRAYER | `l1data\vile1.dun` (pre) + `vile2.dun` | {35,36} | `vile1.dun` transparency | `AddBetrayerObjs()`, `InitLzTriggers()`, return portal at (35,32) (`quests.cpp:167-174`) |
| SL_ARENA_CHURCH | `arena\church.dun` | {29,22} (visual {28,20}) | arena | `LoadArenaMap` (`setmaps.cpp:129-131`) |
| SL_ARENA_HELL | `arena\hell.dun` | {34,26} (visual {33,26}) | arena | `:132-134` |
| SL_ARENA_CIRCLE_OF_LIFE | `arena\circle_of_death.dun` | {30,26} (visual {29,26}) | arena | `:135-137` |
Sentinels: `SL_FIRST_ARENA = SL_ARENA_CHURCH`, `SL_LAST = SL_ARENA_CIRCLE_OF_LIFE`;
`GetArenaLevelType` (`setmaps.h:19-30`) → church=CATHEDRAL, hell/circle=HELL.
`LoadSetMap` also drives quest state (e.g. SL_SKELKING flips `Q_SKELKING` INIT→ACTIVE, `_qvar1=1`,
`setmaps.cpp:121-125`). Arena maps are Hellfire **arena** levels reached by clicking the town
arena tiles, not quest set-pieces.
### V11.5 Unique level features
- **Poison Water Supply** (`foulwatr.dun`, `SL_POISONWATER`): lever + fountains; `WaterDone`
  bitmask (`quests.cpp:38`) and `PureWaterPalette` (`:48`) with `StartPWaterPurify` (`:50-58`);
  auto-complete when only 4 monsters remain (`quests.cpp:176-187`); Pepin cleanses (V5.3).
- **Chamber of Bone** (`bonecha1/2.dun`, `SL_BONECHAMB`): bone demon (AI BoneDemon, BlueFlare2),
  `Q_SCHAMB`; the L2 up-stair trigger is suppressed on the quest tile (V11.2).
- **King Leoric's Tomb** (`sklkng1/2.dun`, `SL_SKELKING`): `LeoricAi`, coffin → `fprst3.smk`
  in-game movie (`monster.cpp:2892`), reward `Uniq(SKCROWN)` (V2.1 row 53).
- **Archbishop's Lazarus / Vile** (`vile1/2.dun`, `SL_VILEBETRAYER`): `OBJ_ALTBOY` altar (MP),
  RedPortals (V6.2), `LazarusAi`+`LazarusMinionAi`, L4 tile 369 exit after DONE.
- **The Butcher** (L2 set piece, `MT_CLEAVER` + `Uniq(CLEAVER)`), **Halls of the Blind**, **Anvil of
  Fury** (`IDI_ANVIL` at `SetPiece.pos + (11,11)`, `items.cpp:2455`), **Gharbad's cave** (L2) +
  goat shrine theme, **Zhar's nest** (L3) + Snotspill, **Cornerstone of the World** (HF, `Q_CORNSTN`),
  **The Defiler** (HF crypt, `Q_DEFILER`), **Na-Krul'sUber rooms** (HF: lever `objects.cpp:1828-1833`,
  `UberRow/UberCol`, `IsUberRoomOpened`, `MT_NAKRUL` death → 3 magic weapons + stream SFX
  `monster.cpp:931-940`), **Hork Demon → Theodore** (`bTheoQuest`, `monster.cpp:925-929`,
  `IDI_THEODORE`), **Little Girl/Celia portal** (`Q_GIRL`, `TOWN_GIRL` after visiting L17),
  **Wandering Trader** (`Q_TRADER`; Nest light radius 9 `missiles.cpp:1338`), **Farmer's
  Orchard/Hive** + the cow variant (V5.4b: rune bomb at (80,62) opens the Hive, `town.cpp:249-300`),
  **Ghoul Dance/graveyard/Crypt** (`Q_GRAVE`, `OpensGrave` at (35..38, 20..24)),
  **Arena set levels** (V11.4), **Nest/Crypt palettes + cycling** (`diablo.cpp:3509-3529`).
---
## V12. MISC ENUMERATIONS — cursors, dungeon enums, gold, durability, XP caps, gating, QoL
### V12.1 Cursors (`Source/cursor_defs.hpp:14-37`)
- `SelectionRegion` (`:14-19`): `None=0, Bottom=1<<0, Middle=1<<1, Top=1<<2` (flag bitset;
  consumed by objdat `selectionRegion` (V4.1) and item pick).
- `cursor_id` (`:22-36`) — all 13, verbatim: `CURSOR_NONE, CURSOR_HAND, CURSOR_IDENTIFY,`)
  `CURSOR_REPAIR, CURSOR_RECHARGE, CURSOR_DISARM, CURSOR_OIL, CURSOR_TELEKINESIS,`)
  `CURSOR_RESURRECT, CURSOR_TELEPORT, CURSOR_HEALOTHER, CURSOR_HOURGLASS, CURSOR_FIRSTITEM`.
- Item cursors: `item_cursor_id` (`items.h`), from itemdat `cursorGraphic` (V1.2, vocabulary:
  116 distinct) + unique override (`UICurs`, V1.5).
- Gold cursors: `GOLD_SMALL`, `GOLD_MEDIUM`, `GOLD_LARGE` via `GetGoldCursor(value)`
  (`items.cpp:2975-2985`).
- Cursor plumbing: `Source/cursor.cpp` (`SetICursor`/`SetYCursor`/`ClearCursor`),
  `CheckCursMove` item pick, hardware cursor `hwcursor.cpp:112`, software blit
  `scrollrt.cpp:214`, visibility rule `:236-247`.
### V12.2 Dungeon / enum master list (port must define all of these)
| enum | file:line | values |
|---|---|---|
| `dungeon_type` | levels/gendung_defs.hpp:16-25 | TOWN, CATHEDRAL, CATACOMBS, CAVES, HELL, NEST, CRYPT, `DTYPE_LAST=CRYPT`, `DTYPE_NONE=-1` |
| `lvl_entry` | :28-37 | MAIN, PREV, SETLVL, RTNLVL, LOAD, WARPLVL, TWARPDN, TWARPUP |
| `_difficulty` | :39-44 | NORMAL, NIGHTMARE, HELL, `DIFF_LAST=HELL` |
| `_setlevels` | :47-60 | NONE, SKELKING, BONECHAMB, MAZE, POISONWATER, VILEBETRAYER, ARENA_CHURCH, ARENA_HELL, ARENA_CIRCLE_OF_LIFE (+sentinels) |
| `DungeonFlag` | levels/dun_tile_data.hpp:26-60 | Populated, Blocked, PlayerBlocked, Visible, Explored, TransFill/TransVal bits, MegaText… (see Volume I §2) |
| `TileType` / `LevelCelBlock` | dun_tile.hpp:12-104 | 8-bit tile encoding |
| `theme_id` | tables/objdat.h:18-37 | 17 themes + `THEME_NONE=-1` (V11.3) |
| `_object_id` | tables/objdat.h:42+ | OBJ_* rows = objdat order (V4.1) |
| `ObjectFlag` | tables/objdat.h | Animated, Solid, Light, Breakable, Trap, MissilesPassThrough, … (atoms in V4.1) |
| `ItemType` | tables/itemdat.h:290 | Misc, Sword, Axe, Mace, Bow, Shield, LightArmor, Helm, MediumArmor, HeavyArmor, Staff, Gold, Ring, Amulet, None (`ItemTypeToString` itemdat.cpp:720-740) |
| `ItemSpecialEffect` | :386 | AffixAura,硬化 etc. — see itemdat `specialEffects` vocabulary below |
| `ItemSpecialEffectHf` | :419 | Hellfire additions (FireDamage/LightningDamage/… ) |
| `AffixItemType` | :609 | Weapon, Bow, Shield, Armor, Staff, Misc (bitmask) |
| `item_quality` | items.h:41-45 | NORMAL, MAGIC, UNIQUE |
| `_unique_items` | items.h:47+ | 91 entries incl. `UITEM_INVALID` sentinel; order == unique_itemdat rows |
| `MonsterAIID` | tables/monstdat.h:22-63 | 41 entries (V2.4) + `Invalid=-1` |
| `MonsterClass` | :66-70 | Undead, Demon, Animal |
| `MonsterAvailability` | :92-95 | Never, Retail, Always, (Hellfire band) |
| treasure bits | :86-88 | `T_MASK=0xFFF`, `T_NODROP=0x4000`, `T_UNIQ=0x8000` |
| `MFLAGS` | monsters.h | MFLAG_QUEUE/Hidden/Ghostly/Nature/PortalPassed/TargetMonster/NoMove… |
| `MissileID` | tables/spelldat.h:98-176 | 111 entries (68 Diablo, `LastDiablo=DiabloApocalypse` idx 67, Hellfire 69-108 `OrangeExplosion`, `Null=-1`) |
| `MissileGraphicsFlags` / `MissileDataFlags` / `MissileMovementDistribution` | tables/misdat.h:201+, :344-349 | (vocabularies printed in V3.1) |
| `SpellID` | spelldat.h:30-87 | Null … RuneOfStone (`LAST=RuneOfStone`, `Invalid=-1`) |
| `SpellType` | spelldat.h:20 | Skill, Spell, Invalid, Book, Scroll, Staff, Charge |
| `SpellDataFlags` | spelldat.cpp:48-55 | Fire, Lightning, Magic, Targeted, AllowedInTown |
| `quest_id` | questdat.hpp:74-99 | 24 + `Q_INVALID=-1` |
| `quest_state` | questdat.hpp:62+ | NOTAVAIL, INIT, ACTIVE, DONE + HIVE_* (V6.2) |
| `_talker_id` | towners.h:26-40 | 13 types |
| `_speech_id` | textdat.h:20-448 | `TEXT_*` ids (V8.1) |
| `SfxID` | sound_effect_enums.h:117 | row order == effects.tsv (V9.1) |
| `HeroSpeech` | sound_effect_enums.h | 103 events (V10.7) |
| `sfx_flag` | effects.cpp:132-140 | Stream, Misc, Ui, Monk, Rogue, Warrior, Sorcerer |
| `SelectionRegion` / `cursor_id` | cursor_defs.hpp:14-37 | V12.1 |
| `TalkID` | stores.h:44-70 | 24 states (V5.3) |
| `panel_button_id` | control_panel.cpp:131-141 | Charinfo, Qlog, Automap, Mainmenu, Inventory, Spellbook, Sendmsg, Friendly |
| `_music_id` | engine/sound.h:27-36 | 8 tracks + NUM_MUSIC |
| `shrine_type` | objects.cpp:65-100 | 34 values (V4.3) |
| `GameLogicStep` | game_mode.hpp | order per Volume I §1 |
itemdat `specialEffects` vocabulary (distinct atoms in use, V1.2): RandomStealLife.
itemdat `miscId` vocabulary (33 distinct): AMULET, ARENAPOT, AURIC, BOOK, EAR, ELIXDEX, ELIXMAG, ELIXSTR, ELIXVIT, FULLHEAL, FULLMANA, FULLREJUV, GR_RUNEF, GR_RUNEL, HEAL, MANA, MAPOFDOOM, NONE, NOTE, OILACC, OILBSMTH, OILOF, OILSHARP, REJUV, RING, RUNEF, RUNEL, RUNES, SCROLL, SCROLLT, SPECELIX, STAFF, UNIQUE.
### V12.3 Monster pack placement rules
See V2.4 — `PlaceGroup` (`monster.cpp:308-380`, incl. the `yp += Displacement(...).deltaX` bugfix
comment), `AddMonsterType`/`AddMonster` (`:566-603`, `:1592`), `PlaceUniqueMonst`/`PlaceUniqueMonsters`
(`:463`, `:504-610`), per-level `totalmonsters` budget, `dMonster` occupancy, `CanPlaceMonster`,
`dTransVal` room matching, leash radius 4, minion HP×2 + leader intelligence.
### V12.4 Item drop tables by monster level
`RndItemForMonsterLevel(monster.level(difficulty))` (V1.7) is the whole table: 60% no-drop gate
(`>40` ⇒ none), 37.5% gold, else `dropRate`-weighted over items with `iMinMLvl ≤ monsterLevel`,
filtered by `IsItemAvailable` (V12.8). Unique monsters use `RndUItem(monster)` with `uper=15`.
The `dropRate` column values in the file (vocabulary: 0, 1, 2)
are the weights; `1412e25ef Allow for item drop rates up to 255` is the commit that widened them.
### V12.5 Gold rules
`GOLD_SMALL_LIMIT 1000`, `GOLD_MEDIUM_LIMIT 2500`, `GOLD_MAX_LIMIT 5000` (`Source/items.h:28-30`);
`int MaxGold = GOLD_MAX_LIMIT` (`items.cpp:104`) — the per-stack cap, re-read when a stack is
split/merged (`items.cpp:2775-2790`, “strip the top off gold piles larger than MaxGold”,
player.h:973); cursor by size `GetGoldCursor` (`items.cpp:2975-2985`); `MakeGoldStack` (`items.cpp:~4676`);
gold pickup/merge `AddGoldToPlayer`; shop fit `StoreGoldFit` (`stores.cpp:1414-1425`);
monsters drop gold via `IDI_GOLD` (dropRate 1); Wirt's peek fee 50 (`stores.cpp:1706`).
### V12.6 Auto-pickup
`Source/qol/autopickup.cpp/.h`: option-gated radius pickup, gold toggle, belt top-up, town toggle
(`autoPickupInTown` default **false**, `options.cpp:862`), suppression while dialogs/panels are
open; never picks up items the player cannot use (`CanUseItem`) unless they are potions/gold.
### V12.7 Durability / repair
`iDurability` from itemdat (V1.2); jitter `ItemRndDur` (`items.cpp:1475-1478`);
break checks in `player.cpp:474-522` (weapon after hit, shield after block;
`DUR_INDESTRUCTIBLE` and 0 skip; on reaching 0 ⇒ `RemoveEquipment` + `CalcPlrInv`);
affixes `IPL_DUR`/`IPL_MODDUR` (`items.cpp:1048-1061`); repair: `SpellID::ItemRepair`
(`spells.cpp`) and Griswold `SmithRepairItem(item._iIvalue)` → `_iDurability = _iMaxDur`
(`stores.cpp:1837`); recharge: `SpellID::StaffRecharge`, `WitchRecharge`, ShrineStone (V4.3);
Hellfire oils modify durability/to-hit through the same IPL powers (V4.4).
### V12.8 Experience penalties / level caps
See V10.1: level-delta scaling, MP `min(next/20, 200*level)` cap (`player.cpp:2440`),
`_pExperience` hard clamp to the level-50 threshold (`:2445-2448`), `NextPlrLevel` stat points
from V10.3 `lvlLife/lvlMana` + 5 stat points per level (`player.cpp`).
### V12.9 Spawn (shareware) vs full — complete gate list
**Correction to a natural assumption:** `Source/restrict.cpp` (34 lines) is *not* content gating
— it only implements `ReadOnlyTest()` (`:19-28`: write+delete `Diablo1ReadOnlyTest.foo` in
PrefPath, `DirErrorDlg` on failure). All gating is `gbIsSpawn` branches:
- Quests: `InitQuests` wipes all to `QUEST_NOTAVAIL` (`quests.cpp:94-99`); `CheckQuests`/`CheckQuestKill`
  early-return (`:139-141`, `:229`); `ForceQuests` (`:203-205`).
- Monsters: `IsMonsterAvailable` rejects `availability == Retail` (`monster.cpp:3159-3167`).
- Items: `IsItemAvailable` (`items.cpp` ~`:2360-2390`) rejects **ids 62–70** (medium+heavy armour) and
  **105,107,108,110,111,113** (unavailable scrolls); `GetBookSpell` clamps the book level to 5
  (`items.cpp:620`); staff spell generation clamps level to 10 (`:1249`); `SetupAllUseful` spawn branch.
- Spells: `GetSpellBookLevel`/`GetSpellStaffLevel` skip StoneCurse/Guardian/Golem/Elemental/BloodStar/
  BoneSpirit (+Apocalypse for staff) (`spells.cpp:280-345`); `spells.cpp` spawn checks at `:74+`.
- Monsters on Diablo: `!gbIsSpawn` gates (`monster.cpp:3688,3701-3704`) + `:4087`.
- Heroes: Rogue/Sorcerer locked out with a purchase message (`selhero.cpp:294`); main-menu spawn art.
- Missiles: `menu.cpp:36-38` forces INTRO music in spawn; town warps closed by quest flags (`town.cpp:210-234`).
- Saves: `spawn_%d.sv` prefix + passwords `adslhfb1`/`lshbkfg1` (`pfile.cpp:60-67`); `pack.cpp`
  spawn strip rules (pack_test cases); `sgGameInitInfo.bSpawn` in demo/settings blobs.
- Cows: spawn resets `CowClicks` at the 4th click so the portal joke never fires (`towners.cpp:501-505`).
### V12.10 Hellfire content = the built-in `hf` mod tables (authoritative delta list)
Mechanics: V0.2. `mods/hf/txtdata/` files, measured deltas vs `assets/txtdata/` (same header,
same column order — rows are appended, some existing rows are edited):
| table | base rows | hf rows | appended | modified (common prefix) |
|---|---|---|---|---|
| monsters/monstdat.tsv | 112 | 138 | 26 | 0 |
| missiles/misdat.tsv | 68 | 108 | 40 | 0 |
| missiles/missile_sprites.tsv | 46 | 59 | 13 | 0 |
| items/item_prefixes.tsv | 83 | 86 | 3 | 9 |
| items/item_suffixes.tsv | 95 | 98 | 3 | 20 |
| items/unique_itemdat.tsv | 90 | 110 | 20 | 2 |
| spells/spelldat.tsv | 36 | 51 | 15 | 2 |
| sound/effects.tsv | 814 | 917 | 103 | 10 |
| sound/effects-unused.tsv | 15 | 1 | −14 (replaced) | 1 |
| towners/towners.tsv | 12 | 15 | 3 | 3 |
| towners/quest_dialog.tsv | 10 | 13 | 3 | 10 + different header (3 extra quest cols) |
| classes/classdat.tsv | 6 | 6 | 0 | 2 |
| classes/sorcerer/attributes.tsv | 27 | 27 | 0 | 1 |
| classes/sorcerer/starting_loadout.tsv | 9 | 9 | 0 | 3 |
- **Hellfire-only monsters (26 appended monstdat rows)**: `MT_HELLBOAR` Hellboar, `MT_STINGER` Stinger, `MT_PSYCHORB` Psychorb, `MT_ARACHNON` Arachnon, `MT_FELLTWIN` Felltwin, `MT_HORKSPWN` Hork Spawn, `MT_VENMTAIL` Venomtail, `MT_NECRMORB` Necromorb, `MT_SPIDLORD` Spider Lord, `MT_LASHWORM` Lashworm, `MT_TORCHANT` Torchant, `MT_HORKDMN` Hork Demon, `MT_DEFILER` Hell Bug, `MT_GRAVEDIG` Gravedigger, `MT_TOMBRAT` Tomb Rat, `MT_FIREBAT` Firebat, `MT_SKLWING` Skullwing, `MT_LICH` Lich, `MT_CRYPTDMN` Crypt Demon, `MT_HELLBAT` Hellbat, `MT_BONEDEMN` Bone Demon, `MT_ARCHLICH` Arch Lich, `MT_BICLOPS` Biclops, `MT_FLESTHNG` Flesh Thing, `MT_REAPER` Reaper, `MT_NAKRUL` Na-Krul.
- **Hellfire-only spells (15 appended spelldat rows)**: `Mana`, `Magi`, `Jester`, `LightningWall`, `Immolation`, `Warp`, `Reflect`, `Berserk`, `RingOfFire`, `Search`, `RuneOfFire`, `RuneOfLight`, `RuneOfNova`, `RuneOfImmolation`, `RuneOfStone`.
- **Hellfire-only missiles (40 appended misdat rows)**: indices 69–108 of `MissileID`
  (Mana, Magi, LightningWall(+Control), Immolation, SpectralArrow, FireballBow, LightningBow,
  ChargedBoltBow, HolyBoltBow, Warp, Reflect, Berserk, RingOfFire, Search, Jester,
  RuneOfFire/Light/Nova/Immolation/Stone, BigExplosion, HorkSpawn, OpenNest,
  OrangeFlare, BlueFlare, RedFlare, YellowFlare, BlueFlare2, YellowExplosion, RedExplosion,
  BlueExplosion, BlueExplosion2, OrangeExplosion) — listed verbatim in the hf table.
- **Unique items**: 20 appended HF uniques + 2 modified (e.g. rebalanced powers) → 110 rows.
- **Affixes**: 3 appended prefixes / 3 appended suffixes, 9/20 modified (Diablo value rebalances).
- **Towners**: 3 appended rows (extra cow-farm/farmer variants), 3 modified; `quest_dialog.tsv`
  gains 3 quest columns and 3 rows (the hf table has 13 rows, 21 columns).
- **Class data**: `classdat.tsv` rewrites 2 rows, `sorcerer/attributes.tsv` 1,
  `sorcerer/starting_loadout.tsv` 3 (Hellfire class tuning).
- **Sound**: 103 appended rows (HF monster/quest voices), 10 modified, and `effects-unused.tsv`
  replaced by a 1-row file.
- **Non-table HF content**: `mods/hf/nlevels/` (Nest/Crypt level data), `mods/hf/ui_art/`,
  `mods/hf/data/`, `mods/hf/lua/`, `manifest.ini`.
- **Still code-branched** (135 `gbIsHellfire` sites in root `Source/*.cpp`, Volume I §11):
  vendor caps/values (`NumSmithBasicItemsHf`, `MaxVendorValueHf`, `itemLevelAddHf`), shrine pool 34 vs
  26 (`objects.cpp:1302`), Wirt price sign flip (`stores.cpp:1003-1005`), premium slot count,
  `!gbIsHellfire && FlipCoin(4)` staff rule (`items.cpp:1239`), `AddSpectralArrow` damage fix,
  Jester/Devastation/Doppelganger combat branches (`player.cpp:524-660`), 25-vs-17 level count
  (`loadsave.cpp:51`), `bCowQuest`/`bTheoQuest` flows (V5.4), `hfmusic`/`hfvoice`/`hfmonk`/`hfbard`/
  `hfbarb` archives (`assets.cpp:570-586`), `selstart` mode switch, `IsTownerPresent` (`towners.cpp:735-750`).
- Items additionally gated by `IsItemAvailable`'s Diablo-only filter (`items.cpp`):
  `IDI_LGTFORGE`, the `IDI_OIL..IDI_GREYSUIT` range, ids 83–86 (oils), id 92 (Scroll of Search),
  ids 161–165 (runes), `IDI_SORCERER` (Short Staff of Mana), plus `IDI_BARDSWORD/IDI_BARDDAGGER`
  when the `testBard` option is on (`// Bard items are technically Hellfire-exclusive` comment).
### V12.11 QoL modules (`Source/qol/`, 16 files)
| module | behaviour (verified from the files) |
|---|---|
| `autopickup.cpp/h` | option-driven auto pickup in range; gold toggle; belt refill; `autoPickupInTown` default false (V12.6) |
| `stash.cpp/h` | 100 pages of grid storage, `GetStashSavePath` (`pfile.cpp`), saved via `loadsave.cpp:2740-2760`; search/highlight; shared stash in MP; `stash_test.cpp` vectors |
| `visual_store.cpp/h` | `StoreUi::VisualGrid` replaces the 20-line text store with an item grid (`stores.cpp:1701-1705` dispatch) |
| `itemlabels.cpp/h` | ground item labels coloured by quality (`_iMagical`), item-name + stat deltas |
| `floatingnumbers.cpp/h` | damage/heal/XP numbers over targets; pure cosmetics, no logic |
| `monhealthbar.cpp/h` | monster health bars above actives |
| `xpbar.cpp/h` | XP bar segmented from `Experience.tsv` (V10.1), option `experienceBar` (`player.cpp:2449`) |
| `chatlog.cpp/h` | MP chat history + timestamps; fed by `control_chat.cpp`, `CMD_STRING` 80-char cap |
| `chat_log` helpers + `control_chat_commands.cpp` | slash commands incl. cow/theo quest gates (`:159-161`) |
### V12.12 Transcription checklist for the Go data packages (row-count assertions)
Every count below is the number the Go table must equal when loading the base files; the
HF-mod variant is the second number (V12.10). A `go test` per table asserting the count plus
first/last-row field equality closes this volume's acceptance criteria.
| Go table | base rows | +hf mod rows |
|---|---|---|
| itemdat | 168 | 168 |
| item_prefixes | 83 | 86 |
| item_suffixes | 95 | 98 |
| unique_itemdat | 90 | 110 |
| monstdat | 112 | 138 |
| unique_monstdat | 100 | 100 |
| misdat | 68 | 108 |
| missile_sprites | 46 | 59 |
| objdat | 109 | 109 |
| questdat | 24 | 24 |
| spelldat (+Null row) | 37 | 52 |
| textdat | 338 | 338 |
| towners | 12 | 15 |
| quest_dialog | 10 | 13 |
| effects | 814 | 917 |
| Experience | 50 | 50 |
| classdat | 6 | 6 |
| attributes (per class) | 27 | 27 |
| sounds (per class) | 103 | 103 |
| animations (per class) | 27 | 27 |
| sprites (per class) | 13 | 13 |
| starting_loadout (per class) | 9 | 9 |
*Uniqueness invariant:* the affix+item+unique totals per the user's brief are 168 + 83 + 95 + 90 =
**436 item-domain rows**, 112 + 100 = **212 monster rows**, 68 + 46 = **114 missile rows**,
**109 object rows**, **24 quest rows**, **36 spell rows**, **338+139 text rows**,
**12+10 town rows**, **814+15 audio rows**, **50 XP rows**, **6 + 6×(27+27+103+13+9) = 1080 class rows** (ERRATUM devilGo-hy8e: the old text said 1122; the parenthesised sum is 179, 179×6 = 1074, +6 = 1080 — disk-verified by counting the 30 per-class files).
