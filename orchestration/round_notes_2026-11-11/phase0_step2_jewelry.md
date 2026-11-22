# Phase 0 gear, step 2: jewelry (trinkets, rings, necks, cloaks), 2026-11-22

Brief: `briefs/phase0_gear_step2_1122.md`. No search was run, and no preset was written.

## Candidates

`RANK_ONLY` needs a ranked list to start from, and only Enhancement has one. So for every spec the candidates
come from a stat-only ranking of the dry run's pool: each option's stats (random suffix included) times rough
weights per spec (`orchestration/stat_weights.md`'s main stats, 1% hit or crit at 20 to 30 AP or 10 to 12 SP,
the spec's schools). The top 12 per slot per spec make 147 items. Every pool trinket (the pool has about 100)
and every jewelry item with an effect or hidden stats join them, since a trinket's use can lift it past stat
order. That makes 248 items in all.

Most of each spec group's top 5:

| Slot | Specs | Candidates |
|---|---|---|
| trinket | Warrior, Ret, Enh | Behemoth Bones, Counterattack Lodestone, Blackhand's Breadth, Lucky Shovelnose Tuft |
| trinket | Rogue, Feral | Behemoth Bones, Counterattack Lodestone, Blackhand's Breadth, Thaelemaches' Talisman |
| trinket | Hunter | Behemoth Bones, Counterattack Lodestone, Blackhand's Breadth, Thaelemaches' Talisman |
| trinket | Casters | Weakness Analyzer, Briarwood Reed, Zarla's Snack Pouch, Eye of the Beast |
| finger | Warrior, Ret, Enh | Legionnaire's Band, Opal Ring, Prismatic Band, Granite Ring (of the Tiger) |
| finger | Rogue, Feral | Opal Ring, Prismatic Band, Warden's Signet of Ferocity, Band of the Cultist |
| finger | Hunter | Opal Ring, Prismatic Band, Granite Ring (of Agility), Eagle's Head Signet |
| finger | Casters | Maiden's Circle, Band of the Damned, Jereman's Bauble, Warden's Signet of Brilliance |
| neck | Warrior, Ret, Enh | Amulet of the Darkmoon, Prismatic Pendant, Granite Necklace, Vermilion Necklace |
| neck | Rogue, Feral | Amulet of the Darkmoon, Prismatic Pendant, Granite Necklace, Vermilion Necklace |
| neck | Hunter | Prismatic Pendant, Granite Necklace, Amulet of the Darkmoon, Maelstrom's Tendril |
| neck | Casters | Orb of the Darkmoon, Lady Maye's Pendant, Archlight Talisman, Demon Worshipper's Necklace |
| back | Warrior, Ret, Enh | Cape of the Black Baron, Adventurer's Cape, Master's Cloak, Masterwork Cape |
| back | Rogue, Feral | Cape of the Black Baron, Blackveil Cape, Shifting Cloak, Shadow Prowler's Cloak |
| back | Hunter | Cape of the Black Baron, Deathguard's Cloak, Adventurer's Cape, Shifting Cloak |
| back | Casters | Crystalline Threaded Cape (out), Adventurer's Cape, Archivist Cape, Eternal Cloak |

The on-use trinkets with modelled effects (Everlook Pathcarver, the Forever burst trinkets, Hand of Justice and
the like) rank low on stats alone, so they're in only as part of the whole trinket pool.

## Effects modelled

From client 1.60.1.71037 (wago ItemEffect, SpellEffect, SpellMisc, SpellCategories), in `sim/common/item_effects.go`,
Forever only:

- Linken's Boomerang (11905): 113 to 187 Physical, 3 min cooldown, 10s on the burst trinket category (1141). No
  defense type, so it can't miss or crit. The disarm and knockdown do nothing to a boss. Test:
  `TestForeverLinkensBoomerang` (`sim/shaman/enhancement/forever_trinkets_test.go`).
- Cloak of Fire (14134): 25 Fire to every enemy near you every 5s for 15s, 2 min cooldown, no shared category,
  can't miss or crit. Test: `TestForeverCloakOfFire` (same file).

The dry run no longer lists either as unmodelled.

## Effects left unmodelled

A comment above Linken's Boomerang lists them:

- Cannonball Runner (13382) and Arcanite Dragonling (16022) are summons. The client has the cannon's spell
  (Cannon Fire, 85 to 139 Holy to nearby enemies) but not how often it fires, and nothing of the dragonling's.
- Six Demon Bag (7734) picks its spell on the server. The client has only a dummy effect.
- Ramstein's Lightning Bolts (13515) is a 2s stun in Forever (Classic's 200 Nature is gone).
- The other 30-odd are defensive or utility: absorbs (Arena Grand Master, Defiler's Talisman, Mark of
  Resolution), heals (Lifestone, Defender of the Timbermaw), reflectors, resistances, dispels, teleports
  (the insignias, Dimensional Ripper), speed, roots, fishing, Earthen Sigil's mana.

