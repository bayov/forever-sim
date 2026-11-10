# Client updates: pass 12 (2026-11-10, master da272241e5)

## Sources

| Source | Result |
|---|---|
| wago.tools (`wago.py latest --probe 300`) | 1.60.1.70888 (was 70862) |
| `table_summary.py 70862 70888` | Spell 2, SpellEffect 14 changed + 2 added, SpellMisc 4, SpellCooldowns 2 (+1/-1), SpellClassOptions +3, ItemSparse 16 changed + 1 added, ItemEffect 4; ChrRaces, SpellPower, SpellAuraOptions, SpellItemEnchantment, trait trees identical; no player damage spell row changed |
| Blizzard hotfixes thread 2403040 | post 5 (2026-11-10 04:23 UTC): the November 9 hotfixes; 70888 carries their client side |
| foreverchanges.pro | 70888 "holds the server's hotfixes of November 7 to 9" + Fear Ward in Shadowform; November 10 class tuning not in it. Server hotfix to 70888 (16:51 UTC 11-10): Arcane Shot and Summon Hawk cost 11% base mana, Great Stamina gives pet health, Nature's Swiftness on Wild Growth |
| Dev notes thread 2360696 | not re-fetched (forum rate limit hit once this pass; the hotfix thread covers the build) |

## DPS-relevant changes and sim status

| Row | Change | Sim status |
|---|---|---|
| MAGE-HF-02 | Arcane Concentration chance split across the Arcane Missiles channel (11-09 hotfix) | **wrong**: full 10% per missile. Scratch build: mage__arcane -0.73%. Finding client_update-20261110-1 |
| MAGE-HF-03 | Flamestrike ticks no longer use Clearcasting | landed |
| SHA-HF-01 | Lightning Overload deals the correct Lightning Bolt damage | landed (probably; the sim already uses half the bolt) |
| SHR-U-21/22 | ilvl 65 trinkets 272437-9 stat cuts and longer use cooldowns; Flamebreak Talisman use 50 -> 33, 3 -> 5 min | stale db / use effects not simulated; no request wears them. Finding client_update-20261110-2 (low) |
| HUN-HF-02/03 | Arcane Shot, Summon Hawk cost 11% base mana (about 189 at 60 vs 190) | wrong by about 1 mana; below noise, no finding |
| DRU-U-18 | Natural Instinct mask takes Lacerate's bleed and Pounce Bleed | n/a (no DPS request uses either) |

Measured (180 s, 20k iterations, `dps_beforeafter.sh`, master CLI vs master + chance/5 on missile ticks):

| Request | Seed | Before | After | Delta |
|---|---|---|---|---|
| mage__arcane | 1 | 969.28 | 962.07 | -7.21 (-0.74%) |
| mage__arcane | 1000003 | 969.26 | 962.25 | -7.01 (-0.72%) |
| mage__frostfire | 1 | 1064.26 | 1064.26 | 0 |

## Ledger

Added `## Build 1.60.1.70888 (first seen 2026-11-10)` to `orchestration/update_ledger.md` with 10 rows (MAGE-HF-02/03, SHA-HF-01, DRU-U-18, SHR-U-21/22, HUN-HF-02/03/04, DRU-HF-01) and updated the Summary counts (773 -> 783).

For the orchestrator: set `.agents/skills/dps-report/state.json` `foreverClientBuild` to 1.60.1.70888. The class tuning Blizzard announced for November 10 (pets, Wrack versus the DoTs, Enhancement Strength) is not in 70888; the next build or hotfix should carry it.
