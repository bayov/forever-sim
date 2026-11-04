# Healer mechanics: pass 2 (2026-11-04 UTC, master 0facd9428, client 1.60.1.70790)

Due because 54 watched files changed since pass 1 (a1b1163d2). Pass 1's full per-spell tables are in
`ref/healer_mechanics/verdicts.md` (updated in place this pass).

## Get current

| Source | Seen | Healer effect |
|---|---|---|
| wago.tools builds | 70790 newer than 70755 | Riptide only (below); no healing talent's TraitNodeEntry or curve changed (the 11 MaxRanks and ~40 curve changes are rogue, warrior, hunter, Lightning Overload, Flurry, Rend and Tear, Shadow and Flame, Mental Dexterity) |
| Dev notes thread 2360696 | post 9, updated 2026-11-04T02:12 (no newer; read from the mechanics pass's fetch at 18:52, our own fetch was rate-limited) | none new |
| Blue posts | newest "Forever Launch Time" 2026-11-04T18:13 | no mechanics |
| update_ledger.md | no 70790 section yet | n/a |

### 70790 Riptide (tier 1)

| Row | 70755 | 70790 | Sim (Forever) | Verdict |
|---|---|---|---|---|
| Effect 0 (1072707, 1236793, 1236796) | HEAL 480/612/841, coef .214 | DUMMY, base 0 | no instant heal | match |
| Cooldown (84834, 94614, 94615) | category 6000 | 12000 | 12 s shared timer (T1 5P -1 s) | match |
| Mana (304804, 310863, 310864) | 245/300/385 | 150/180/230 | 147/180/231 (0.6 x 70755, rounded) | off by 3/0/1 (finding 11) |
| HoT ticks, Chain Heal bonus | 89/115/161, coef .1; 25% | unchanged | same | match |

The notes-only Riptide (shaman call 1) is now client data, so the sim's notes switch can read the client instead.

## What changed on master (watched paths), healer parts read in full

| Commit | Change | Verdict |
|---|---|---|
| b2c16f671 | raid damage share x N < 1: hit one player that often, nobody otherwise | fixes finding 1; mean damage now matches the fit |
| bc8e5473b | Swiftmend keeps its HoT under Forever (still Blooms on expiry); druid doc tables | fixes findings 2, 9 (druid) |
| 4648ff419 | Ancestral Fortitude via PseudoStats.ArmorMultiplier (Unit.Armor() reads it at hit time) | fixes finding 3 |
| d64804b9f | no Benediction on Divine Favor or Oath of Light; Libram of Holy Alacrity call answered | fixes findings 4, 9 (paladin) |
| dda14db65 | Vestments of Conviction 5P also on Prayer of Mending's cast spell (the jump heal has no CD timer) | fixes finding 5 |
| c938028ee | Restoration synthetic set on head and chest (2P +10 Spirit live), 4P 17 healing + 9 SP, 5P Riptide -1 s | fixes finding 6; values match 1300960/1301093/1301713 |
| 0aed93989 | HealingCritChance adds the school crit bonus; current health clamps when a Health bonus goes negative | fixes findings 7, 8 |
| 859bdef1a | Meditation adds to SpiritRegenRateCasting instead of overwriting it (set 4P +.15 now survives) | correct (client a134 is additive) |
| 3f8d901b5 / penance.go | Litany of Light reads the Penance heal bolt's SpellCode | correct (keeps the channel out of the Penance damage/cost masks) |
| health.go PayHealthCost, mana.go mana-gain threat unmultiplied | tank-side; no healing effect | n/a here (tank auditors) |
| shaman damage (hotfix 113380), Warden, bear, Protection paladin | not healing | n/a |
| tools/healbench | tank role keeps idle fillers (Prayer of Mending has targets), tanks face the boss | fine |

## Engine checks re-done

Raid damage share (fixed), heal crit chance (school bonus now on direct heals, same units as SpellCritChance), health
clamp on max health loss (processDynamicBonus), Meditation and other casting-regen sources all additive. Nothing else
in the engine moved.

## Tests

`go test -tags=with_db` on sim/shaman/enhancement, sim/paladin/holy, sim/priest/healing, sim/druid, sim/core (healing
filters): all pass. New coverage: shaman rank table, Riptide tick crits, Ancestral Fortitude, T1 set; paladin lower
ranks, Holy Shock/Light's Vigil, Benediction, Holy Power; druid ranks and Swiftmend's kept HoT; priest suites.
Still untested: the three engine fixes (share < 1, health clamp, school crit on direct heals), heal assignments,
Lowest Health by percent, cancelled heals, two absorbs on one unit (finding 10, updated).

## Findings

- New: healer_mechanics-20261104-11 (low) Riptide mana 147/180/231 vs client 70790's 150/180/230.
- Resolved: 1-9 (commits above).
- Open: 10 (tests), updated.
