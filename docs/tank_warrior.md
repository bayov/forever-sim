# Protection warrior tank

The Protection warrior (`TankWarrior`: `sim/warrior/tank_warrior`, `ui/tank_warrior`) as a tank, built to the tank
workers' brief (`orchestration/brief_tanks.md`, 2026-11-03): the rules audit of every ability and talent a tank
warrior uses or can reach, what was fixed and built, the tests, the rotation template, the synthetic gear, the three
presets and their numbers, and the calls only the user can make. `docs/tank_sim.md` has the scoring (the objective,
the encounters, the bench) and its Rules audit's first pass, which this builds on.

## Rules audit

Against Forever's newest client, **1.60.1.70753** (2026-11-02, which carries the launch realms' October 30 hotfix:
`orchestration/update_ledger.md`, "Build 1.60.1.70753"; no warrior row changed in it) and **Blizzard's November 2
development notes** (`orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`, "- Warrior"), redone
2026-11-04 after the update (`orchestration/round_70753/warrior.md`). None of the notes' warrior lines are in client 70753
or a hotfix, so each is a **notes-only value** (standards 37), modelled as the notes say; the sim is Forever only
(standards 38). Read with
`tools/sources/wago.py` (SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellDuration, SpellCooldowns,
SpellAuraOptions, SpellShapeshift, SpellClassOptions, ItemSet, the trait rows of `tools/forever_talents/client/warrior.json`).
Threat is rarely in the client, so its sources are tiered: **T1** the client's own E_THREAT effect, **T2**
wowsims/forever's in-game threat captures (`tools/database/overrides/spell_overrides.go`, 2026-10-07, and the commits
`notes.md` cites), **T3** Classic's values (vmangos and the classic-warrior wiki), stated as such.

