# Round 70753: Warlock

## Follow-up: exploratory build search (2026-11-03, after the first landing)

The user asked for a wider search. New seeds (each a talent search, 180 s, 2000 iterations a candidate, then a knob
search at 3000), all checked against the tree rules first:

| Seed | Result at 180 s | Verdict |
|---|---|---|
| Wrack + Demonic Sacrifice (Imp sacrificed, no demon out), 05500003035001351-0050003201-045 | 05500021035001351-0050003201-005004, 1032.3 | **New Affliction preset** |
| Destruction + Demonic Sacrifice (Imp), 055-0050003201-055030500310035 | 154-0050003201-055030500310035, 1033.3 | 30 behind Destruction, same play |
| Shadow Bolt Affliction + Demonic Sacrifice (Imp), 2550102003520005-0050003201-055 | unchanged, 934.0 | 82 behind |
| Wrack + Improved Shadow Bolt / Ruin side, 05500003035001351--0550305002 | climbed back to the Succubus Wrack build, 999.6 | nothing new |
| Wrack with Amplify Curse required | 1000.4 | even with 999.9 |
| Destruction with Shadowburn (last 20%) required, 25500000035--055030510310035 | 1064.5 | even with 1063.4, and Shadowburn's shard isn't charged |

The Wrack + Imp sacrifice build: knob searches at 120 and 300 s moved nothing (300 s alone wanted doomAlign=6, +1, a
length trade-off). Races at 180 ± 20 s, 20,000 iterations, seeds 1 / 1000003: Undead 1034.1 / 1034.0, Orc 1032.8 / 1033.3,
Human 1030.7 / 1031.2, Gnome 1026.5 / 1026.7, Troll 1023.4 / 1023.8, so Undead (best overall and Horde), Human (Alliance).

What changed (commit 734fe0e and after):
- `warlock__affliction` and its two curse duty builds are the Wrack build with the Imp sacrificed, Undead. Finals:
  1004.9 / 1034.1 / 1032.2 at 120 / 180 / 300 s; curse duty 1025.1, recklessness duty 1019.9 at 180 s.
- Shadow Bolt Affliction is its own preset and request, `warlock__affliction_shadow_bolt` (Orc, 997.4 / 1015.6 /
  996.8), with its old rotation and Gnome override renamed to it.
- The Succubus Wrack build (`warlock__affliction_wrack`, added earlier this round, never in the report) is gone. The
  `warlock__affliction-duty-curse.Gnome.json` override (tuned for the Shadow Bolt rotation) is gone.
- mksettings: `-spec affliction` is the Imp sacrifice setup, `-spec affliction_shadow_bolt` the Succubus one.
- compsearch `leftOut`: `warlock__affliction_shadow_bolt` in place of `warlock__affliction_wrack`.
- Destruction (1063.4) stays the class default.

For the orchestrator's regeneration: `warlock__affliction` changed build (talents, pet, race, rotation), and
`warlock__affliction_shadow_bolt` is the old `warlock__affliction` under a new id.

## Final report (2026-11-03, first landing; superseded for Affliction by the follow-up above)

Landed on master at 65c6660 (merge of `round-70753-warlock`; full sim test suite green with -timeout 40m, tsc clean). Ledger, sources and method below.

**What changed** (sources: Blizzard's development notes of 2026-11-02; the 70753 client and the 2026-11-03 hotfix carry
none of the warlock damage lines, so these are notes-only values, standards 37):
- Wrack: 0.286 spell damage a tick (was 0.143), base 36 + 2.5 a level above 40 (86 at 60). Enables the new Affliction
  Wrack preset; no other preset takes Wrack.
- Shadow and Flame: 4% a point (was 2%). Destruction +65.9 DPS at 180 s, Destruction fire +19.6.
- 2026-11-04: hotfix push 113527 (logged 2026-11-03) carries both values (Wrack: SpellEffect 1349556 of 1316697, 36 base,
  2.5 a level, 0.286; Shadow and Flame: curves 111671 and 111672, 4/8/12/16/20), so they are hotfix values now, not
  notes-only. No DPS moves.
