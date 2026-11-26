# Phase 0 healer: restoration druid (2026-11-26)

Branch phase0-tankheal-druid_restoration, on origin/phase0-tankheal merged at 88b7eb4eb (proto regenerated).
List in `ui/restoration_druid/bis/level60_phase0_druid_restoration.bis.json`, gear set in
`ui/restoration_druid/gear_sets/level60_phase0_druid_restoration.gear.json`, both in
`tools/healbench/req/druid_restoration.json` and `druid_restoration_raid.json`. SkyborneWindshaper, Horde, talents
05302001--5253005153113051 (kept).

## How the list was made

- Full search on the pre-merge code: 127.0 to 128.7 at 5000 iterations (neck, legs, ring swaps, Idol of
  Rejuvenation confirmed in the ranking).
- Following the Disc tip (few swaps), a full RANK_ONLY from the ranked tops (Emberfury, Cloak of Fire, Tombstone
  Breastplate, Blackhand's Breadth): it settled at 127.1, below the first list, so the first list was kept.
- After the merge the same list scored 126.85 (-2.1, sim changes). RANK_ONLY on weapons and trinkets:
  Hammer of Divine Might (2H) for Hand of Edward the Odd and Obsidian Bauble, +0.5 confirmed at 5000.
- Since the score moved, RANK_ONLY over all slots: no swap confirmed. Final list 126.55 (127.0 at the 5000-iteration
  final measure). Biggest unconfirmed: trinket Barov Peasant Caller +0.1, staff Glowing Brightwood +0.1.
- phase0_check.py: no flags for druid_restoration.
- Container restarted twice during the first search (idle reclaim), both runs redone.

## Picks

| Slot | Pick | Source | Flag |
|---|---|---|---|
| head | Feralheart Cowl | Dungeon Set 2 copy quest, giver not in data | doubtful |
| neck | Medallion of the Dawn | Argent Dawn Insignia quest, Naxx era | doubtful |
| shoulder | Feralheart Spaulders | Dungeon Set 2 copy quest | |
| back | Cloak of the Cosmos | open | |
| chest | Feralheart Vest | Dungeon Set 2 copy quest | doubtful |
| wrist | Feralheart Wraps | Dungeon Set 2 copy quest | |
| hands | Feralheart Hands | Dungeon Set 2 copy quest | |
| waist | Belt of the Archmage | Tailoring 295, pattern on AH | |
| legs | Elderwild Pants | Leatherworking 300, pattern has no known source | doubtful |
| feet | Feralheart Galoshes | Dungeon Set 2 copy quest | |
| finger | Watcher's Signet of Brilliance | Watchers Revered, no rep source found | doubtful |
| finger | Don Mauricio's Band of Domination | open | |
| trinket | Chained Essence of Eranikus | Quest, Sunken Temple | |
| trinket | Eye of the Beast | Quest, Blackrock Spire | |
| two-hand | Hammer of Divine Might | Kormok, summon unconfirmed in Forever | |
| idol | Idol of Rejuvenation | Mor Grayhoof, summon unconfirmed | |

Enchants: head Lesser Arcanum of Voracity, neck Healing Power, back Fire Resistance, chest, legs and feet Mystic
Rugged Armor Kit, wrist Healing Power, hands Mystic Thick Armor Kit, weapon Mighty Spirit (doubtful).

## Tags added

- phase0_content.json: Bleak Howler Armguards, Stormshroud Shoulders, Dreadmist Sandals, Ghoul Skin Leggings,
  Battle Healer's Cloak (both, Horde), Idol of Rejuvenation, Elunite Cord, Spiderchitin Wrap, Tome of the Darkspear
  Prophecy (Horde), Stormcloth Boots, Fallbrush Handgrips, Ghostloom Leggings, Wildheart Spaulders, Furbolg Medicine
  Pouch, Sacred Bangle, Caretaker's Cord (open). Grovekeeper legs, Sedge Boots, Idol of Swiftness (doubtful).
  Grovekeeper Grips (out, raid essence).
- phase0_enchants.json: Enchant Cloak - Minor Resistance.

## Set trap

Feralheart is the start and the list keeps its 6pc. Wildheart 8pc in the same slots, same enchants, healbench
1000 iterations: tank 1.182 against 1.188, raid 1.338 against 1.340. No trap. The ranking also never confirmed a
Wildheart piece (Wildheart Spaulders -0.15).

## Class check

Start row and final list: no class restriction excludes Druid. Feralheart and Caretaker's Cord are Druid only,
Eye of the Beast allows Druid, the rest have no allowlist (db.json).

## Rotation and mana

Talents and rotation not changed. No druid goes OOM on any boss in either role. Tightest: Chromaggus tank, 827
mana left (projected OOM 126s), synthetic 1882. Vaelastrasz projects OOM at 153s (tank) but ends with about 5100.

## Faction row

No pick is a faction reward (Battle Healer's Cloak and the Darkspear items are top 5 only), so no Alliance row.

## Score, carry in percent, 7 bosses, 1000 iterations

| Boss | Tank P0 | Tank syn | Raid P0 | Raid syn |
|---|---|---|---|---|
| magmadar | 194.0 | 204.9 | 173.4 | 182.4 |
| golemagg | 178.9 | 190.2 | 243.4 | 249.4 |
| onyxia | 133.7 | 138.7 | 94.5 | 95.3 |
| broodlord | 86.6 | 88.5 | 89.4 | 89.3 |
| vaelastrasz | 87.9 | 93.1 | 102.9 | 96.3 |
| chromaggus | 90.4 | 101.1 | 110.8 | 109.5 |
| ragnaros | 60.2 | 62.3 | 123.6 | 123.8 |
| mean | 118.8 | 125.5 | 134.0 | 135.1 |

Gap Phase 0 minus synthetic: tank -6.7, raid -1.1.

## For the user

- Doubtful picks: Feralheart Cowl and Vest, Medallion of the Dawn, Watcher's Signet, Elderwild Pants, Mighty Spirit.
- Hammer of Divine Might and Idol of Rejuvenation need the Dungeon Set 2 summons (Kormok, Mor Grayhoof) in Forever.
- No public resto druid pre-raid list was reachable for the hand check.
