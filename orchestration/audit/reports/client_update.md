# client_update pass 18 (2026-11-26, master bfbee78fcc)

Build: **1.60.1.71139** (wago.tools `latest --probe 300`), baseline 71091. Note for `.agents/skills/dps-report/state.json` `foreverClientBuild`: 1.60.1.71139.

## What was checked

- `table_summary.py 71091 71139`: SpellEffect +3/~23, SpellMisc ~67, SpellInterrupts -15/~3, SpellAuraOptions -7, SpellClassOptions -7, ItemSparse +2/~82 (flags and bonding only), ItemEffect ~6. Trait*, SpellPower, SpellCooldowns, SpellItemEnchantment, ChrRaces identical.
- `db2diff.py 71091 71139` (346 lines), `items_diff.py` (no request item, no stat change), `enchant_diff.py` (nothing).
- Sources: Blizzard hotfix thread t/2403040 posts 14-16 (JSON read fine this time), foreverchanges.pro (71139 current; hotfixes 11-24 and 11-25).
- Each DPS-relevant line grepped in `sim/` and its value read. Open findings in `findings/mechanics.md` checked so nothing is filed twice.

## What 71139 is

The client copy of the November 23 world-buff fix (zone-change interrupts removed), the Darkmoon Faire fix, the November 24 realm hotfixes (Fist of Shahram on the wielder) and the temporary bind-on-pickup poisons. Blizzard's November 25 hotfixes (Balance druid rework, pet crit, hawk crit, Judgement refresh, Exorcism crit) are server side; the client's talent tables are unchanged.

## DPS-relevant rows (ledger section "Build 1.60.1.71139", 13 rows)

| Row | Change | Status |
|---|---|---|
| DRU-HF-04/05/06 | Balance: Starlight Wrath (-0.5 s Starfire), Improved Starfire 2 pts, Eclipse +20% Wrath crit, Moonkin Form +10% Arcane/Nature, Moonfury removed | missing; already filed as mechanics-20261125-1 (high) |
| HUN-HF-11 | Pets crit on basic attacks | landed |
| HUN-HF-12 | Hawks +~5% crit | wrong; new client_update-20261126-1 (medium) |
| PAL-HF-10 | Judgement refresh under 20 s | landed |
| PAL-HF-11 | Exorcism can crit | landed |
| WL-HF-07/08 | Demons' melee crit was 0%; Succubus 5% too much crit | wrong / ambiguous; new client_update-20261126-2 (medium) |
| SHA-HF-11 | Chain Lightning, Clearcasting with 5 Maelstrom stacks | minor, not checked |
| SHR-HF-10 | Blackblade of Shahram hastes the wielder | landed |
| SHR-HF-11 | Sayge's fortunes: the doubling was two 10% auras; 71139 keeps one | landed (sim 10%); resolves client_update-20261124-1 |
| SHR-HF-12 | Blackhand's Breadth physical, unresistable | wrong; already mechanics-20261125-2 |

Also: Fengus' Ferocity 115 spell power is now in the client (SHR-HF-07); its fix is still unmerged on `forever-update/2026-11-24-71091` (update on client_update-20261030-4).

## Conclusions

- Two new medium findings (hawk crit, demon crit), neither sized: both are server-side numbers with no client value.
- The Balance rework is the biggest change this build window; it is mechanics-20261125-1, not refiled here.
- Resolved: client_update-20261124-1 (Sayge's 10% is right). mechanics-20261124-1 asks the same question and can close on SHR-HF-11.
- Left open: the Shahram sim lets up to 8 Fist auras stack (`item_effects.go:573`); whether the game refreshes instead is untested and not part of this hotfix.
