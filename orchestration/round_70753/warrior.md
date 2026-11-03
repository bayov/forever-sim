# Warrior: round 70753 (November 2 development notes)

Worker: the Warrior class worker, branch `round-70753-warrior`. Brief: `briefs/warrior.md`.

## Phase A: research and mechanics

### Sources checked (2026-11-03)

- Dev notes post 9 (2026-11-02), `orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`, "- Warrior"
  section, Races and Bug Fixes.
- Client 1.60.1.70753 (wago.tools, `wago.py fetch 1.60.1.70753 --set client_update`). **None of the warrior changes in the
  notes are in it:** Slam's category cooldown is still 18000 on every rank (1464, 8820, 11604, 11605, 1240193, 1310196-200),
  Improved Slam 12862 effect 2 is still -3000 with no rank curve, there is no Reckless Swings spell, Improved Whirlwind
  24431 is the old Burning Crusade row (aura 107 misc 14, -30), Flurry 12966 still 30 on the buff with the 5 point tree.
  The ledger's 70753 section agrees: its Trait changes are in the paladin, druid and mage trees only.
- Server hotfixes for 70753 (`hotfixes.py fetch 1.60.1.70753 --set talents --set spelltext`): no rows at all yet.
- foreverchanges.pro /patch-notes: lists the warrior lines under "November 3, 2026 — Announced for launch / Blizzard's
  notes" (Defensive Stance 80%, Reckless Swings, Improved Whirlwind, Slam 15 sec, ...). Nothing in its data yet.
- So the notes are the highest tier that has these changes. They postdate the client and hotfix data, so there is no
  disagreement to resolve: the client simply hasn't caught up. Each code comment says "client 70753 still has the old
  value". When a hotfix or a newer build carries them, re-check the values (the notes give no Reckless Swings GCD,
  no spell id, and no ranks for Improved Whirlwind's rows).

### Ledger

| Note line | Tier and evidence | Sim status before | Fix |
|---|---|---|---|
| Victory Rush tooltip fixed | notes | n/a (not in a DPS rotation; needs a killing blow) | none |
| Defensive Stance threat 30% → 80% | notes | n/a | tank/healer sessions (tanking) |
| Slam cooldown 15 sec (was 18) | notes; client 70753 still 18000 | wrong (18 sec) | `sim/warrior/slam.go`: 15 sec under Forever. Test `TestForeverWarriorSpellValues` |
| Arms: Improved Tactical Mastery 3 points, 5/10/15 Rage (was 5 points of 3) | notes; client 70753 still 3 a point (12295) | wrong (3 a point) | `sim/warrior/stances.go`: 5 a point, points capped at 3 (old 5 point strings keep 15). Test `TestForeverWarriorImprovedTacticalMastery`. Tree max rank: Phase B (A1's remap) |
| Arms: Two-Handed Weapon Specialization 2/4/6% (was 1/2/3%) | notes; client 70753 still 1% curve | wrong (1% a point) | `sim/warrior/talents.go`: 2% a point under Forever. Test `TestForeverWarriorTwoHandedSpecialization` |
| Arms: Bloodthrill 2 points, 10/20% (was 5 points of 4%) | notes; client 70753 still 4% a point | wrong | `talents.go`: 10% a point, capped at 2 (old 5 point strings get 20%). Test `TestForeverWarriorBloodthrillChance`. Tree: Phase B |
| Arms: Improved Slam takes 2.5/5 sec off Slam's cooldown (was 3/3) | notes; client 70753 still -3000 flat | wrong (3 sec at either rank) | `slam.go`: 2.5 sec a point. Tests in `TestForeverWarriorSpellValues` (15 / 12.5 / 10 sec) |
| Fury: Blood Craze and Gore Drinker don't generate threat | notes | n/a for DPS (Gore Drinker isn't simulated: healing only) | tank/healer sessions |
| Fury: Blood Craze won't activate when sitting | notes | n/a | none |
| Fury: Booming Voice also cuts Piercing Howl's cost 5-25% | notes | n/a (Piercing Howl isn't in a DPS rotation or the sim) | none |
| Fury: Enrage to row 3 (was 4) | notes | tree only | Phase B (tree) |
| Fury: new talent Reckless Swings (row 4, 1 point, 15 sec cd, 25% of base health, next 4 melee attacks +5 Rage each, 8 sec) | notes (no client spell yet) | missing | Phase B: needs the talent field. Plan: an off-GCD self buff with 4 charges, each landed white swing (main or off hand) adds 5 Rage and uses a charge; the health cost is modelled as nothing for DPS (brief), and the 5 Rage is flat, not doubled by a crit (standards 21 doubles the swing's own Rage). Both are assumptions to note |
| Fury: Gore Drinker to row 5, under Blood Craze | notes | tree only (no DPS) | Phase B (tree) |
| Fury: Death Wish no longer required for Flurry | notes | tree only | Phase B (tree) |
| Fury: Flurry 3 points, 10/20/30% for 3 swings (was 5 points of 5%) | notes; client 70753 still the 5 point tree | wrong | `talents.go` `makeFlurryAura`: 10% a point, capped at 3, under Forever. Tests `TestForeverWarriorFlurrySpeed`, `TestForeverWarriorDeepWoundsAndFlurry`. Tree: Phase B |
| Fury: new talent Improved Whirlwind (row 6, 2 points, -1/2 sec Whirlwind cd, +15/30% to the closest enemy) | notes (24431 in the client is the old row) | missing | Phase B: needs the talent field. Plan: cooldown 10 − points sec; +15% a point on the main and off hand hits on the first target only |
| Fury: Bloodthirst needs Death Wish | notes | tree only | Phase B (tree) |
| Protection: Defiance 7/13/20% | notes | n/a | tank/healer sessions (tanking) |
| Talents lose the low-rank proc penalty | notes | n/a (no warrior talent has one) | A2 (shared part) |
| Races: Dwarf Artisan's Precision, Human Ruthless 1% / Spirit 10%, Orc Savage removed and Blood Fury 2 min, Skyborne Wind-Blessed 2%, Tauren Endurance, Troll Berserking 15 sec, Undead Touch of the Grave 15 sec | notes (10-26 racial pass), client 70753 | landed (ledger 70753 section, `sim/core/racials_forever.go`) | none; Phase B race sweep |
| Bug fix: race base stats (Tauren warrior +5 Str over Human) | notes | A2 | A2; Phase B race sweep |
| Bug fix: Throw's 0.5 sec windup | notes | A2 | A2 (no warrior rotation throws) |
| Bug fix: roots from procs get diminishing returns (Improved Hamstring) | notes | n/a (bosses) | none |

### DPS effect of Phase A (goldens)

`TestP1DPSWarrior` (Forever, talents `30305013-25053005152110501`: 5 points of Flurry, now read as 3 at 30%, and
Improved Tactical Mastery 5, now 15 Rage): Average-Default 1518.5 → 1562.9 (+2.9%). Tank goldens didn't move.

### Open for Phase B / questions

- If A1 builds the warrior tree from client 70753, it won't have these changes (no Trait rows). The warrior tree then
  needs a manual edit from the notes (Improved Tactical Mastery 3, Bloodthrill 2, Flurry 3 without the Death Wish
  prerequisite, Enrage row 3, Reckless Swings row 4, Gore Drinker row 5 under Blood Craze, Improved Whirlwind row 6,
  Bloodthirst under Death Wish). Question for the orchestrator: does A1 do that, or do I?
- The rank caps in the code (`min(points, N)`) keep old strings sane until the presets are remapped; they can stay.
