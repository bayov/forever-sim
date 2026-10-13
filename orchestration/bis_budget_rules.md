# How real level 60 epics are budgeted per archetype, and how to price our synthetic BiS pieces

Measured 2026-10-13 to answer the user's ruling: predict Forever Phase 1 BiS as accurately as possible, with Forever's own pricing where it is known (ilvl <= 30 only) and real level 60 epics (Era, and SoD's level 60 phases) for everything else. It follows orchestration/forever_item_budget.md (section "ilvl <= 30 only") and does not change standards.md or gear_targets.md.

Scripts (orchestration/tools/, run with python3 from any directory): bis_data.py (loaders and archetypes), bis_model.py (p-norm fitter), bis_archetype.py (Q1, Q2), bis_rules.py (Q3), bis_trinkets.py (Q4), bis_proposal.py (Q5).

## Short answers

1. **Blizzard priced every archetype the same.** On the p-norm with the measured weights, caster, physical, hybrid and tank epics sit on one budget line. At p = 1.7, the exponent that fits Era level 60 best, every archetype lands within 0.3% of the line. At p = 1.5, hybrids look 2.5% "over", but that is only the stat count: residuals climb with the number of stats at p = 1.5 and are flat at p = 1.7. SoD is noisier (rms 9-11% against Era's 2.7%) and gives casters about 3-6% less than physical pieces.
2. **Only the stat count differs, and the linear total follows it.** Era epics carry 3.8 stats per piece for casters, 3.1 for physical, 3.9 for physical + Int (hunter-type mail) and 5.3 for hybrids. Against one linear line that becomes caster 1.01, physical 0.92, hunter-type 1.00 and hybrid 1.17.
3. **Special rules:**
   - **Caster weapons:** Forever doubles them. A Forever caster staff carries 2.1 chest budgets (2.4 times a physical two-hander's linear stats), and a caster one-hander 1.8 chest budgets (4.7 times a physical one-hander). Era is 1.03 and 0.61.
   - **Spell power:** 0.86 holds.
   - **Hit and crit:** the same on caster and physical pieces (1.0-1.05).
   - **Healing:** costs 0.53 of a spell power point.
   - **Two-handers:** cost 2.3-2.4 one-handers. Wands are priced like bows.
   - **Relics:** carry no stats.
   - **Hybrids:** SoD itemized them as Str + Holy spell power for Ret and Str + Int with no spell power for Enhancement. Forever's own low-level hybrid pieces sit on the common line in p-norm value and carry 1.15 times a physical piece's linear stats.
4. **Trinkets:** a stat trinket is worth 0.74 of a chest on the p-norm, or 0.56 x L(ilvl) in linear stats. An on-use trinket averages 0.74 (spell power) to 1.0 (attack power) of that over a 180 s fight.
5. **Proposal:**
   - Spend each piece linearly at slot mod x (1.125 ilvl + 2.0) x f(archetype, slot tier), with f measured on real epics.
   - Check each piece on the 1.7-norm. That check pins the stat count to the archetype's real one (caster 4 / 4 / 3-4, physical 3, hunter 4 / 4 / 3, hybrid 5-6 / 5 / 4-5 for T1 / T2 / T3), so stat count can't be gamed.
   - Give Forever's caster weapons their double budget.
   - Add two ilvl 71 trinkets at 0.56 x L(71).
   - Against today's 1a (re-priced at spell power 0.86), sets without trinkets change as follows: casters -5 to -6%, physical melee -8%, hunters -1%, Ret / Enh +3%, feral -8%.

## Data

- **Era:** Wowhead's Classic gear planner dump (versionNum 11300), quality 3-4, with source zones and NPCs. Phase 1 content is Molten Core, Onyxia, Azuregos, Kazzak, and the BRD / BRS / Stratholme / Scholomance items at ilvl 60-66.
- **SoD:** wowsims/sod's db.json, SoD's own records (ids 200000+), phases 4-6. Phases 7-8 have no tooltips in the dump, so their effects are unknown.
- **Forever:** the listing at ilvl 5-30.
- **Clean pieces:** a piece is "clean" when its tooltip has no Use:, proc or non-stat Equip: line and every stat is priced. Tank stats (defense, dodge, parry, block, block value) are priced here, so tank pieces count too.
- **Archetypes:**
  - **tank:** carries a tank stat;
  - **hybrid:** carries a physical stat (Str, Agi, AP, melee or ranged crit / hit) and a spell stat (spell power, healing, school power, MP5, spell crit / hit);
  - **physical:** carries physical stats only. It splits into "physical" and "physical+int", the second carrying Int, Spirit or MP5 too (hunter / shaman / paladin mail);
  - **caster:** carries spell stats or Int / Spirit only. It splits into "caster dps", "healer" (healing on top), "caster int/spi" and "caster mp5".
- **SoD's all-hit and all-crit:** SoD's "+1% hit / crit with all spells and attacks" counts once and does not make a piece physical or caster.
- **Measured weights:** spell power 0.86, healing on top 0.45, school power 0.71, AP 0.52, RAP 0.40, crit 1.01, hit 1.09, spell crit 1.01, spell hit 1.09 and MP5 2.39, all per point (14 crit rating or 10 hit rating per 1%). The tank weights come from the free fit: defense 1.56, dodge 12.4 per 1%, parry 20.5 per 1% and block 4.5 per 1%.

## Q1. Archetype pricing: one budget line

### Era level 60 (armor, blues ilvl 45-70 + epics 55-80, 1,215 clean pieces)

The exponent scan below fits every weight freely. The archetype factors (relative to physical) come from a second fit with a factor per archetype:

| p | rms | sp | hit | spell hit | crit | spell crit | AP | MP5 | caster | hybrid | tank |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1.5 | 3.65% | 0.84 | 1.10 | 0.98 | 1.02 | 0.99 | 0.52 | 2.40 | 1.003 | **1.041** | 1.020 |
| 1.6 | 3.05% | 0.85 | 1.10 | 1.03 | 1.02 | 1.00 | 0.51 | 2.40 | 1.003 | 1.022 | 1.007 |
| **1.7** | **2.66%** | 0.85 | 1.05 | 1.05 | 1.01 | 1.01 | 0.51 | 2.41 | **1.000** | **1.003** | **1.001** |
| 1.75 | 2.74% | 0.85 | 1.05 | 1.03 | 1.01 | 1.01 | 0.50 | 2.41 | 0.999 | 0.995 | 0.997 |
| 2.0 | 3.93% | 0.85 | 1.06 | 0.99 | 1.01 | 1.02 | 0.49 | 2.43 | 0.995 | 0.960 | 0.988 |

At the best exponent (p 1.7), with the weights fixed at the measured set (spell power 0.86), the archetype factors are caster 1.001 (90% 0.998-1.003), hybrid 1.001 (0.998-1.006) and tank 1.001 (0.996-1.005).

The tables below give each epic's value / budget - 1, as the median with the mean +/- standard error and n:

| p 1.7 | T1 head / chest / legs | T2 shoulder / hands / waist / feet | T3 neck / back / wrist / finger | all armor |
|---|---|---|---|---|
| caster | -0.8% (-0.8 +/- 0.4, 30) | -0.3% (-0.2 +/- 0.3, 48) | -0.3% (+0.1 +/- 0.3, 69) | -0.3% (147) |
| physical | -0.2% (-0.1 +/- 0.5, 21) | +0.8% (+0.6 +/- 0.3, 37) | +0.0% (+0.3 +/- 0.3, 57) | +0.2% (115) |
| hybrid | -0.0% (+0.5 +/- 0.6, 17) | -0.5% (-0.7 +/- 0.4, 36) | -1.0% (+0.3 +/- 0.8, 9) | -0.3% (62) |
| tank | -1.3% (3) | +1.2% (8) | -0.3% (13) | -0.0% (24) |

| p 1.5 | T1 | T2 | T3 | all armor |
|---|---|---|---|---|
| caster | -1.3% (-1.0 +/- 0.4, 30) | -1.3% (-0.9 +/- 0.4, 48) | -0.3% (+0.3 +/- 0.4, 72) | -1.0% (150) |
| physical | -1.8% (-1.7 +/- 0.6, 22) | -0.2% (-0.2 +/- 0.4, 37) | -1.2% (-0.8 +/- 0.4, 57) | -1.1% (116) |
| hybrid | +3.5% (+3.3 +/- 0.7, 17) | +2.0% (+1.8 +/- 0.5, 36) | +2.2% (+3.7 +/- 1.0, 9) | +2.5% (62) |
| tank | -1.6% (3) | +0.5% (8) | -0.5% (13) | -0.1% (24) |

- **Residual by stat count:**
  - p 1.5: 2 stats -2.6%, 3 stats -0.7%, 4 stats +0.8%, 5 stats +2.5%, 6 stats +3.7%, 7 stats +5.5%.
  - p 1.7: flat, between -0.4% and +0.7%.
  - The p 1.5 "hybrid premium" is this slope: hybrids carry 5.3 stats. Healer pieces (-0.9%), caster-dps pieces (-0.1%) and hunter-type physical + Int pieces (+0.3%) sit on the line too.
- **Phase 1 subsets (p 1.5 model):** MC / Onyxia / world boss epics are caster -0.8% (n=23) and physical -1.6% (12). Pre-raid dungeon items are caster -0.3% (87) and physical -1.6% (70).
- **Spell power sensitivity (p 1.5):** the caster-dps residual is -4.7% at sp 0.55, -0.2% at 0.86 and +4.4% at 1.0. The fit is best at 0.84-0.86 (rms 3.43-3.47%, against 4.67% at 0.55).
- **Verdict:**
  - Era priced archetypes identically; only the stat count differs.
  - The Era level 60 exponent is 1.7, not 1.5. At ilvl <= 30, 1.5 and 1.75 could not be told apart.
  - Era epic budget at p 1.7: V = 0.661 ilvl + 2.92 (46.6 / 49.9 / 53.2 at 66 / 71 / 76).

### SoD phase 4-6 epics (609 clean armor pieces)

- **The fit is weaker:** the rms is 9-11% against Era's 2.7-3.5%. SoD's item levels track budgets loosely; phase 5 pieces sit about 8% above the phase 4 and 6 lines.
- **SoD's own weights (p 1.5, free fit):**
  - spell power 1.04 (1.01-1.06);
  - crit 1.23 and hit 1.21 per rating point;
  - MP5 3.8, AP 0.57, healing-only power 0.55, school power 0.82.
  - SoD pays more per point of spell power, hit and crit than Era.
- **Budget:** at Era's weights SoD's budget is about 12% above Era's at ilvl 66 (V = 0.564 ilvl + 20.2).
- **Archetype factors against physical** (V per phase, weights free; p 1.5 / p 1.7):

| | p 1.5 | p 1.7 |
|---|---|---|
| caster | 0.972 (0.953-0.997) | 0.962 (0.941-0.990) |
| hybrid | 0.997 (0.965-1.022) | 0.979 (0.941-0.999) |
| tank | 0.956 | 0.929 |

  With Era's weights the caster factor is 0.935.
- **Residuals at Era's weights:**
  - caster dps -2.9% (n=323), physical -1.7% (121), hybrid +0.1% (13), tank +5.0% (79);
  - physical + Int +10.9% (63): SoD's paladin, shaman and druid mail and leather with 4.7 stats.
- **Verdict:** SoD gives caster pieces slightly less than physical ones (3-6%), but that is inside its own scatter. Forever matches Era (forever_item_budget.md section 5), so Era is the reference.

### Forever ilvl 5-30 (armor; spell power 0.86, healing 0.40, MP5 2.0)

| exponent | caster (vs physical) | hybrid (vs physical) | hybrid, Forever-new ids only |
|---|---|---|---|
| p 1.5 | 1.009 (1.000-1.017) | 1.036 (0.997-1.061) | 1.018 |
| p 1.7 | 1.005 | 1.015 (0.985-1.041) | 1.028 (0.98-1.07) |

There is no significant archetype premium.

## Q2. Stats per piece and budget split

Stat lines per armor piece: healing on top of spell power counts as one line, and AP + RAP as one. The headline columns are Era Phase 1 raid epics (MC, Onyxia, Azuregos, Kazzak) and all Era epics:

| archetype | Era P1 raid epics: mean (T1 / T2 / T3) | all Era epics 55-80: mean (T1 / T2 / T3) | usual range | Era pre-raid ilvl 60-66 | SoD P4-6 epics (T1 / T2 / T3) |
|---|---|---|---|---|---|
| caster dps | 4.1 (4.3 / 4.3 / 3.6), n=46 | 3.8 (4.0 / 3.9 / 3.4), n=203 | 3-5 | 3.3 | 3.9 (4.2 / 3.9 / 3.4) |
| healer | 4.5 (4.6 / 4.5 / 4.0), n=27 | 4.2 (4.5 / 4.3 / 3.7) | 4-5 | 3.4 | 3.7 |
| physical | 3.1 (4.0 / 3.3 / 2.4), n=24 | 3.0 (3.3 / 3.0 / 2.8), n=169 | 2-4 | 2.4 | 3.4 (3.7 / 3.5 / 3.0) |
| physical + Int (hunter-type) | 4.5 (5.0 / 4.0 / -), n=12 | 4.1 (4.4 / 4.1 / 3.6), n=72 | 3-5 | 4.0 | 4.6 (4.9 / 4.4 / 3.8) |
| hybrid | 5.4 (5.6 / 5.2 / -), n=10 | 5.4 (5.7 / 5.3 / 4.8), n=94 | 4-7 | 4.0 (n=3) | 4.3 (4.9 / 3.8 / 4.0) |
| tank | 3.3, n=16 | 3.4 | 3-4 | 2.9 | 3.9 |

Distribution in all Era epics:
- caster dps: 27% have 3 stats, 45% have 4 and 21% have 5;
- physical: 25% have 2, 49% have 3 and 25% have 4;
- hybrid: 16% have 4, 36% have 5, 36% have 6 and 11% have 7.

The full per-tier distributions are in bis_archetype.py's output.

Budget split of armor pieces, as linear shares (%) at the measured weights:

| archetype | primaries (Str / Agi / Int / Spi) | Stamina | spell power | hit | crit | AP | MP5 | tank |
|---|---|---|---|---|---|---|---|---|
| caster dps, Era P1 raid | 39.4 | 25.0 | 26.3 | 0.8 | 5.8 | - | 2.8 | - |
| caster dps, all Era epics | 31.5 | 27.2 | 30.5 | 2.0 | 5.5 | - | 3.2 | - |
| caster dps, SoD P4-6 | 27.1 | 25.9 | 33.0 | 2.7 | 10.5 | - | 0.8 | - |
| healer, all Era epics | 43.2 | 20.6 | 24.9 | - | 3.0 | - | 8.3 | - |
| physical, Era P1 raid | 41.7 | 30.2 | - | 7.1 | 13.9 | 7.1 | - | - |
| physical, all Era epics | 44.0 | 32.6 | - | 7.1 | 10.0 | 6.3 | - | - |
| physical, SoD P4-6 | 44.8 | 27.2 | - | 9.2 | 11.9 | 6.8 | - | - |
| physical + Int, all Era epics | 55.7 | 28.3 | - | 2.6 | 9.0 | 4.4 | - | - |
| physical + Int, SoD P4-6 | 49.5 | 27.5 | - | 4.3 | 13.1 | 5.5 | - | - |
| hybrid, all Era epics | 49.7 | 22.2 | 16.0 | 0.2 | 6.1 | 0.8 | 5.0 | - |
| hybrid, SoD P4-6 | 37.6 | 21.5 | 18.8 | 5.4 | 16.6 | - | - | - |
| tank, all Era epics | 31.8 | 39.8 | - | 0.7 | 1.8 | 1.0 | - | 24.8 |

The Stamina share by archetype is in the Stamina column. All pieces, BiS or not, carry 21-33% Stamina; tanks carry 40%. BiS sets pick lower-Stamina pieces, which is why SoD BiS sets average 13-23% (the per-spec Stamina ruling stands).

## Q3. Special rules

### Weapons: slot modifiers per archetype (chest = 1, fitted together with the armor)

| | Era level 60, p 1.7 (n) | Era, p 1.5 | SoD P4-6, p 1.5 | Forever ilvl 5-30, p 1.5 |
|---|---|---|---|---|
| two-hand, physical | 1.00 (46) | 0.97 | 1.19 (16) | 1.00 (71) |
| two-hand, caster (staff) | 1.04 (28) | 1.03 | 1.42 (11) | **2.08 (26)** |
| two-hand, hybrid | 0.98 (4) | 0.99 | 1.14 (4) | 1.17 (3) |
| one-hand, physical | 0.43 (77) | 0.41 | 0.41 (39) | 0.42 (104) |
| one-hand, caster | 0.63 (12) | 0.61 | 0.89 (22) | **1.78 (16)** |
| held off-hand, caster | 0.57 (24) | 0.56 | 0.50 (13) | 0.52-0.56 (22) |
| shield, tank | 0.59 (16) | 0.58 | 0.53 (9) | 0.52 (12) |
| ranged (bow / gun / crossbow) | 0.32 (33) | 0.31 | 0.29 (9) | 0.31 (38) |
| wand | 0.32 (17) | 0.30 | 0.26 (8) | 0.26-0.32 (26) |

- **Caster weapons:**
  - Era priced a caster one-hander at 1.47 times a physical one, and a staff like a physical two-hander (1.04).
  - Forever gives caster weapons about one extra chest budget of spell power. Examples:
    - Ceranium Rod went from Stamina 7 to Stamina 7 + 14 spell power at the same ilvl;
    - Rod of the Sleepwalker carries 32 spell power at ilvl 29;
    - 32 Classic items gained spell power with nothing removed: 14 two-handers, 11 wands and 1 main hand among them.
  - In linear stats a Forever staff carries 2.44 times a physical two-hander, and a caster one-hander 4.67 times a physical one-hander. Against Forever's own line, a caster main hand + held off-hand (1.73 + 0.58) carries the same linear stats as a staff (2.31).
  - The pattern is flat from ilvl 8 to 29.
  - The listing's ilvl 31-70 caster weapons (not evidence under the ilvl <= 30 ruling) show the same 1.7-1.8 (one-hand) and 2.1-2.2 (two-hand) on Classic ids Forever rebudgeted. For example, Runesword of the Red went from 64 to 104 spell power and High Warlord's War Staff from 71 to 106.
  - SoD did half of this (one-hander 2.2 times physical, staff 1.2 times).
- **Two-hand vs one-hand:** physical two-hander / one-hander = 2.33 (Era p 1.7), 2.36 (p 1.5), 2.4 (Forever). Staff / caster one-hander = 1.66 (Era) and 1.17 (Forever).
- **Ranged slot:**
  - Wands cost the same as bows and guns: 0.97-0.98 of them in Era.
  - Forever wands sit at 0.26-0.32, like Era's.
  - Relics (idols, totems, librams) carry no stats in Era or SoD (3 Era epic relics, all effects), so they stay worth their slot through their effect, as today.
- **Healing vs damage:** a healing-only point costs 0.53-0.54 of a damage-and-healing point in Era (0.45 / 0.85), 0.53 in SoD (0.55 / 1.04) and 0.47 in Forever ilvl <= 30 (0.40 / 0.86). Healer pieces sit on the same line as caster-dps pieces (-0.9% against -0.1% at p 1.7).
- **Hit and crit, caster vs physical:**
  - Era p 1.7: melee / ranged hit 1.05 (1.01-1.09), spell hit 1.05 (1.00-1.10), crit 1.01 (1.00-1.02), spell crit 1.01 (0.99-1.02).
  - At p 1.5: hit 1.10 / spell hit 0.98, crit 1.02 / spell crit 0.99.
  - Caster and physical items price hit and crit the same; 1.0 for both is within 5%. SoD prices both at about 1.2-1.3.
- **Stamina share by archetype:** see Q2.
- **SoD's hybrid DPS pieces (phases 4-6):**
  - Only 36 of 1,491 pieces carry physical and spell stats, almost all of them paladin.
  - Ret (Lawbringer, Judgement, Avenger's, 27 pieces) carries Str + Stamina + **Holy** spell power + hit / crit at 4.4 stats per piece. Its split is primaries 36%, Stamina 22%, spell power 19%, hit 6%, crit 17%.
  - Enhancement (Earthfury, Ten Storms, 29 pieces) carries Str (+ Agi) + Stamina + **Int, no spell power** (0.2%) + hit / crit at 4.6 stats. Its split is primaries 54%, Stamina 25%, hit 7%, crit 13%.
  - Feral (Cenarion, Stormrage, 38 pieces) carries Str + Agi + Stamina + Int + hit / crit at 4.7 stats, with no spell power.
  - In p-norm value SoD hybrids sit on the common line (+0.1%, n=13). In linear stats they carry 1.08 of SoD's line, against physical 0.94 (1.15 times).
- **Forever ilvl 5-30 hybrid pieces:**
  - 36 of 1,164 armor pieces with stats (27 Forever-new: Crusader's, Totemic, Sterling Silver, Kodohide...) carry 2.9 stats against 2.0 for casters and physical.
  - In p-norm value they sit on the line (hybrid / physical 1.02-1.04, interval includes 1).
  - In linear stats they carry 1.11 of Forever's line (Forever-new 1.14), against physical 0.98-0.99: 1.13-1.15 times.
  - Their spell power share is high (Crusader's Chain Helm: Str 11, SP 14).
- **Forever's stat-to-spell-power swaps:**
  - There are 36 single swaps (one primary removed, spell power added, nothing else changed).
  - Of the 14 swaps of 4+ points, 9 are point for point (rate 1.0): Hillman's 4 to 4 and 5 to 5, Rose Mantle 5 to 5, Spidersilk Boots 7 to 7, Robes of Arcana 8 to 8.
  - The other 5 are at 0.80-0.88: Minor Channeling Ring 4 to 5, Windsong Drape 4 to 5, High Robe of the Adjudicator and Barkshell Tunic 7 to 8, Warsong Sash 9 to 11. None is under 0.8.
  - Rounding allows 0.86 on 25 of 36 and 1.0 on 30 of 36.
  - So Forever converted old items at about 1:1, while the regression prices spell power at 0.86 over all ilvl <= 30 armor: 0.93 on Classic ids and 0.80 on Forever-new items, which are designed from scratch. Era level 60 independently gives 0.84-0.85.
  - Verdict: 0.86 (range 0.80-1.0) is the best single price. The swaps are the generous end.

## Q4. Trinkets

### Passive stat trinkets

21 stat-only trinkets (Era blues and epics at ilvl 55+, SoD phases 4-6):

| item (ilvl) | stats | value / V, p 1.7 | linear / L(ilvl) |
|---|---|---|---|
| Briarwood Reed (q3, 60) | 29 sp | 0.76 | 0.36 |
| Blackhand's Breadth (q3, 63) | 2% crit | 0.82 | 0.39 |
| Shard of the Scale (q4, 71) | 16 MP5 | 0.77 | 0.47 |
| Drake Fang Talisman (q4, 75) | 56 AP, 2% hit, 1% dodge | 0.80 | 0.73 |
| Neltharion's Tear (q4, 83) | 44 sp, 2% spell hit | 0.80 | 0.63 |
| Tear of the Dreamer (SoD, 80) | 46 sp, 1% hit | 0.75 | 0.55 |
| Ruby-Encrusted Broach (SoD, 80) | 18 Int, 23 sp, 1% crit | 0.60 | 0.56 |

- Median trinket modifier: 0.70 (p 1.5) and 0.74 (p 1.7) of a chest, above jewelry's 0.55, because trinkets carry 1-3 stats.
- In linear stats, epic stat trinkets carry 0.56 x L(ilvl) (median, n=15, range 0.36-0.73), the same as a ring.

### On-use trinkets

Average weighted points over a 180 s fight, used on cooldown from time 0, against the passive budget of a stat trinket of the same ilvl (0.74 x V, p 1.7):

| item (ilvl) | on-use | in-fight uptime | average / passive budget |
|---|---|---|---|
| Talisman of Ephemeral Power (66; SoD 71) | 175 sp, 15 s / 90 s (SoD 184) | 17% | 0.73 (SoD 0.72) |
| Zandalarian Hero Charm (68) | 204 sp decaying (about x0.67), 20 s / 2 min | 22% | 0.74 |
| Draconic Infused Emblem (q3, 63) | 100 sp, 15 s / 75 s | 25% | 0.84 |
| Hibernation Crystal (71) | 350 healing, 15 s / 90 s | 17% | 0.71 |
| Earthstrike (66; SoD 75) | 280 AP, 20 s / 2 min (SoD 328) | 22% | 0.94 (SoD 0.98) |
| Jom Gabbar (81) | 65 AP growing by 65 every 2 s (about x5.5), 20 s / 2 min | 22% | 0.99 (SoD 1.07) |
| Nat Pagle's Broken Reel (q3, 68) | 10% spell hit, 15 s / 75 s | 25% | 0.98 |
| Slayer's Crest (90) | 64 AP passive + 260 AP, 20 s / 2 min | 22% | 0.72 + 0.65 |
| Restrained Essence of Sapphiron (90) | 40 sp passive + 130 sp | 22% | 0.75 + 0.54 |

- Pure on-use epics: median 0.74 of the passive budget (n=9, range 0.70-1.07); spell power 0.73 (n=4), attack power 0.98 (n=4).
- Passive + on-use trinkets (Naxxramas-era): 1.19 in total (n=5).
- By archetype:
  - Phase 1 casters get on-use spell power (Talisman of Ephemeral Power) or passive spell power / spell crit (Briarwood Reed, Eye of the Beast).
  - Physical specs get on-use attack power (Earthstrike, Devilsaur Eye) or passive crit / attack power (Blackhand's Breadth, Hand of Justice's 20 AP + proc).
  - No real trinket is hybrid.
- A synthetic on-use trinket at ilvl 71 should average 0.74 (spell power) to 1.0 (AP / Str) of 0.74 x V(71) = 36.8 weighted points over a 180 s fight:
  - spell power: about 27 points, e.g. 190 sp for 15 s on a 90 s cooldown;
  - AP: about 37 points, e.g. 330 AP for 20 s on a 2 min cooldown.

## Q5. Proposal: replacing standards 1a

### Which rule matches real epics?

rms log error over the 355 clean Era epic armor pieces, for all pieces / without 3-sigma outliers:

| rule | rms |
|---|---|
| one linear total for every piece (1a style) | 11.9% / 11.7% |
| linear total x archetype-and-tier factor | 8.2% / 7.3% |
| p-norm, p 1.7, measured weights | 2.8% / 2.2% |

The p-norm is the real rule. Its linear consequence is set by the stat count: the linear / value ratio is n^0.41 at p 1.7 (1.32 / 1.51 / 1.72 / 1.88 / 2.07 for 2-6 stats), and real epics reach 97% of the even-split maximum. So fixing each archetype's stat count fixes its linear total. The procedure uses both rules: linear totals per archetype for comparability, and the p-norm as the shape check.

### Procedure

1. **Weights:** spell power 0.86 (generic), healing on top 0.45, school-only power 0.70, AP 0.5, RAP 0.4, crit and hit rating 1.0 (14 / 10 per 1%; caster and physical alike), MP5 2.4, primaries 1. Slot mods: 1 / 0.75 / 0.5625.
2. **Linear total per armor piece:** slot mod x (1.125 ilvl + 2.0) x f, within +/-1%. f is the median real Era epic of the archetype and slot tier (90% interval of the median):

| f | T1 head / chest / legs | T2 shoulder / hands / waist / feet | T3 neck / back / wrist / rings |
|---|---|---|---|
| pure caster (caster dps, 12 builds) | 1.053 (1.01-1.08) | 1.026 (0.99-1.05) | 0.995 (0.97-1.01) |
| physical melee (warriors, rogues, 9 builds) | 0.950 (0.92-1.00) | 0.959 (0.92-0.98) | 0.911 (0.90-0.92) |
| hunter (physical + Int, 3 builds) | 1.028 (0.98-1.04) | 1.056 (1.01-1.13) | 0.977 (0.88-0.99) |
| hybrid Ret / Enhancement (2 builds) | 1.230 (1.17-1.27) | 1.163 (1.14-1.23) | 1.131 (1.09-1.16) |
| feral (1 build; physical) | 0.950 | 0.959 | 0.911 |

3. **Shape check:** each piece's 1.7-norm value at the measured weights must be within +/-6% of slot mod x (0.661 ilvl + 2.92). Real epics scatter 2.7% around it. Together with step 2 this pins the stat count to the archetype's real one. For example, a physical T1 piece passes with 3 stats (-1%), but fails with 2 (+17%) or 4 (-12%). Stat lines per piece that pass:

| archetype | T1 | T2 | T3 |
|---|---|---|---|
| caster | 4 | 4 | 3-4 |
| physical | 3 | 3 | 3 |
| hunter | 4 | 4 | 3 |
| hybrid | 5-6 | 5 | 4-5 |

   These match the real averages (3.8, 3.1, 4.1 and 5.3). A lopsided split (one dominant stat) raises the value by a few percent, so it can bring a borderline count inside the window. The set-level 5a mix is reached across the 12 pieces.
4. **Weapons**, as multiples of L(ilvl):
   - physical two-hander at 76: 0.90;
   - physical one-hander at 71: 0.39 each;
   - hunter melee two-hander: 0.98 (physical x 1.08);
   - Ret / Enhancement two-hander: 1.14 (physical x 1.26, the hybrid / physical armor ratio);
   - feral two-hander: 0.90;
   - caster staff at 76: 2.21 (physical x Forever's 2.44), or caster main hand at 71: 1.82 (physical x 4.67) plus held off-hand at 71: 0.53. Both total about 193;
   - physical ranged weapon (bow, gun, thrown) at 71: 0.24;
   - wand: 0.26;
   - relic, idol, totem or libram: no stats (its effect is its value, unchanged).
5. **Trinkets:** two at ilvl 71. A stat trinket has 1-3 stats, a linear total of 0.56 x L(71) = 45.9 and a 1.7-norm value of 0.74 x V(71) = 36.8. An on-use trinket averages (amount x weight x in-fight uptime) 27 weighted points for spell power and 37 for AP / Str, the sizes of Talisman of Ephemeral Power and Earthstrike.

### Table

Expected weighted linear total per piece (weights as in step 1). The "today" column is standards 1a in its own weights, with spell power at 0.55. Today's sets re-priced at spell power 0.86 come out x1.159 for casters (5a mean spell power share 28.2%) and x1.119 for Ret / Enhancement (about 21%).

| piece | today 1a | pure caster | physical melee | hunter | hybrid Ret / Enh | feral |
|---|---|---|---|---|---|---|
| T1, ilvl 66 (head, chest, legs) | 74.0 | 80.3 | 72.4 | 78.4 | 93.8 | 72.4 |
| T2, ilvl 66 (shoulder, hands, feet) | 57.0 | 58.7 | 54.9 | 60.4 | 66.5 | 54.9 |
| waist, ilvl 71 | 65.0 | 63.0 | 58.9 | 64.8 | 71.4 | 58.9 |
| T3, ilvl 71 (neck, back, wrist, two rings) | 47.5 | 45.8 | 42.0 | 45.0 | 52.1 | 42.0 |
| **armor, 12 pieces** | **695.6** | **709.0** | **650.6** | **706.1** | **812.7** | **650.6** |
| two-hander, ilvl 76 | 94.8 | staff 193.1 | 79.0 | 85.4 (melee) | 99.6 | 79.0 |
| one-hander, ilvl 71 | 35.5 (physical); caster slots 126.3 | main hand 149.3 + off-hand 43.5 | 31.9 each | - | - | - |
| ranged slot, ilvl 71 | 27.0 | wand 21.3, or relic | 19.9 | bow 19.9 | libram / totem | idol |
| trinket, ilvl 71 (x2) | none | 45.9 each | 45.9 each | 45.9 each | 45.9 each | 45.9 each |

Set totals (armor + weapons + ranged, no trinkets):

| archetype | proposal | today (1a weights) | today re-priced at sp 0.86 | change | with two trinkets |
|---|---|---|---|---|---|
| pure caster, staff + wand | 923.4 | 848.9 | 983.9 | **-6.1%** | +3.2% |
| pure caster, staff + relic (Balance, Elemental) | 902.1 | 821.9 | 952.6 | **-5.3%** | +4.3% |
| pure caster, main hand + off-hand + wand | 923.1 | 848.9 | 983.9 | -6.2% | +3.1% |
| physical melee, two-hander (Arms, Fury 2H) | 749.5 | 817.4 | 817.4 | **-8.3%** | +2.9% |
| physical melee, dual wield (Fury, rogues) | 734.3 | 793.6 | 793.6 | **-7.5%** | +4.1% |
| hunter, two-hander + bow | 811.5 | 817.4 | 817.4 | **-0.7%** | +10.5% |
| hybrid Ret / Enhancement, two-hander + relic | 912.4 | 790.4 | 884.9 | **+3.1%** | +13.5% |
| feral, two-hander + idol | 729.5 | 790.4 | 790.4 | **-7.7%** | +3.9% |

What it means for stat amounts:
- **Casters:** with the same 5a proportions, armor + staff spell power goes from about 421 to about 399 (-5%). Pricing spell power at 0.86 instead of 0.55 takes more than the doubled caster weapon gives back.
- **Physical specs:** they lose about 8% of every stat, because real physical epics carry 3.1 stats and sit 5-9% under the average line.
- **Hunters and hybrids:** they gain relative to everyone else, because their real pieces carry more stats.

Under the proposal (no trinkets), caster sets carry 20-26% more linear points than physical melee sets (902-923 against 734-750). Most of that is the staff, as in Forever's own low-level gear. Today, re-priced at spell power 0.86, the gap is 17-24%.

### Uncertainty

| source | effect |
|---|---|
| The budget line L | -5% to +3% at ilvl 66, -7% to +3% at 71, -10% to +3% at 76 (forever_item_budget.md section 6). It moves every archetype together. The ruling's line was fitted with a 0.77 middle-tier mod: with 0.75 the same pieces give 1.158 ilvl + 0.58 (+1% at 66). The f values are measured against the line as written, so they absorb this. |
| f per archetype and tier | +/-2-4%: the 90% intervals of the medians above (n = 9-49 per cell). Phase 1 raid epics alone agree within their scatter (caster 1.05, physical 0.90, n = 15 / 9). |
| Hybrid stat count | The central f (1.13-1.23) assumes Forever's level 60 Ret / Enh pieces carry 5-6 stats, like Era hybrids and our 5b sets. If Forever itemizes them like SoD (4-4.5 stats), hybrid / physical is 1.15 instead of 1.26 and the hybrid armor drops about 8% (to about 745). |
| Feral | Physical (0.92-0.95) as Era Phase 1 BiS. If Forever gives feral pieces Int like SoD (4.7 stats), f is the hunter's (+8%). |
| Caster weapons | Forever's ilvl <= 30 rule is well supported (26 staves and 16 one-handers, flat across ilvl, and the same multiple on rebudgeted ilvl 31-70 listing items). If Forever's level 60 tuning reverted to Era's (staff 1.35 x L), the caster weapon would be 118 instead of 193 (-8% of the set). |
| Spell power weight | 0.86 (0.80-1.0). At 0.80 casters get about 7% more spell power for the same points, and at 1.0 about 14% less. |
| Trinkets | Stat trinkets 0.36-0.73 x L(71) (n=15). On-use 0.70-1.07 of the passive budget (n=9). |
| Exponent | p 1.7 (Era level 60 best; 1.5-1.75 at ilvl <= 30). It changes only the shape check (step 3), not the table: the f values are measured linear ratios. |

Overall, the per-archetype set totals carry about +/-5% before the shared line uncertainty. The two largest single assumptions are the caster weapon rule and the hybrid stat count, and both are named in the table.
