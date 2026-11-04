# Mechanics pass 65 (2026-11-04, master 6b480b060)

Quiet: no new findings.

## Range

fb584eafa..6b480b060 (68 watched files). Non-test commits under `sim/**/*.go`:

| Commit | What | Mechanics check |
|---|---|---|
| 8f304b240, b57f5cc0a | Downranking penalty moved into `sim/core/downrank.go` and applied to every ranked non-physical spell (user ruling 2026-11-04); `DownrankLevel` for rank-less sub-spells (Arcane Missiles / Penance ticks, Insect Swarm, totems, imbue and seal procs) | Formula 1 - 0.05 x (level - 17 - learn level), clamped. Settled by the ruling; moved from open items to Settled in verdicts. Level 60 presets cast ranks learned at 43+, so factor 1 at 60; report hashes unchanged per commit. Every file with `Rank:` sets `RequiredLevel` or `DownrankLevel` except Blessing of Sanctuary (no coefficient). |
| fbddbcba5 (buffs.go) | Power Infusion adds +20% healing | Client 70755 10060: aura 136 bp 20 and aura 79 bp 20. Verified, row added. |
| absorb.go, spell_result.go, unit.go | Absorb pipeline for shields on friendly units; helpful HoT snapshot uses `CasterHealingMultiplier` | Runs only on units with shields (healing sims); enemy targets carry none, so DPS presets are untouched. |
| fa66465bd, 72b236b48, fa128f029, 620ff0fdb | Holy paladin and healing priest (heals, healing talents, Binding Heal out) | Class scope (healing); no shared mechanic beyond the rows above. |
| 7cd399e88, 316a600ac, f6a98acd3 | Warrior / rogue / paladin server hotfix pushes 113539 / 113483 / 113542 | Class data, ledger rows by the client update auditor. |
| 280fd26e2 | Chain Lightning 0.714 every rank (standards 39) | Class, ruling. |

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70755 still newest (gate). |
| Hotfixes | wago.tools 70755 pushes since 113574: 113526 (Creature), 113528 (GlobalStrings), 113595 (Great Stamina 5041/5042 rank text only), 113600, 113604, 113605 (items). No spell mechanics. |
| Dev notes thread | 9 posts, no edits after post 9's 2026-11-04T02:12:27Z. |
| Blue tracker | Kaivax, Aggrend, Gnedi, Linxy, Brenoch: nothing after 11-03 21:41 (Linxy). |
| Daily update | update_ledger.md: rogue hotfix 113483 and paladin 113542 rows (class). No forever-update merges. |

## Conclusions

Every verdict holds on 70755. Open findings unchanged. Downranking scope is now a user ruling, not an open item.