- Demons: 2 health a point of the warlock's Stamina (was 7.5). No DPS effect.
- Already in the sim and now tested: Spellstone Shadow only, Bane of Agony's live ticks, Amplify Curse +50% on Agony's base
  (findings client_update-20261027-3 and mechanics-20261027-2 resolved), Orc Command +5% on the demons (A2's racials).
- Talent tooltips: `tools/forever_talents/notes/warlock.json` (Shadow and Flame, Wrack).
- New preset: Affliction Wrack (`05500003035001351-0002-0050055003`), deep Affliction where Wrack is half the damage; its
  talents, rotation (`level60_affliction_wrack`), preset build and report request `warlock__affliction_wrack`.
- Class default: Destruction (highest at 180 s) replaces Demonology in the UI defaults and `isDefault`.

**Per build** (20,000 iterations, seed 1; 180 s is 180 ± 20 s on the request, 120 and 300 fixed). Old = master before the
round (DPS report v-current published numbers at 180 s; the pre-round binary at 120 / 300 s):

| Build | Race | Talents | Old 120 / 180 / 300 | New 120 / 180 / 300 |
|---|---|---|---|---|
| Destruction (default) | Orc | 255000000351--055030500310035 | 976.6 / 997.5 / 978.5 | 1040.4 / 1063.4 / 1043.5 |
| Demonology | Orc | 051-0305003131221001351-0550004 | 1037.4 / 1054.5 / 1028.4 | 1036.8 / 1053.9 / 1028.2 |
| Affliction | Orc | 2550102003520005--0550305003 | 997.9 / 1015.8 / 997.3 | 997.4 / 1015.6 / 996.8 |
| Affliction Wrack (new) | Orc | 05500003035001351-0002-0050055003 | – | 966.0 / 999.9 / 992.7 |
| Destruction fire | Undead | 252-0050003201-0050345003101351 | 965.4 / 977.1 / 954.2 | 983.6 / 996.7 / 974.0 |

Duty presets at 180 s (old → new): Affliction curse 1005.2 → 1004.9, recklessness 999.6 → 999.2; Demonology curse
1043.8 → 1043.6, recklessness 1037.4 → 1036.6; Destruction curse 987.2 → 1052.7, recklessness 981.7 → 1046.4;
Destruction fire curse 967.5 → 987.0, recklessness 961.2 → 980.7. The small drops on Affliction and Demonology are A2's
race base stats (−0.2 to −0.8).

**Races** (180 ± 20 s, 20,000 iterations, seeds 1 and 1000003): every preset keeps its race.

| Build | Human | Gnome | Orc | Undead | Troll | Best overall / Horde / Alliance |
|---|---|---|---|---|---|---|
| Affliction | 1013.5 | 1011.2 | **1015.6** | 1012.5 | 1007.2 | Orc / Orc / Human |
| Affliction Wrack | 995.4 | 990.6 | **999.9** | 997.3 | 988.0 | Orc / Orc / Human |
| Demonology | 1051.3 | 1050.8 | **1053.9** | 1050.7 | 1045.3 | Orc / Orc / Human |
| Destruction | 1061.5 | 1059.1 | **1063.4** | 1060.1 | 1052.5 | Orc / Orc / Human |
| Destruction fire | 991.1 | 988.5 | 994.4 | **996.7** | 985.7 | Undead / Undead / Human |

Seed 1000003 (leader / runner-up): Affliction Orc 1015.6 / Human 1013.3, Wrack Orc 999.8 / Undead 997.4, Demonology
Orc 1054.7 / Human 1052.2, Destruction Orc 1063.5 / Human 1061.5, fire Undead 997.0 / Orc 995.2. Gnome sims each build
with its race override (Eureka timing) where one exists; Affliction Wrack has none (Gnome is 9 behind, an override
couldn't close that).

**Audit** (the DPS passes, applied to the warlock builds):
- results: all 13 requests re-verified and published at 20k / seed 1; the new request equals mksettings' output for its
  arguments (no drift); the only unread-field leftovers are the standing `sanctityAura` / `curseOfShadow` ones.
- rotations: talent and knob searches for all five builds (nothing moved for the four existing ones); Affliction Wrack's
  timeline checked (Wrack chains while Corruption and the bane are up, dots refreshed, a missed Wrack recast). Its lines
  state their intent in the meta and presets (standards 31). Its 120 s number trails more (−31 against −16 at 180) since
  the opening dots take a larger share of a short fight; the knob search is at 180 s, and Affliction stays the spec's
  default either way.
- gear: unchanged this round (no warlock gear edit).
- consumables: `tools/audit/consumables/scan.sh` clean for the warlock builds; the new request copies Affliction's.
- mechanics / client_update: Phase A ledger below; the two parked Amplify Curse findings resolved.

**Open questions**
1. Notes-only values (Wrack, Shadow and Flame, demon health): keep the notes' numbers and re-check the hotfix rows when the
   realms open on November 4. Recommendation: keep.
2. Report: `warlock__affliction_wrack` is a new build for the orchestrator's next regeneration (no duty variants; the
   four existing builds cover curse duty).
3. Shared, not changed: `sim/core/spell_result.go`'s comment on `liveTick` still names Bane of Agony as a dot that doesn't
   tick live; it does (TestWarlockBaneOfAgonyLive). Comment only.

Shared: one line in `sim/compsearch/pool.go`, `warlock__affliction_wrack` added to `leftOut` (the comp search keeps one build per spec; Affliction is the Affliction build). No shared sim code changed.

Worker for the Warlock DPS specs (Affliction, Demonology, Destruction Shadow Bolt, Destruction fire), branch
`round-70753-warlock`. Brief: `briefs/warlock.md`.

## Status

- Phase A (research and mechanics): done, pushed on `round-70753-warlock`.
- Phase B: done on A1's and A2's sources (master 5eb381750 and later); see the final report above.

## Phase B so far (2026-11-03)

- Client trees: A1 found no change in the warlock tree, so no preset string moved. `tools/forever_talents/notes/warlock.json`
  puts the notes' Shadow and Flame and Wrack values in the talent tooltips.
- Talent searches (180 s, 2000 iterations a candidate, 10000 to confirm), each from the preset, then a knob search: nothing
  moved for Affliction (1015.9), Destruction (1063.6), Destruction fire (996.4) or Demonology (1055.4). Seeds that tried
  Conflagrate and Shadow and Flame on Affliction can't reach Shadow and Flame and lose (934.8).
- New preset, Affliction Wrack: a talent search from deep Affliction with Wrack kept (25500020135201351--0550303, 976.6)
  ends at 05500003035001351-0002-0050055003, 999.9 at 180 s; the knob search moved nothing. 988.1 / 999.9 / 993.4 at
  120 / 180 / 300 s against 1015.6 / 1015.9 / 998.3 for Affliction (10000 iterations). Wrack is 520 of its 1000 DPS (22.6
  casts), so it plays differently from the Shadow Bolt build: shipped as its own talents, rotation, preset build and
  report request (`warlock__affliction_wrack`, no duty variants: the four existing builds cover curse duty).
- Rotation check (timeline, one fight): Wrack chains back to back while Corruption and the bane are up, Siphon Life,
  Corruption and Immolate are refreshed when they drop, the Bane of Doom / Agony swap works, a Wrack that misses is recast
  after the GCD. It ends with about 2000 mana unspent (it never needs a Life Tap past the opening), which no knob turned
  into damage.

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
  pet sheets, which it stays outside Forever (`pet.go`, gated under `Env.IsForever()`).
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
