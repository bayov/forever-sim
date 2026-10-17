# Pass: synthetic gear

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running gear auditor's handoff (2026-10-17); keep it current when the method changes.

This file is the brief for whoever runs the synthetic gear audit next, with only the repo to go on. The role brief is `orchestration/auditor_gear.md`, with the common auditor rules in `auditors_common.md`. The rules you check against are in `orchestration/standards.md` and `orchestration/gear_targets.md`.

The auditor never changes code, presets or docs, and doesn't commit (the audit run commits its files). No PR comments. Don't use `tools/rotopt`, and don't put model names in anything you write. Findings go to `orchestration/audit/findings/gear.md`.

## What is audited

The Level 60 presets run on synthetic Phase 1 gear:
- Items with ids 990000-998999, defined in `tools/database/forever_synthetic_items_<class>.go`. The shaman's Enhancement items sit in the base `forever_synthetic_items.go`.
- Id blocks by class: shaman 990, warrior 991, paladin 992, hunter 993, rogue 994, priest 995, mage 996, warlock 997, druid 998.
- The sets that use them: `ui/**/gear_sets/level60*.gear.json`. There are 23 sets at baseline.
- The on-use trinket effects in `sim/<class>/forever_synthetic_trinkets.go`.
- Real items and enchants a set uses: `tools/database/forever_items.go` and `forever_enchants.go`.

## Baseline (end of pass 36, 2026-10-17 07:43 UTC)

- **Master:** `origin/master` d2633f4a4 is clean. Its preset gear is identical to 6c64ef928 (pass 32, full audit). The game client is 1.60.1.70425.
- **Rulings:** the last ruling commits seen are ae4d920df on `origin/claude/wow-forever-dps-presets-m9u15h` (the coordination branch) and d2633f4a4 on master.
- **Class branches:** each `origin/claude/wow-forever-dps-presets-m9u15h-<class>` had no gear changes of its own against master (`git diff origin/master...<branch>` was empty after filtering).
- **Reference outputs:** in `orchestration/audit/ref/gear/` (see MANIFEST.md).
- **The 23 sets:**
  - Warrior: Arms, Fury.
  - Paladin: Ret.
  - Hunter: ranged (BM/MM/Lone Wolf), Survival.
  - Rogue: Daggers, Combat Daggers, Mutilate, Swords, Seal Fate Swords.
  - Priest: Shadow, Smite.
  - Shaman: Elemental, Enhancement.
  - Mage: FFB/Frost/Fire, Arcane.
  - Warlock: main, Demonology.
  - Druid: Balance, Balance Wrath, Feral.

## A pass, step by step

```
source tools/audit/setup.sh                # AUDIT_WORK and the toolchain
export AUDIT_TREE=$PWD                     # repo checkout (master)
export AUDIT_WORK=$AUDIT_WORK/gear; mkdir -p $AUDIT_WORK
export AUDIT_REF=$AUDIT_TREE/orchestration/audit/ref/gear
export AUDIT_GITREF=origin/master
git fetch -q origin
F='synthetic|gear.json|forever_enchants|trinkets|items.go'
```

1. **Rulings.** Look for new commits that touch `orchestration/standards.md` or `orchestration/gear_targets.md`, on master and on the coordination branch:
   ```
   git log --oneline <last-seen>..origin/master -- orchestration/standards.md orchestration/gear_targets.md
   git log --oneline ae4d920df..origin/claude/wow-forever-dps-presets-m9u15h -- orchestration/standards.md orchestration/gear_targets.md
   ```
   Read each diff. Most new standards cover raid buffs, Rage, racials or ranks, not gear. Examples: 21 (Rage), 23 (GCD raid actions), 25 (Grace of Air) and the Flametongue Totem ruling. Only items 1b, 2, 3, 4, 5, 5b, 6, 7, 8a and 22 are gear rules.