Verdicts: **ok** (the sim matches), **fixed** (wrong, fixed 2026-11-03 or, after the update, 2026-11-04), **notes**
(a notes-only value, modelled as the notes say), **call** (no source confirms the sim's value; Calls for the user),
**n/a** (not simmed, worth nothing in a boss fight's score).

### Abilities

| Ability (spell) | Client 70753 | Sim | Source, tier | Verdict |
|---|---|---|---|---|
| Defensive Stance (71, passive 7376) | −10% damage taken, −10% dealt, +30% threat (the notes: **80%**); Defiance multiplies in (tooltip `(30+100)×(1+s1/100)`) | 0.9, 0.9, **1.8** × (1 + Defiance's 7 / 13 / 20%): ×2.16 at 3/3 with a shield | client; threat the notes (client 70753 and 70755 still 30%) | notes (master 0dcb0d87a) |
| Shield Block (2565) | 75% block, 2 charges, 7 s, 10 Rage, 5 s cd, no GCD, Defensive Stance + shield | same; a recast refreshes both charges | client; wowsims/forever | ok |
| Shield Slam (23925, rank 4) | 655 ± 4.57% + block value, 20 Rage, 6 s, GCD, shield, any stance; "very high threat", no number | 625–685 + block value, no AP; threat = damage + **508** | client; threat none | ok (damage); **call** (threat) |
| Revenge (11601, rank 5) | 121 ± 19.7%, 5 Rage, 5 s, GCD, Defensive Stance; tooltip's AP ratio since 70707 reads 22.5% | 97–145 + 25% AP (to verify), threat 2.25 × damage, no flat | AP: the 70707 tooltip (22.5% in stance = 25%), ruled 2026-11-04 over the pre-70707 T2 log fit 0.277 ± 0.009; threat: T2 capture | ok; **to verify** (a level 60 Revenge log after 70707) |
| Defensive State (5302, Revenge's window) | 5 s, 1 charge, on block, dodge or parry | 5 s, used by Revenge | client | ok |
| Sunder Armor (11597, rank 5) | −450 armor a stack, 5 stacks, 30 s, 15 Rage, GCD; E_THREAT **206** | was 261 flat; **now 206 + 5% of AP** | T1 (206) + T2 (rank 1 capture: 34 = its E_THREAT, + 0.05 × AP) | **fixed** |
| Thunder Clap (11581, rank 6) | 103, 20% slow 30 s, 20 Rage, 6 s cd, GCD, Battle or Defensive, 4 targets | 103 + 7% AP, threat 2.5 × damage | damage T2 (rank 1 log); threat T3 (Classic 2.5×; wowsims/forever leaves 1× unmeasured) | ok (damage); **call** (threat) |
| Demoralizing Shout (11556, rank 5) | −196 −1.4/level from 54 (204 at 60), 45 s, 10 Rage, GCD; E_THREAT 43 + 0.8/level (to 64) | was 43.2; **now 47** a target | T1 | **fixed** |
| Heroic Strike (11567, rank 8) | weapon + 138, 15 Rage, next swing; no threat effect | was +145 flat; **now 2.54 × 56 = 142.2** | T2 (ranks 3 and 4 captured at 2.54 a spell level) | **fixed** |
| Cleave (20569, rank 5) | weapon + 50, 2 targets, 20 Rage | was +100; **now 1.53 × 60 = 91.8**, split between the targets hit | T2 (rank 2 captured) | **fixed** |
| Bloodrage (2687, 29131) | 10 Rage + 1 a sec for 10 s, 1 min, no GCD, **costs 20% of base health** | Rage right; health cost was missing, **now taken** (counts as damage taken for the score) | client; wowsims/forever bloodrage.go | **fixed** |
| Shield Wall (871) | −60% damage 12 s, 15 min, GCD, Defensive Stance + shield | same | client | ok (health-based button: out of scoring unless on a timer) |
| Last Stand (12975, 12976) | +30% max health 20 s, 3 min, no GCD | same | client | ok (out of scoring) |
| Charge (11578, rank 3) | 15 Rage, out of combat, Battle Stance (Defensive with Vanguard), no threat for its Rage | same | client; T2 (6067c8469d) | ok |
| Battle Shout (11551, rank 6) | 111 + 0.6/level AP, 3 min, 10 Rage, raid-wide | the raid's buff; the tank doesn't cast it | client | ok (raid's) |
| Slam (11605, rank 5) | weapon + 87, 15 Rage, 1.5 s cast that resets the swing, 18 s cooldown (the notes: **15 s**; Improved Slam out of a tank's reach) | 15 s under Forever | notes | notes; in the template as `slamRage`, which the searches leave off (below) |
| Reckless Swings (1329555 cast and buff, 1329554 Rage; hotfix push 113539) | 15 s cooldown, **1 s GCD** (SpellCooldowns 102442), 25% of base health (SpellPower 315180), the next 4 melee attacks +5 Rage each (curve 128140), 8 s | the same: 15 s, 1 s GCD, 25% of base health taken (damage taken for the score), 4 landed swings at 5 Rage | client (113539 hotfix rows) | ok (the health cost was fixed 2026-11-04; master's 113539 round moved the id and the GCD) |
| Shield Bash (1672, rank 3) | 45 damage, 10 Rage, 12 s, interrupts | not simmed | client | n/a (an interrupt; 45 × 1.5 threat for 10 Rage loses to Sunder's ~400) |
| Concussion Blow (12809) | 5 s stun, 10 Rage, 45 s, no GCD, no damage | not simmed | client | n/a (bosses don't stun; Shield Slam's prerequisite) |
| Taunt (355), Mocking Blow (20560), Challenging Shout (1161) | taunts: 8 s, 2 min, 10 min | not simmed | client | n/a (one tank, one boss: nothing to taunt off) |
| Disarm (676) | 10 s, 1 min, 20 Rage | not simmed | client | n/a (raid bosses) |

### Talents a tank build can reach

Protection (all 18), Arms and Fury rows a 51-point build can reach. Values are the client's trait curves.

| Talent (spell) | Client | Sim | Verdict |
|---|---|---|---|
| Improved Bloodrage (12301) | +25 / 50% of all Bloodrage's Rage | same | ok |
| Shield Specialization (12298, 1310318) | +1% block a point, 20% a point for 5 Rage on a block | same | ok |
| Iron Will (12962) | stun and fear duration | not simmed | n/a |
| Anticipation (12297) | +4 Defense a point | same | ok |
| Improved Revenge (12797) | +20% Revenge damage a point | same | ok |
| Improved Thunder Clap (12287) | −2 Rage a point | same | ok |
| Last Stand (12975) | above | | ok |
| Master of Defense (1310316, 23602) | 50% a point for 5 Rage on a dodge or parry, with a shield | same | ok |
| Improved Disarm (12313) | Disarm's cooldown | not simmed | n/a |
| Defiance (12792) | +5% threat a point in Defensive Stance with a shield, multiplied with the stance's (the notes: **7 / 13 / 20%**) | 7 / 13 / 20%, multiplied | notes |
| Improved Sunder Armor (12308) | −1 Rage a point | same | ok |
| Vanguard (1310317) | Charge in Defensive Stance | same (an opener: the template's `opener`) | ok |
| Improved Shield Bash (12311) | Shield Bash silences | not simmed | n/a |
| Improved Shield Wall (12312) | −5.5 min a point | same | ok |
| Concussion Blow (12809) | above | | n/a |
| Focused Rage (29787) | −1 Rage a point off the class mask's spells: Shield Slam, Revenge, Sunder Armor, Thunder Clap, Demoralizing Shout, Heroic Strike, Cleave and the other attacks; not Shield Block, Bloodrage or Battle Shout | the sim's offensive flag: the same tank spells | ok (Death Wish and Sweeping Strikes are in the mask but not flagged, DPS-only) |
| Bastion (16538) | +2% Physical damage a point with a shield | same | ok |
| Shield Slam (23922) | above | | ok |
| Deflection (16462, Arms) | +1% parry a point | same | ok |
| Improved Heroic Strike (12282, Arms) | −1 Rage a point | same | ok |
| Improved Charge (12285, Arms) | +3 Rage a point to Charge | same | ok |
| Improved Tactical Mastery (12295, Arms) | +3 Rage kept a point (the notes: **3 points of 5 Rage**) | 5 a point, 3 points: the Battle Stance Charge opener keeps 10 + 5 a point | notes |
| Booming Voice (12321, Fury) | −5% shout cost a point, area | cost same; area n/a | ok |
| Cruelty (12320, Fury) | +1% crit a point | same | ok |
| Lingering Rage (1323964, Fury) | out of combat Rage decay | not simmed | n/a |
| Unbridled Wrath (12322, Fury) | 12% a point for 1 Rage on a white hit | same | ok |
| Blood Craze (16487, Fury) | 1/2/3% health over 6 s after a crit taken or a hit over 20% of health; the notes: **makes no threat** | simmed since 2026-11-02 (`docs/tank_sim.md`); its heal's threat multiplier 0 under Forever (659cad198) | notes |
| Enrage (12317, Fury) | 30% chance on any damaging hit taken for +2% Physical damage a point, 12 s; the notes move it to **row 3** (10 Fury points) | same; `TestForeverTankProcTalents` reads it at Fury's 8th place | ok |
| Reckless Swings (Fury row 4, col 2; the 10th Fury talent since 113539) | above, under Abilities | 16 Fury points reach it with Shield Slam (35 Protection, no Arms) | ok |
| Piercing Howl (12323, Fury) | snare | not simmed | n/a |

Not reachable by a tank build: Gore Drinker (the notes move it to row 5 under Blood Craze, and it makes no threat: 20 Fury
points and Blood Craze before it, so with Shield Slam's 31 it needs 53 of 51 points; a tank without Shield Slam loses far
more threat than its 0.5 / 1% heals give back, so it stays unsimmed), Death Wish, Flurry (row 6, no Death Wish needed now,
but 25 Fury points), Improved Whirlwind, Mortal Strike, Sweeping Strikes and the deep Arms and Fury talents are DPS
talents the talent search may still try.

**Battleplate of Glory** (set 2103, Forever's Tier 1 Protection set, items 280905 to 280910): 2-piece +7 Defense
(1300970), 3 Intervene's cooldown −10 s (not simmed: no Intervene), 4 −1.2% dodged and parried (not simmed: the core has
no parry reduction), 5 Shield Wall −30 s. The synthetic set carries the 2-piece only (standards 6).

### Hotfix push 113539 (2026-11-04)

Master's warrior round for push 113539 (`orchestration/round_70753/hotfix_113539_warrior.md`) re-imported the Fury tree
from the client: Raging Blows is gone (its node is now Improved Whirlwind, row 6), so every Fury talent after Dual Wield
Specialization moves up one place, and Reckless Swings is a real spell (1329555 / 1329554) on a 1 s GCD. The tank
presets spend no Fury point past Enrage (the 8th), so their strings read the same in the new tree and nothing moved:
re-run on the merged sim, Max threat 4,709.0 ± 5.3, Balanced 3,795.3 ± 7.5, Max survival 1,485.0 ± 19.2 (6,000), each
exactly as before, with byte-identical APLs, so the Numbers below stand. Remapped: `TestForeverTankRecklessSwings`'s
string (`-00000000001` to `-0000000001`) and its GCD check (none to 1 s); the template's `recklessSwingsRage` now
defaults to 0 (never), as every preset already set it. The GCD only makes Reckless Swings dearer, and no search kept it
when it was free of one, so the talent searches weren't rerun.

### What the update changed (2026-11-04)

- **Defensive Stance 80% and Defiance 7 / 13 / 20%** (the notes, on master since 0dcb0d87a): ×2.16 threat at Defiance 3/3
  with a shield, was ×1.495 (+44.5%). The tank tests now expect it.
- **Blood Craze's heal makes no threat** (the notes; the orchestrator, 659cad198). Gore Drinker isn't simmed (above).
- **The talent tree** (master's warrior round, b47131474): Arms keeps its order (Improved Tactical Mastery and Bloodthrill
  lose points), Fury's changed from its 8th talent on. The tank strings here (`35--251533120331221351`,
  `35-05-251530120331201351`) spend no Fury point past the 7th, so they read the same; the tests' Enrage string moved to
  `-00000005`.
- **Reckless Swings**, new in Fury row 4, is in a tank's reach (16 Fury points with Shield Slam): its 25% of base health is
  now taken, and the template has a line for it (`recklessSwingsRage`). **Slam's** 15 s cooldown: a `slamRage` line.
- **SafeDPS** (the orchestrator): the reference member makes 1 threat a point of damage with Salvation gone (was 0.7).
  Against it the stance's +44.5% threat is a wash: TPS +45%, SafeDPS +2%.
- **Hit to the cap** (standards 4): the synthetic set had 7.2% hit; Finger II now carries 8 hit for 8 crit, 8.0%. No
  warrior tank talent gives hit; a Dwarf (Artisan's Precision, +1%) would sit 1% over.

### What changed (2026-11-03)

- **Sunder Armor's threat** 261 → 206 + 5% of attack power (T1 + T2), about 270 at the tank's 1,300 AP: a little more.
- **Heroic Strike** +145 → +142.2, **Cleave** +100 → +91.8 (T2), **Demoralizing Shout** 43.2 → 47 (T1).
- **Bloodrage costs 20% of base health** (client): the warrior's 1,689 base health makes it 338 a cast, as damage
  taken for the survival score.
  - **To verify, base health** (2026-11-04): Bloodrage's 20% and Reckless Swings' 25% are taken from our class base
    health at the level (the emulator convention). wowsims/forever 841a787806 takes it from a base that counts base
    Stamina (Orc warrior at 60: 2629, not 1689), no evidence cited; a level 60 Bloodrage in a combat log settles it.
  - Both self costs are healing the tank needs (the tank score's replay) but not DTPS or TMI, which measure the boss's
    damage (`PayHealthCost`, sim/core/health.go; until 2026-11-04 they were in TMI; DTPS never counted them).
- **The tank keeps its own Sunders** (`tankscore.OwnDebuffs`): the bench requests' raid debuffs had a warrior tank's
  Sunder ramp (5 stacks in 5 GCDs, then forever) on top of the tank's own. A warrior tank now drops the ramp's tank
  part (other warriors' pull Sunders stay), and its rotation builds and refreshes the stacks.
- DPS goldens move TPS only (the Heroic Strike and Cleave threat); no DPS number moves, `report_hashes.txt` untouched.

### Tests

`sim/warrior/tank_warrior/abilities_test.go` (`go test -tags=with_db ./sim/warrior/tank_warrior/ -run TestForeverTank`),
next to `forever_test.go`'s Shield Wall, Shield Block, Shield Slam, sheet and score tests and dps_warrior's damage rolls:

- `TestForeverTankAbilityCosts`: Rage cost, cooldown and GCD of Shield Block, Shield Slam, Revenge, Sunder Armor,
  Thunder Clap, Demoralizing Shout, Heroic Strike and Cleave, bare and with the P1 talents (Improved Thunder Clap,
  Improved Sunder Armor, Focused Rage, Improved Heroic Strike); no Shield Slam without the talent.
- `TestForeverTankShieldBlockSpends`: a cast takes 10 Rage and gives 75% block.
- `TestForeverTankThreat`: each threat button's threat per landed hit against its source's formula in Defensive
  Stance with Defiance 3/3, over 200 hits each.
- `TestForeverTankDefensiveStance`: 0.9 / 0.9 / 1.3, Defiance only with a shield.
- `TestForeverTankDebuffs`: Thunder Clap's 20% for 30 s, Demoralizing Shout's 204 for 45 s, Sunder Armor's 5 × 450.
- `TestForeverTankRevengeWindow`, `TestForeverTankShieldNeeded`, `TestForeverTankBloodrage` (Rage with 0/1/2 points
  of Improved Bloodrage, the health cost), `TestForeverTankLastStand`, `TestForeverTankVanguard`.
- `TestForeverTankTalentStats` (Shield Specialization, Anticipation, Deflection, Cruelty, Bastion with and without a
  shield, each on against off), `TestForeverTankTalentModifiers` (Improved Revenge, Improved Shield Wall, Booming
  Voice, Focused Rage on and off its mask).
- `TestForeverTankRecklessSwings` (2026-11-04, 1 s GCD since 113539): 15 s cooldown, a 1 s GCD, 25% of base health, 4 landed swings at 5 Rage more each,
  none without the talent.
- `talents_test.go` (2026-11-04): `TestForeverTankUnbridledWrath` (60% of landed white swings at 5/5, 1 Rage each,
  never on Heroic Strike or Cleave), `TestForeverTankEnrage` (+2% Physical a point for 12 s), `TestForeverTankBloodCraze`
  (1% of max health a point over 3 ticks after a crit or a hit over 20% of max health, none after a small hit, no
  threat), `TestForeverTankHeroicStrikeDamage` (weapon + 138), `TestForeverTankShieldBlockCharges` (the boss's real
  blocks spend the 2 charges, the 75% goes with the last).
- `self_cost_test.go`: `TestForeverTankSelfCosts`, Bloodrage and Reckless Swings raise the healing needed but leave
  DTPS and TMI as they are; base health 1,689.
- `TestForeverTankProcTalents`: over 20 tanked fights, Shield Specialization 5/5 and Master of Defense 2/2 give 5 Rage
  on every block and every dodge or parry, 1/5 and 1/2 at 20% and 50% (within 4 standard errors), and Enrage procs on
  30% of the damaging hits taken.

## Rotation template

`tools/rotopt/warrior_tank.go` (`-template warrior_tank`), searched with `-tank` on the scoring encounters (the
Anniversary six without Ragnaros, their mean). Its knobs are what a warrior tank decides, and nothing in it reads the
tank's health:

- **Shield Block** whenever its buff is down, on cooldown (a recast refills both charges) or never, and from what Rage.
- **Its own debuffs**: Demoralizing Shout and Thunder Clap kept up (refreshed under so many seconds left) or not, and
  Thunder Clap on cooldown; ahead of Shield Slam and Revenge or after them. `tankscore.OwnDebuffs` takes the raid's
  Thunder Clap, Demoralizing Shout and Roar and the tank's Sunder ramp off, so the tank pays for each.
- **Shield Slam and Revenge**, either first.
- **Sunder Armor** to 5 stacks and refreshed in its last 3 sec, and as the filler above so much Rage.
- **Heroic Strike** above so much Rage, **Bloodrage** below so much (it costs health).
- **The opener**: Bloodrage before the pull, Charge (from Defensive Stance with Vanguard, or the Battle Stance dance),
  both or neither.
- **Mighty Rage Potion** at the pull, when short of Rage, or never.
- **Reckless Swings** (2026-11-04) on cooldown or below so much Rage, and **Slam** above so much Rage (never by default).
- **Shield Wall on a timer**: never, at the pull, or for the last so many seconds (it reads no health).

Last Stand stays out: it changes max health, which the survival replay doesn't follow.

## Round 4 consumables (2026-11-04, findings tank_fairness -10, -11 and -9)

The bear's consumable A/B (`orchestration/tank_handoff/druid/round3/`) run for the warrior at each weight: every
consumable slot, one change at a time against the request, paired `abtest -preset <p>` on the six scoring bosses at
the preset's own weight, screened at 3,000 iterations (seed 1), the winners decided at 10,000 on seeds 1 and 2 (disjoint).
Raw output, scripts (`gen.py`, `ab.sh`, `ab2.sh`) and an index: `orchestration/tank_handoff/warrior/round4/`.

| Change (six-boss mean, 10,000) | Max threat | Balanced | Max survival |
|---|---|---|---|
| **Flask of Natural Precision** for the Titans | +662.2 ± 3.4 / +662.5 ± 3.5 | +533.2 ± 5.1 / +536.9 ± 4.9 | +377.6 ± 11.2 / +386.4 ± 11.1 |
| **Elemental Sharpening Stone** for the Dense | +18.7 ± 0.4 / +19.0 ± 0.4 | +18.9 ± 0.9 / +21.0 ± 0.9 | +14.2 ± 2.1 / +16.6 ± 2.4 |
| Dirge's chops for the Dumpling (on the new flask and stone) | -30.4 ± 0.8 (3,000) | -6.3 ± 2.1 (3,000) | **+46.0 ± 2.5 / +43.8 ± 2.3** |
| Greater Stoneshield Potion for the Mighty Rage Potion (same) | -252 (3,000) | -183 (3,000) | **+83.4 ± 11.0 / +92.4 ± 10.4** |
| both together | | | **+118.5 ± 10.9 / +124.3 ± 10.3** |

(Seed 1 / seed 2.) The Precision flask gives 5% expertise in the Natural flasks' zones (`naturalFlaskZone`, true in every
bench encounter, as for the bear), which takes 5% off the boss's parry and dodge of the warrior's attacks; the boss's
parry also hastes its swing, so it hits the tank less often as well as the tank landing more attacks. It was never
searched for the warrior: round 3's Natural Accuracy lost to the Titans (-23 / -55 / -150) and Aggression and Swiftness
are +44 / +4 / -109 and +27 / -29 / -142 on the old base, 490 to 680 behind Precision on the new one (3,000). Everything else
holds on the new base (3,000 iterations, six-boss mean, Max threat / Balanced / Max survival): Mongoose over Greater
Agility (-50 / -49 / -48), Elixir of the Phalanx over Superior Defense (-30 / -74 / -172), Greater Fortitude (without
-26 / -58 / -131 on the old base), Juju Power over Giants (-11 / -11 / -12), Juju Might over Winterfall Firewater
(-4.5), the Mighty Rage Potion over Great Rage (-144 / -170 / -201), the Dragonbreath Chili (off: -2 +- 7 / +8 +- 10 /
+11 +- 21: a tie, kept), Tender Wolf Steak (-26 / -16 / +22 +- 12, behind Dirge's) and Grilled Squid or Blessed Sunfruit
(-25 to -31). The Dense Weightstone doesn't fit a warrior's weapon (it scores as no stone). The Zanza slot, Rumsey Rum
Black Label and the Goblin Sapper Charge came in a second pass (below, "the full raid set").

Own-weight six-boss mean, 10,000 iterations, before (round 3's requests) → after, seed 1 / seed 2:

| Preset | Seed 1 | Seed 2 |
|---|---|---|
| Max threat | 3,695.5 → **4,381.7** (+686.0 ± 3.4) | 3,687.7 → **4,373.5** (+685.9 ± 3.5) |
| Balanced | 3,010.7 → **3,567.5** (+556.9 ± 5.1) | 3,005.7 → **3,564.3** (+558.7 ± 4.9) |
| Max survival | 1,202.0 → **1,717.3** (+515.2 ± 10.6) | 1,193.8 → **1,721.2** (+527.4 ± 11.1) |

At 1,000 iterations (`summary.tsv`): 4,382.8 / 3,562.0 / 1,696.7. The talents and knobs were searched on the old
consumables and are not re-searched here: the flask moves what the boss's parry is worth, so a rotopt pass on the new
requests may move a knob or a talent.

**The UI presets match the bench** (finding -11, 2026-11-04). `ui/tank_warrior/presets.ts` had `queueDelay: 250`
(the bench's requests carry none; 250 costs 2% at Max threat), and loaded Thunder Clap, Demoralizing Shout, Demoralizing
Roar and the tank Sunder ramp, which `tankscore.OwnDebuffs` takes off for the warrior because its rotation keeps them up
(it pays their Rage and GCDs). The UI now has queueDelay 0 and none of the four, the bench's flask, stone, and for Max
survival its food and potion (`EncounterLevel60TankMaxSurvival`), and the bench's fight as a preset encounter (the Level 60
boss, a 600 HPS healer every 2 s, 180 s, the Natural flasks' zone) that the three builds load, as the bear's UI does; the
reaction time is the UI's default 200 ms, the bench's. Rotation, gear, talents, race, professions, raid buffs, party
and individual buffs were already the bench's. Nothing stripped is there for a real reason: the four debuffs are the
warrior's own to cast and the UI's rotations cast them (or leave them out, as Max threat's `tc=0` does).

### Round 4 consumables, the full raid set (2026-11-04, tank-fixes6)

Standards 11 for the tanks too: the Zanza / Blasted Lands slot, Rumsey Rum Black Label and the explosives, which the
first pass left out under the bear's "call 10" (a worker's choice, not a ruling; dropped). Shared set, all three:
Flask of Natural Precision (the Natural flasks' zone on for all six bosses, Onyxia too), Elixir of the Mongoose, Elixir
of the Phalanx, Elixir of Greater Fortitude, Juju Power, Juju Might, Elemental Sharpening Stone, Dragonbreath Chili,
**Rumsey Rum Black Label** and the **Goblin Sapper Charge**, which the three rotations throw right after the trinkets
(once a fight, its 5 min cooldown). Zanza slot: **R.O.I.D.S.** for Max threat, **Spirit of Zanza** for Balanced and Max
survival. Food and potion as before: Smoked Desert Dumpling and Mighty Rage Potion; Max survival Dirge's chops and the
Greater Stoneshield Potion. The warrior already has Engineering (with Blacksmithing), so no profession changes.

Method as the first pass (paired `abtest -preset <p>`, own weight, six scoring bosses): every option alone on the old
requests at 3,000 (seed 1), the stack and its leave-outs, then flask, food, elixirs, stone, potion and chili re-checked
on the stacked set, the close calls at 10,000 on seeds 1 and 2. Raw output, scripts and an index:
`orchestration/tank_handoff/warrior/round4/raid/` (`README.txt`, `ab_summary.txt`).

| Each item on the final set (leave it out, 10,000, seed 1 / seed 2) | Max threat | Balanced | Max survival |
|---|---|---|---|
| Goblin Sapper Charge | +181.4 / +198.5 (± 3.8) | +184.2 / +185.8 (± 5.5) | +212.3 / +211.8 (± 9.9) |
| Zanza slot (R.O.I.D.S. / Spirit / Spirit) | +64.7 / +64.9 (± 0.1) | +77.5 / +77.4 (± 1.5) | +160.3 / +155.1 (± 2.3) |
| Rumsey Rum Black Label | +10.8 / +10.6 (± 0.4) | +21.1 / +22.2 (± 0.9) | +46.4 / +48.0 (± 1.4) |

The sapper is instant and off the GCD; its 450 to 750 Fire at the pull is threat when the raid's threat ceiling is
lowest, and its 375 to 625 self-damage costs little. Every other explosive loses: Thorium Grenade, Dense and Solid
Dynamite, Goblin Land Mine, EZ-Thro Dark Bomb and Thorium Grenade and the Holy Hand Grenade (all 1 sec casts that
reset the swing) cost 270 to 430 at the pull (3,000), and a grenade or Dark Bomb after the sapper (the shared 1 min
timer: at 60 and 120 s) costs 14 to 18 (3,000), so none is taken (standards 11). The Zanza slot (10,000, seed 1 / 2, against
the pick): Max threat Ground Scorpok -18.5 / -23.0, Spirit -30.4 / -30.4; Balanced R.O.I.D.S. -13.2 / -13.0, Scorpok
-15.2 / -25.1; Max survival Scorpok -65.8 / -51.2; Lung Juice -36 to -80 (3,000). Sheen and Swiftness of Zanza, Gizzard
Gum and Cerebral Cortex do nothing for a warrior in the sim (0.0); the Darnassus Gift Collection (+30 Agility, a Love
is in the Air holiday item of an Alliance city) was screened (+50 / +73 / +126 on the old requests, under the picks)
and not taken. Alcohol: Gordok Green Grog -3.7 / -6.3 / -16.9 against the Black Label (Rumsey Rum Dark the same), Rum Light less,
Kreeg's Stout nothing. Re-checked on the final set and kept (3,000 unless noted): the Precision flask (Titans -393 to
-654, the other Natural flasks -464 to -672), Mongoose (Greater Agility -43 to -53), Phalanx (Superior Defense -29 to
-127), Greater Fortitude (none -26 to -117), Juju Power (Giants -12 to -14), Juju Might (Firewater -5.5), the
Elemental stone (none -32 to -43), Dumpling at Max threat and Balanced (Tender Wolf Steak -22.4 / -27.9 and -14.3 /
-20.1, Dirge's -14.9 / -14.2 at Balanced, 10,000), Dirge's at Max survival (Dumpling -27.7 / -27.7, Wolf Steak -24.3 /
-11.2, 10,000), the Stoneshield at Max survival (Mighty Rage -82.9 / -91.8, 10,000) and the chili (off: -8.8 / -13.5,
-20.2 / -14.3, -5.3 ± 9.6 / -3.5 ± 10.3, 10,000).

Own-weight six-boss mean, 10,000 iterations, before (the first pass's requests) → after, seed 1 / seed 2:

| Preset | Seed 1 | Seed 2 |
|---|---|---|
| Max threat | 4,381.7 → **4,636.8** (+255.3 ± 3.8) | 4,373.5 → **4,645.5** (+272.2 ± 3.8) |
| Balanced | 3,567.5 → **3,851.5** (+284.3 ± 5.3) | 3,564.3 → **3,850.5** (+286.2 ± 5.6) |
| Max survival | 1,717.3 → **2,141.2** (+423.7 ± 10.0) | 1,721.2 → **2,140.0** (+418.9 ± 10.1) |

At 1,000 iterations (`summary.tsv` rows: `round4/summary_warrior_rows.tsv`): 4,644.5 / 3,867.3 / 2,138.7. The UI
presets carry the same: `ui/tank_warrior/presets.ts` (Max threat its own consumables and encounter preset, for
R.O.I.D.S.) and the sapper action in `ui/tank_warrior/apls/level60_*.apl.json`, which are the requests' rotations. As
in the first pass, talents and knobs were not re-searched on the new consumables.

## Presets (2026-11-04, round 2; round 3 on top)

**Round 3** (2026-11-04, after the second tank audit; master fd9538c: mana-gain threat unmultiplied, tank_mechanics
-12, and disjoint tank seeds, tank_rotations -13). The warrior gains no mana; at 1,000 iterations the presets bench
3,700.2 / 3,008.5 / 1,205.0 (Max threat 3,694.8 in round 2: Shield Wall at 13 s moved it; Balanced and Max survival
as in round 2; round 4's consumables above move all three to 4,382.8 / 3,562.0 / 1,696.7, and the full raid set to 4,644.5 / 3,867.3 / 2,138.7). The same search as round 2
(talents and knobs alternated, all nine races, the gear A/Bs) moved one thing: Max threat's Shield Wall in the last
13 s instead of 12 (+4.6 ± 1.4 on seed 1, +3.5 ± 1.3 on seed 2, 10,000 iterations; 14 s +1.7 / +0.6, 11 s
-8.8 / -6.6). Every small call was checked on a second, disjoint seed (`-seed 2`). Raw output, commands and the A/B
index: `orchestration/tank_handoff/warrior/round3/` (`README.txt`, `ab_summary.txt`).

Own-weight six-boss mean, `rotopt -eval`, 10,000 iterations (± about 4 / 5 / 10), before → after:

| Preset | Seed 1 | Seed 2 |
|---|---|---|
| Max threat | 3,690.9 → **3,695.6** | 3,684.0 → **3,687.5** |
| Balanced | 3,010.6 (unchanged) | 3,005.6 |
| Max survival | 1,202.1 (unchanged) | 1,193.9 |

3×3 check (3,000 iterations, each request at each weight): Max threat 3,700.0 / 2,904.8 / 519.4, Balanced
3,622.1 / **3,012.2** / 1,182.4, Max survival 3,423.3 / 2,868.4 / **1,204.0**: each preset leads at its own weight.
Not adopted: Shield Block from 40 Rage at Max threat (+1.6 / +0.9 ± 0.8), Balanced's Sunder filler at 15 (+2.8 / +0.9),
Max survival's at 20 (+4.6 / +6.4 ± 5 to 6), the armor trinket held 40 s (-5.1 / -2.3 against 35 s), Shield Wall at 16 s
or 18 s at Balanced and Max survival (all ties or losses). The paired error held across seeds at Max threat, but at
Balanced and Max survival two seeds disagreed by more than it allows (cloak at Balanced -5.6 ± 1.8 against +0.5 ± 1.7),
so those A/Bs ran four seeds.

### Round 3 gear A/Bs

Paired, 10,000 iterations, B − A at each preset's weight, seeds 1 / 2 (/ 3 / 4):

| A/B | Max threat | Balanced | Max survival |
|---|---|---|---|
| Cloak: Superior Defense (+70 armor, 1889) vs Agility +5 (1219587) | -6.3 / -6.0 (± 1.0) | -5.6 / +0.5 / -3.4 / -5.1 (± 1.7) | -10.1 / +1.2 / +0.1 / -1.3 (± 4.4; mean -2.5 ± 2.2) |
| Wrist: Superior Stamina (1886) vs Superior Deflection (1248665) | -4.8 / -3.9 (± 1.6) | -10.4 / -0.8 / +0.3 / -7.3 (± 2.9) | -31.6 / -13.0 / -11.5 / -32.7 (± 7) |
| Shield: Lesser Block (+2 block, 863) vs Greater Stamina (+9, 1886), round 3b | -4.4 / -3.3 (± 0.9) | -6.9 / -9.8 (± 1.6) | -18.7 / -22.8 (± 4.1) |
| Armor trinket 990617 on the pull vs 991018 | -209.6 / -209.3 | -179.7 / -178.8 | (below) |
| Armor trinket held for the last 35 s vs 991018 | -192.1 / -192.2 | -146.3 / -145.4 | (below) |

Agility stays on the cloak (it wins or ties at every weight) and Superior Deflection on the wrists. The shield carries
1886 (+9) in all three requests and both gear sets; it beats Lesser Block (863, the paladin's pick) at every weight
on seeds 1 and 2 (raw output in `orchestration/tank_handoff/warrior/round3b/`). (929 is only a bracer and boots enchant in the database; the boots
keep it).

**Max survival's armor trinket** (tank_rotations-20261104-11, a call): 990617 held for the last 35 s against Balanced's
991018 on the pull, at w = 3: **+12.8 ± 5.4** (seed 1), **+15.9 ± 4.8** (seed 2). Per boss (seed 1 / 2): Broodlord
+437 / +433, Onyxia +51 / +65, Magmadar -76 / -72, Golemagg -57 / -49, Chromaggus -101 / -106, Vaelastrasz -178 / -176.
Without Broodlord (the five others' mean): **-72.0 ± 3.3 / -67.6 ± 3.7**, so 991018 on the pull would win by about 70.
The preset keeps the armor trinket (call 7).

### Round 2

Round 2 re-searched the three presets on the engine after the first tank audit (the boss faces the tank, logs' swing
gaps, Phalanx and Greater Fortitude on every tank) and the warrior fixes (Revenge 0.25 attack power, Unbridled Wrath on
auto attacks only, shield Greater Stamina +9, self health costs out of TMI). `tools/rotopt -template warrior_tank -tank
<preset>` on the six scoring encounters (their mean, Ragnaros benched), the synthetic Level 60 set, 1,000 iterations a
candidate and 3,000 to confirm; gear, Shield Wall and Sunder A/Bs with the rotations auditor's paired harness
(`tools/audit/tank_rotations/abtest`) at 3,000 iterations (10,000 where a gain stayed under twice its error). Talent and
knob searches alternated until neither moved. Raw output: `orchestration/tank_handoff/warrior/round2/`
(`ab_summary.txt` lists every A/B).

| Preset | Talents | Knobs that differ from Balanced | Gear | Race | Consumables beyond the shared set (round 4) | Score at own weight (tankbench, 1,000) |
|---|---|---|---|---|---|---|
| Max threat (w = 1/3) | `04-05050015-250330020330001351`: Deflection 4, Blood Craze 1, Anticipation 3, Improved Sunder Armor 3 | `tc=0` (no Thunder Clap), `hsRage=30`, `revengeFirst=1`, `sunderRage=15`, `wallAt=13` (round 3; was 12) | attack power on-use (991018) | Undead | R.O.I.D.S., Desert Dumpling, Mighty Rage Potion | 3,700.2 (round 3); 4,644.5 (round 4, full raid set) |
| Balanced (w = 1, default) | `05-05050005-250530020310001351`: Deflection 5; Cruelty, Unbridled Wrath, Enrage 5/5; Protection 31 with Shield Slam, Improved Sunder Armor 1, no Last Stand or Vanguard | (the base) | attack power on-use (991018) | Undead | Spirit of Zanza, Desert Dumpling, Mighty Rage Potion | 3,008.5; 3,867.3 (round 4) |
| Max survival (w = 3) | Balanced's (its own search ends there) | `trinketAt=35` (the armor trinket held for the last 35 s) | armor on-use (990617), `level60_phase1_synthetic_armor_trinket.gear.json` | Undead | Spirit of Zanza, Dirge's chops, Greater Stoneshield Potion | 1,205.0; 2,138.7 (round 4) |

All three wear Superior Deflection (Defense +9) on the wrists, Enchant Shield - Greater Stamina (spell 20017, +9 Stamina, enchant 1886, the id the bracer's Superior Stamina shares) on the shield.

Balanced's knobs: `shieldBlock=1` (whenever its buff is down), `sbRage=0`, `demo=0`, `tc=1`, `tcRefresh=2`,
`debuffsFirst=0`, `revengeFirst=0`, `sunderUp=1`, `sunderRage=30`, `hsRage=50`, `bloodrageRage=30`, `opener=0`,
`chargeFrom=0`, `potion=1`, `shieldWall=2`, `wallAt=17` (Shield Wall in the last 17 s), `trinketAt=0`,
`recklessSwingsRage=0`, `slamRage=0`.

**Shared builds** (tank_fairness-20261104-4): Balanced and Max survival end on the same talents, race and knobs; they
differ only in the trinket (and so `trinketAt`). The searches found nothing else that moves Max survival past its noise,
and nothing is forced apart.

**What moved Max threat**: with Unbridled Wrath off Heroic Strike, Improved Heroic Strike's 3 points went to Deflection
(round 1, +19) and one Anticipation point to Deflection (round 2), then the Sunder filler at 15 Rage instead of 30:
+6.9 ± 1.6 at 3,000 (tank_rotations-20261104-7; it was -0.3 ± 0.8 on the old talents, +6.7 at 20, +6.4 at 10, 15 vs 20
-0.2 ± 1.0). At Balanced and Max survival Sunder at 15 is noise (+2.8 ± 1.8, +1.6 ± 4.6), so they keep 30.

### Shield Wall timing (tank_rotations-20261104-2)

Paired, 3,000 iterations, against the last 12 s (the scoring mean; Vaelastrasz in brackets):

| Last | Max threat | Balanced | Max survival |
|---|---|---|---|
| 14 s | -0.9 ± 3.3 (-53) | +10.5 ± 4.3 (-42) | +52.4 ± 10.1 (-56) |
| 16 s | -8.4 ± 3.6 (-108) | +13.3 ± 5.6 (-114) | +78.8 ± 14.1 (-209) |
| **17 s** | | **+20.7 ± 5.4** (-112) | **+113.2 ± 13.5** (-163) |
| 18 s | -1.4 ± 4.0 (-115) | +15.3 ± 5.9 (-174) | +99.1 ± 14.9 (-330) |
| 20 s | -17.9 ± 4.4 (-198) | -10.4 ± 6.6 (-346) | +34.9 ± 16.9 (-851) |

Balanced and Max survival take 17 s, the best of the grid (round 3 kept it); rotopt's `wallAt` is now on a 1 s grid from 6 s and its knob
search on the final builds kept 17 (16 and 18 within noise of it). Vaelastrasz (a 17 ± 8 s fight) loses at every wall
time past 12, about as much at 17 s as at 16 s; the other five gain. Max threat kept 12 s in round 2 (nothing past it gained); round 3 moved it to 13 s (+4.6 / +3.5 on two seeds, 10,000).
The finding's larger gains (+46, +175 at 18 s) were on master before the fixes; on this branch they are a third and a half
of that.

### Gear A/Bs (standards 36, tank_gear-20261104-2 and -3)

Paired, on the six scoring bosses at each preset's weight:

| A/B | Max threat | Balanced | Max survival |
|---|---|---|---|
| Armor on-use 990617 (1,665 armor 20 s) on the pull vs attack power on-use 991018 (3,000) | -208.9 ± 1.2 | -176.0 ± 2.0 | -70.8 ± 5.6 |
| Armor on-use held for the last 35 s vs 991018, both with Shield Wall at 17 s (10,000) | | -133.6 ± 2.5 (its best hold, 30 s; wall at 12 s, 3,000) | **+19.0 ± 4.8** |
| The same on the final presets (3,000) | -210.4 ± 1.2 | | +21.5 ± 10.8 (vs 991018 on the pull), +106.1 ± 10.8 (vs 991018 held 35 s) |
| Superior Deflection (Defense +9, 1248665) vs Superior Stamina (+9, 1886) on the wrists (10,000) | **+4.3 ± 1.7** | **+6.6 ± 2.8** | **+25.8 ± 7.8** |
| The same on the final presets (3,000) | +8.1 ± 3.1 | +8.2 ± 5.3 | +33.9 ± 12.5 |

The armor trinket's Max survival gain is all Broodlord (+459): it loses on Magmadar (-75), Golemagg (-51), Vaelastrasz
(-165) and Chromaggus (-103), gaining Onyxia (+48). Held 30 s it gains less (-13 against the wall at 17 s), held 25 s
-75; with Shield Wall at 12 s instead, every hold (25 to 40 s) scores 50 to 135 under the attack power trinket with the wall at 17 s. Superior Deflection's formula-to-spell mapping is unverified and its
Phase 1 availability unknown (a call, below).

### Races

rotopt `-eval`, the preset's talents, knobs and gear, 3,000 iterations (`round2/races_*.txt`):

| Race | Max threat | Balanced | Max survival |
|---|---|---|---|
| Undead | **3,696.9** | **3,012.2** | **1,204.0** |
| Tauren | 3,545.1 | 2,851.8 | 973.4 |
| the Skyborne (both) | 3,546.3 | 2,817.1 | 855.6 |
| Human | 3,539.8 | 2,818.7 | 874.0 |
| Night Elf | 3,530.0 | 2,811.3 | 900.3 |
| Troll | 3,527.4 | 2,808.8 | 871.0 |
| Orc | 3,522.7 | 2,801.8 | 853.2 |
| Dwarf | 3,521.7 | 2,806.2 | 861.5 |
| Gnome | 3,518.5 | 2,801.2 | 855.2 |

Undead's lead at every weight (+150, +160, +231) is Touch of the Grave's drain healing the tank, modelled at its "up to
5% of max health" cap (`sim/core/racials_forever.go`): a call (below). Without it, Tauren leads at Balanced and Max
survival, and the Skyborne and Tauren tie at Max threat. The presets keep Undead, as before.

What the searches dropped: Slam (15 s under Forever), Reckless Swings, Demoralizing Shout (the tank pays for it,
`tankscore.OwnDebuffs`), Vanguard and the Charge opener, Last Stand (its health change isn't followed by the replay, see
Rotation template), Gore Drinker (unreachable with Shield Slam), Improved Heroic Strike (since Unbridled Wrath stopped
proccing on Heroic Strike).

Files: `tools/tankbench/req/warrior_protection_{max_threat,balanced,max_survival}.json`,
`ui/tank_warrior/apls/level60_{max_threat,balanced,max_survival}.apl.json`, `ui/tank_warrior/presets.ts` (the three
builds, Balanced first), `ui/tank_warrior/gear_sets/level60_phase1_synthetic{,_armor_trinket}.gear.json`.
`tools/tankbench/req/warrior_protection.json` is legacy (Warbringer gear, Orc, the old talents `35--251533120331221351`),
kept until the healing session decides (tank_fairness-20261104-5). It is left as it was: `sim/raid_healing_test.go`,
`tools/healbench`, `tools/encounters/from_logs.py` and `TestForeverTankScore` read it, and replacing it breaks
`TestForeverTankScore` (no iteration needs 500 HPS on the synthetic set) and `TestHealingFightHealComm`.

## Numbers (2026-11-04, round 2)

`tools/tankbench`, 1,000 iterations, the presets' requests on all seven Anniversary encounters (Ragnaros benched, not
scored), Undead. Raw output: `orchestration/tank_handoff/warrior/bench_table.txt`; the rows are in
`orchestration/audit/ref/tank_bench/summary.tsv`.

**Before and after** (the mean of the six at the preset's own weight):

| Preset | Master 5c13094 (round 1 presets) | Fixes only (318fc7e, same presets) | Re-searched | Re-search gain, paired (3,000) |
|---|---|---|---|---|
| Max threat | 3,752.2 | 3,665.3 | **3,694.8** | +45.2 ± 4.8 |
| Balanced | 3,073.7 | 2,982.8 | **3,008.5** | +28.8 ± 6.7 (Vaelastrasz -115) |
| Max survival | 1,124.7 | 1,033.2 | **1,205.0** | +164.3 ± 15.3 (Broodlord +698, Vaelastrasz -344) |

The fixes cost about 90 at every weight (unpaired, about ± 12 at Balanced): all of it SafeDPS and TPS (Balanced TPS
1,960 to 1,904), mostly Unbridled Wrath leaving Heroic Strike and Revenge's 0.25; healing needed didn't move (335.7).

**Max threat preset** (round 2's, Shield Wall at 12 s; round 3's at 13 s benches a six-boss mean of **3,700.2** at
w = 1/3, SafeDPS 2,051.0, healing needed 410.7, TPS 1,968.8: its rows are in `orchestration/audit/ref/tank_bench/summary.tsv`
and the full tables in `orchestration/tank_handoff/warrior/round3/bench/`)

| Encounter | w = 1/3 | w = 1 | w = 3 | SafeDPS | Healing needed | Death floor | TPS | DPS | DTPS |
|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 3,818 | 3,303 | 1,758 | 2,038 | 221 | 0.0% | 1,968 | 545 | 289 |
| Golemagg | 3,720 | 3,089 | 1,195 | 2,018 | 271 | 0.0% | 1,975 | 549 | 323 |
| Onyxia | 3,804 | 3,019 | 662 | 2,099 | 228 | 0.0% | 2,026 | 560 | 330 |
| Broodlord | 3,538 | 2,420 | -934 | 2,049 | 559 | 0.0% | 1,993 | 551 | 399 |
| Vaelastrasz | 4,013 | 3,595 | 2,341 | 2,111 | 498 | 0.0% | 1,959 | 530 | 570 |
| Chromaggus | 3,276 | 1,841 | -2,462 | 1,996 | 724 | 0.0% | 2,006 | 562 | 590 |
| Ragnaros | 1,977 | -1,877 | -13,439 | 1,952 | 1,831 | 0.1% | 1,944 | 541 | 1,069 |
| **Mean of six** | **3,694.8** | 2,877.8 | 426.7 | 2,051.8 | 416.8 | 0.0% | 1,987.8 | 549.5 | 416.8 |

**Balanced preset**

| Encounter | w = 1/3 | w = 1 | w = 3 | SafeDPS | Healing needed | Death floor | TPS | DPS | DTPS |
|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 3,745 | 3,447 | 2,554 | 1,947 | 128 | 0.0% | 1,866 | 517 | 234 |
| Golemagg | 3,660 | 3,244 | 1,997 | 1,934 | 179 | 0.0% | 1,880 | 523 | 264 |
| Onyxia | 3,722 | 3,143 | 1,405 | 2,006 | 168 | 0.0% | 1,925 | 530 | 286 |
| Broodlord | 3,493 | 2,683 | 251 | 1,949 | 405 | 0.0% | 1,888 | 520 | 333 |
| Vaelastrasz | 3,803 | 3,409 | 2,224 | 2,000 | 470 | 0.0% | 1,862 | 503 | 598 |
| Chromaggus | 3,302 | 2,125 | -1,405 | 1,945 | 594 | 0.0% | 1,942 | 542 | 486 |
| Ragnaros | 1,977 | -1,760 | -12,971 | 1,923 | 1,775 | 0.1% | 1,903 | 528 | 1,025 |
| **Mean of six** | 3,620.8 | **3,008.5** | 1,171.0 | 1,963.5 | 324.0 | 0.0% | 1,893.8 | 522.5 | 366.8 |

**Max survival preset**

| Encounter | w = 1/3 | w = 1 | w = 3 | SafeDPS | Healing needed | Death floor | TPS | DPS | DTPS |
|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 3,552 | 3,288 | 2,495 | 1,842 | 114 | 0.0% | 1,804 | 499 | 224 |
| Golemagg | 3,473 | 3,094 | 1,956 | 1,831 | 163 | 0.0% | 1,823 | 506 | 253 |
| Onyxia | 3,509 | 2,999 | 1,470 | 1,882 | 148 | 0.0% | 1,839 | 505 | 272 |
| Broodlord | 3,334 | 2,685 | 738 | 1,829 | 325 | 0.0% | 1,812 | 499 | 316 |
| Vaelastrasz | 3,546 | 3,175 | 2,064 | 1,866 | 441 | 0.0% | 1,729 | 464 | 579 |
| Chromaggus | 3,133 | 1,977 | -1,493 | 1,856 | 584 | 0.0% | 1,904 | 531 | 477 |
| Ragnaros | 1,785 | -1,952 | -13,163 | 1,827 | 1,775 | 0.0% | 1,857 | 515 | 1,020 |
| **Mean of six** | 3,424.5 | 2,869.7 | **1,205.0** | 1,851.0 | 295.8 | 0.0% | 1,818.5 | 500.7 | 353.5 |

Each preset scores best at its own weight among the three (Balanced at w = 3 is 1,171.0 against Max survival's 1,205.0;
at w = 1, 3,008.5 against 2,877.8 and 2,869.7).

## Calls for the user

1. **Touch of the Grave's drain** (Undead): modelled at its "up to 5% of max health" cap, which makes Undead the best race
   at every weight (+150 Max threat, +160 Balanced, +231 Max survival over the next race). Default: Undead. At a lower
   drain the presets would take Tauren (Balanced 2,851.8, Max survival 973.4; Max threat the Skyborne 3,546.3, Tauren
   3,545.1).
2. **Shield Slam's threat** (flat **508** on top of damage): the client says "very high threat" with no number. Default
   the sim's 508 (Rules audit, Shield Slam row).
3. **Revenge's attack power ratio** (ruled 2026-11-04, standards 40: **0.25, to verify**): the tooltip reads 22.5% in Defensive Stance since 70707, which is 25% before the stance's 10% less damage (wowsims/forever d7a38ae, 2026-11-04, uses 0.25 on every rank, source the tooltip, unmeasured). The sim now takes 0.25 on every rank, down from 0.28 (T2, a log fit of 0.277 ± 0.009 on ranks 1 and 2, taken before 70707). At 25% Balanced scored 3.0% less and Max survival 5.6% less than at 28% (tankbench, the presets' requests; TPS −2.7%). To verify: a level 60 Revenge log after 70707 settles it.
4. **Thunder Clap's threat** (2.5 × damage): T3 (Classic); wowsims/forever leaves it at 1× unmeasured. Max threat drops
   Thunder Clap (`tc=0`) either way; Balanced keeps it up.
5. **Base health for the self costs** (to verify): Bloodrage's 20% and Reckless Swings' 25% are taken from our class
   base (1,689 at 60, the emulator convention); wowsims/forever takes them from a base that counts base Stamina (2,629
   for an Orc), with no evidence cited. A level 60 Bloodrage in a combat log settles it (What changed, above).
6. **Superior Deflection on the wrists**: it beats Superior Stamina at every weight (round 3: +4 to +22 at 10,000 on two to four seeds), so the presets wear it,
   but its formula-to-spell mapping (1248665, client enchant 8214) is unverified and the Forever formulas may not be
   Phase 1 (tank_gear-20261104-3). If it isn't obtainable, the presets go back to Superior Stamina (1886) at the cost above.
7. **Max survival's armor trinket** (tank_rotations-20261104-11, next to the paladin's Templar's Bulwark timer,
   tank_rotations-20261104-3: both hang on Broodlord's first-ability timers, which the encounter file calls probably
   wrong). Round 3, 10,000 iterations: held for the last 35 s it wins by +12.8 ± 5.4 / +15.9 ± 4.8 (seeds 1 / 2) over
   991018 on the pull, all of it on Broodlord (+437); without Broodlord 991018 on the pull wins by about 70 (-72.0 ± 3.3
   / -67.6 ± 3.7). Default: the armor trinket, held (the current choice). If Broodlord's timers aren't trusted, Max
   survival takes 991018 on the pull and is Balanced's build outright.
8. **Shield Wall at 17 s** costs Vaelastrasz (-112 Balanced, -163 Max survival) for gains on the other five; a raid that
   weighs that fight more would keep 12 s.
9. **Max survival's levers that read health** (Last Stand, Shield Wall on low health) can't be scored (brief step 2);
   they're in the UI for the player, not in the presets.