## Hidden stats

Three jewelry items have Classic stats only. None of the sources has Forever's: not in ItemSparse 1.60.1.71037,
`forever_seen_in_game.json`, `forever_revealed_in_game.json` or foreverchanges.pro (no item pages), and
`forever_client_cache.json` marks them invalid.

- Weighted Cloak (20693) and Crystalline Threaded Cape (20697) are Scourge Invasion rewards, and Forever has no
  invasion, so they're out and their stats don't matter.
- Ragnaros Core (17982) has no known source and only fire resistance. Doubtful.

## Availability

`phase0_content.json` got 201 entries (140 trinkets and necks, 61 rings and cloaks), including the 45 candidates
that had only the 2026-11-12 "left unlisted" mark. Every candidate is now listed with a source. Of the new entries:
137 open, 36 doubtful, 27 out, 1 infeasible (Frostwolf Insignia Rank 6, Exalted).

`phase0_enchants.json`: the only cloak enchants with a DPS stat are Minor, Lesser and plain Agility, and all
three were already listed. Nothing added.

## Check

Enhancement's `RANK_ONLY=1 SLOTS=trinket,finger,neck,back` finishes with no unlisted item (exit 0, 5819 sim runs on
4 cores). Picks are unchanged: Blackhand's Breadth and Chained Essence of Eranikus, Watcher's Signet of Brilliance and
Channeler's Ring, Beads of Ogre Might, Arcanoweave Cloak. Both new effects get ranked: Cloak of Fire's use is worth
+1.1 DPS on the boss and +4.4 against 4 mobs (back rank 26 vs 4 mobs), Linken's Boomerang's +1.4 on the boss. Neither
makes the list. `EFFECTS_ONLY` writes effects only for a list's picks, so they can't show there until a pick wears them.

The other 22 specs have no list yet, so `RANK_ONLY` can't run for them (it needs `FROM`). Their candidates are all
listed, so step 3's unlisted check should find none in these slots. One bug turned up on the way:
the first run panicked on Cloak of Fire's ticks (no defense type on a crit check), fixed in 07c1ac4.

To run the script here I needed a `jj` shim (the container has only git) and `protoc` for the generated protos.

## Needs a user call

- Arcanite Dragonling: open (Engineering, schematic a 0.45% drop) but it may cost 300g or more. Its summon isn't
  modelled anyway.
- Alterac Valley: the Frostwolf Insignias (ranks 1 to 5) are doubtful until we know AV is open. They're utility
  only.
- Blackhand's Breadth and Eye of the Beast: Wowhead's Forever data says Blackrock Spire quest reward, but our data
  doesn't name the quest. I kept them open (Blackhand's Breadth is Enhancement's pick).
- Warsong Outriders rings (Legionnaire's Band and others): Honored per our data, for both item levels. Worth a look
  in game.
- 33 doubtful trinkets and necks have no source in our data, mostly Forever-only items from Hyjal, Blackmaw Hold
  or unknown zones (Bolt's Charged Eyeball, Scalding Flamehoof among them).