2. **Changes.**
   ```
   git diff --name-only <baseline>..origin/master | grep -E "$F" | grep -v dps_report
   for c in warrior paladin hunter rogue priest shaman mage warlock druid; do
     git diff --name-only origin/master...origin/claude/wow-forever-dps-presets-m9u15h-$c | grep -E "$F" | grep -v dps_report
   done
   ```
   - Use three dots: a branch that is only behind master isn't a change.
   - For `forever_items.go`, only check whether a changed item id appears in a level60 gear set. New low-level real items are normal and not your concern.
   - For `sim/<class>/items.go`, read the diff. Relic and real-item effects aren't budget matters.
   - **Quiet pass:** nothing filtered changed and there's no new gear ruling. Stop there; no messages.
3. **Audit** when gear did change, on master or on the branch that changed it.
   ```
   tools/audit/gear/extract.sh origin/master $AUDIT_WORK/sNN
   python3 tools/audit/gear/gen_db.py
   python3 tools/audit/gear/lin2.py $AUDIT_WORK/sNN > $AUDIT_WORK/lin.txt
   diff <(sed 's/s37/sX/' $AUDIT_REF/lin2_d2633f4a4.txt) <(sed 's/sNN/sX/' $AUDIT_WORK/lin.txt)
   python3 tools/audit/gear/std.py $AUDIT_WORK/sNN > $AUDIT_WORK/std.txt     # read item levels, enchants, sets; ignore cost lines
   tools/audit/gear/extract.sh d2633f4a4 $AUDIT_WORK/base
   python3 tools/audit/gear/wepcheck.py $AUDIT_WORK/sNN $AUDIT_WORK/base   # look at CHANGED / NEW / DPS-DRIFT!
   python3 tools/audit/gear/trinkets.py origin/master                      # any '!' is a finding
   ```
   - For a changed item, read its Go comment. It names the Era item it stands in for and the speed source.
   - For a real item's numbers, `python3 tools/audit/gear/tt.py ID` fetches the wowhead Era tooltip.
4. **Report.**
   - A break is a finding in `findings/gear.md` (format in `PASS.md`), with the item id, the number and the rule (the standards item), `needs: orchestrator fix`.
   - A clean pass with something new (a new set, new weapons, a ruling that touches gear) gets a short `reports/gear.md`.
   - A quiet pass only updates `state/gear.json`.
   - When the reference outputs change for a known reason, write the new ones as `ref/gear/*_<sha>.txt`, delete the old ones, and record the sha in `state/gear.json` as `baseline`.

## Checks and tolerances (standards 1b and 5a)

**Weights**
- Str, Agi, Sta, Int and Spi: 1.
- Spell power 0.86; school spell power 0.70; healing 0.45.
- AP 0.5; RAP 0.4. AP and RAP on one item count once, as AP.
- Hit and crit rating: 1 each. MP5: 2.4.

**Armor pieces**
- Line: L(ilvl) = 1.125·ilvl + 2.0.
- Slot multipliers: head, chest and legs 1; shoulder, hands, waist and feet 0.75; neck, back, wrist and rings 0.5625.
- Archetype factor f, for each multiplier group (1 / 0.75 / 0.5625):

  | Archetype | f |
  |---|---|
  | caster | 1.053 / 1.026 / 0.995 |
  | physical (warrior, rogue, Feral) | 0.950 / 0.959 / 0.911 |
  | hunter | 1.028 / 1.056 / 0.977 |
  | hybrid (Ret, Enhancement) | 1.093 / 1.103 / 1.048 |

- Linear total must be slot × L × f within ±1%.
- 1.7-norm must be within ±6% of slot × (0.661·ilvl + 2.92).

**Item levels:** head, shoulder, chest, hands, legs and feet 66; neck, back, wrist, waist and rings 71; two-handers 76.

**Weapons, as multiples of L(ilvl)**

| Weapon | Multiple |
|---|---|
| physical 2H / 1H | 0.90 / 0.39 |
| hunter 2H / 1H | 0.98 / 0.42 |
| Ret / Enhancement 2H | about 1.035 (1.04 is accepted); 1H 0.45 |
| Feral 2H | 0.90 |
| caster staff (ilvl 76) | 2.21 |
| caster main hand + off hand | 1.82 + 0.53 |
| ranged | 0.24 |
| wand | 0.26 |
| relics | stat-less |

