# Round 70753: Warlock

Worker for the Warlock DPS specs (Affliction, Demonology, Destruction Shadow Bolt, Destruction fire), branch
`round-70753-warlock`. Brief: `briefs/warlock.md`.

## Status

- Phase A (research and mechanics): done, pushed on `round-70753-warlock`.
- Phase B (re-optimize, races, audit, land): waits for A1's talent trees and A2's racials.

## Sources checked (Phase A, 2026-11-03)

- Client 1.60.1.70753 from wago.tools (`wago.py fetch 1.60.1.70753 --set client_update`), diffed against 70707
  (SpellEffect, Spell, TraitDefinitionEffectPoints and CurvePoint). The only Warlock row that changed is Improved Drains
  (403511: effect 0 drops Wrack's mask bit, a new effect 1 covers Drain Mana), which the 10-30 hotfix had already put in.
  Wrack (1316697), Shadow and Flame (426316 curves, auras 1293816 / 426311), Amplify Curse (18288), Bane of Agony and
  Drain Soul are byte-identical to 70707.
- Server hotfixes for 70753 on wago.tools: none yet (SpellEffect, SpellMisc, SpellName searched). The 70707 hotfix rows
  for 18288 and Bane of Agony are the ones the client already has.
- Wowhead Forever tooltips (dataEnv 16): still the old values (Wrack 36 a second, Shadow and Flame's auras 10%).
- foreverchanges.pro patch notes: the same lines as Blizzard's, worded the same way (Wrack "86 Shadow damage per second at
  level 60 instead of 36 ... twice the spell damage, 0.286 per second"; Shadow and Flame "4% per rank instead of 2%";
  Bane of Agony "now updates its damage while it runs"; Amplify Curse "now really works on Curse of Exhaustion and Bane
  of Agony").
- Where the notes change a value the client still has the old one for (Wrack, Shadow and Flame, pet health), nothing
  newer contradicts them: the client rows are unchanged since 70707, not changed to another value. The sim takes the
  notes as server-side changes, and says so in the code. Open question 1 below.

## Ledger

| Note line | Tier and evidence | Sim status | Fix |
|---|---|---|---|
| Spellstones now only increase Shadow damage | Client 70630 (Misc 36 → 32 on the Spellstone damage rows) | landed (70630) | none; `TestWarlockSpellstone` |
| Pets gain 2 Health per Stamina (was 7.5) | Notes; pet scaling 416189 has 0 points in the client, the server's shares are unpublished | wrong (7.5) | `sim/warlock/pet.go`: 2. No DPS effect (no raid damage reaches the demon). `TestWarlockDemonStatInheritance` |
| Bane of Agony dynamically updates its damage | Notes; FC "updates its damage while it runs" | landed: Forever dots tick live (`core.Dot.liveTick`), Agony included since it calls `dot.Snapshot` | none; new `TestWarlockBaneOfAgonyLive` (the comment in `sim/core/spell_result.go` naming Agony as an exception is stale; shared file, left alone) |
| Drain Soul's shard chance on damage reduced | Notes | n/a: the sim doesn't count Soul Shards (no preset spends one) | none |
| Amplify Curse affects Curse of Exhaustion and Bane of Agony | Notes; client 18288 mask still Curse of Weakness only, Agony's dummy effect 2 carries it (server script) | landed: the sim kept +50% of Agony's base | none; `TestWarlockAmplifyCurseAgony`. Findings client_update-20261027-3 and mechanics-20261027-2 moved to Resolved. Phase B's talent search tries Amplify Curse (rotopt knob `amplifyCurse`) |
| Improved Drains: Drain Mana, no longer Wrack | Client 70753 (403511) = 10-30 hotfix | landed (10-30) | none |
| Wrack: coefficient doubled (0.286 a second), base +2.5 a level (36 → 86 at 60) | Notes and FC; client 70753 still 36 and 0.143 (unchanged since 70707) | missing | `sim/warlock/wrack.go`: 36 + 2.5 × (level − 40), coefficient 0.286 a 1 sec tick. `TestWarlockWrack70753`. No current preset takes Wrack; Phase B's search decides |
| Shadow and Flame 4/8/12/16/20% (was 2/4/.../10%) | Notes and FC; client trait curves 426316 effects 2 and 3 still 2 to 10 | missing | `sim/warlock/talents.go`: 4% a point. `TestWarlockShadowAndFlame70753`. Destruction +66 DPS at 180 s, Destruction fire +19 (3k iterations, current presets) |
| Low-rank proc penalty removed | Notes | n/a: the sim has no such penalty for warlock talents, and presets cast top ranks | none (A2's shared part) |
| Orc: Command (+5% pet damage) back, Savage (1% crit) removed, Blood Fury 2 min | Notes; client 70753 adds 20575/20576 Command | A2 (shared racials) | Phase B: check Command reaches the warlock's demons (registered before the racials?) and re-run the race sweep; Affliction, Destruction and Demonology are Orc now |
| Human Ruthless 1%, Human Spirit 10%; Dwarf Artisan's Precision; Troll Berserking 15 s; Undead Touch of the Grave 15 s; Skyborne haste 2% | Notes, client 70753 | A2 | Phase B race sweep |
| All races' base stats corrected | Notes | A2 | Phase B race sweep |
| Throw windup, proc-root DR, training costs, Grimoire costs | Notes | n/a for warlock DPS | none |
| Threat and healing rows | | n/a | tank/healer sessions |

## Notes-only values

The orchestrator's source rule for this round (`README.md`, a32a20024): a note line that neither the client nor a hotfix
carries is modelled at the notes' value under Forever, citing the notes and "client 70753 unchanged". The only hotfix
since 70753 (foreverchanges.pro, 2026-11-03) has no Warlock lines.

- Wrack: 86 a second at 60 (36 + 2.5 a level above 40) and 0.286 spell damage a tick. Client 70753 unchanged (36, 0.143).
- Shadow and Flame: 4% a point. Client 70753 unchanged (trait curves 2 to 10).
- Demons' health: 2 a point of the warlock's Stamina. The client has no value (416189's points are 0); was 7.5 from beta
  pet sheets.
- Amplify Curse on Bane of Agony: +50% of the base. The client's mask leaves Agony out, but Agony's dummy effect carries
  it; the sim already did this.

Wrack and Shadow and Flame apply under Forever only in effect: both are Forever's talents and exist in no other ruleset.

## DPS effect of Phase A (current presets, 180 s, 3,000 iterations, master vs branch)

| Build | Before | After | Change |
|---|---|---|---|
| warlock__affliction | 1015.85 | 1015.85 | 0 (no Wrack, no Shadow and Flame) |
| warlock__demonology | 1055.37 | 1055.37 | 0 |
| warlock__destruction | 996.39 | 1062.58 | +66.2 (Shadow and Flame 5/5: 20% Shadow) |
| warlock__destruction_fire | 977.65 | 996.89 | +19.2 |

## Open questions

1. Notes vs an unchanged client (Wrack, Shadow and Flame, pet health). The 70753 client keeps the old numbers, and no
   hotfix has the new ones yet. The sim follows the notes, since the client rows weren't changed to anything else and
   foreverchanges reads the notes the same way. Recommendation: keep the notes' values, and re-check the hotfix rows
   once the realms open on November 4.