**Weapon damage and speed (standards 22)**
- Damage per second stays at the Era Phase 1 BiS weapon's value. wepcheck flags drift over 0.02, because damage is rounded to 0.1.
- Speed and spread come from a named Era Phase 1 BiS weapon (the list is in standards 22). Two-handers run 3.4 to 3.7 sec. No fist weapons.

**Trinkets**
- Stat trinket: linear 45.9 ±1%, 1.7-norm 36.8 ±6%, no Stamina.
- On-use trinket: a 180 s average of 27 for spell power, 37 for AP, Str, Agi or RAP. The usual shapes: 140 SP, 335 AP, 415 RAP or 165 Str for 20 s every 2 min.

**Stat mix (5a), hit-adjusted, against `gear_targets.md`**
- Each stat within ±3 points.
- Stamina per spec within ±1.5.
- Spirit about 10% for mana users, hunters included; none for rage and energy users.
- Top stat at most 55% of the non-Stamina share.
- Hybrids (5b): Ret and Enhancement hold their spell power and Intellect amounts; only the physical stats rescale with hit.

**Other rules**
- Set bonuses: 2-piece only (item 6).
- Enchants (item 7):
  - The best Phase 1 enchants, necklace enchants included, with Lesser Arcanums on head and legs.
  - Banned: Arcanum of Focus, Rapidity and Protection; Weapon - Spell Power; Revelation for casters.
  - Allowed: Rugged kits, Spellblasting and the other Forever enchants.
- Real items must be obtainable at Forever launch (item 8a).

## Settled; don't re-flag

- **std.py false flags:** "2H Agility (check source)" on Feral and ranged hunter, and "Gloves - Superior Agility" on the Survival off hand. These are effect-id labels in std.py's ban table, and the enchants were accepted in earlier passes. The cost, armor12, "caster weapons total" and "5a:" lines in std.py follow the old budget model; lin2.py replaced them.
- **Mutilate** reuses the dagger armor and is held to the daggers targets. **Seal Fate Swords** reuses the swords armor. **Combat Daggers** reuses the dagger armor with mace off hand 994118.
- **Off hands after the Windfury main-hand-only and off-hand Rage rulings (2026-10-16):**
  - Combat Swords 994014 and Combat Daggers 994118 use Ebon Hand's 2.5 sec mace; a mace takes Hack and Slash's armor ignore.
  - Fury 991015 uses Deathbringer's 2.9 sec axe. Its commit was marked WIP; re-check the item if it changes again.
- **Enchant Bracer - Lesser Intellect** is +3 since 70425. No preset uses it.
- **Howling Idol** reduces Shifting Power's cooldown by 0.5 s since 70425. That's a relic effect, not a budget matter.
- **The ranged hunter bow** is 993014 at 2.5 sec with Striker's Mark's spread. Survival keeps Rhok'delar's 2.9 sec.
- **The caster main hand** ruling is 46.2. Mage staves are ilvl 76.
- **Onyxia Tooth Pendant** is ilvl 74 (wowhead). The repo's db.json says 65.

## Open items and pending rulings

- **Fury off hand:** work in progress on the warrior branch (Enrage and Heroic Strike after the Rage rulings). A new off-hand item or speed may land; run the full audit when it does.
- No other gear ruling is pending.

## Gotchas

- lin2.py takes the archetype from the part of the file name after `--`. extract.sh names files `master--<path>`. Don't prefix with the class name: master carries every class's sets, and a class prefix gives false archetypes.
- lin2.py prints on-use trinkets as `[... on-use trinket ... (check sim code)]`. That's expected; trinkets.py checks them.
- The paladin's items use a helper constructor. audit.parse finds items by `Id: 99xxxx` blocks, so keep ids literal.
- A set's slot order in the gear.json files is head, neck, shoulder, back, chest, wrist, hands, waist, legs, feet, finger 1, finger 2, trinket 1, trinket 2, main hand, off hand, ranged. Enchant ids are effect ids; read their names from std.py's `ench:` line.
- Branch diffs: use `origin/master...<branch>` (three dots). Two dots lists everything a branch hasn't merged yet as changes.
- Cadence: the audit loop's gate starts a pass when a watched gear file changes on master (checked every 6 hours).
