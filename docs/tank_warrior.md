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
| Revenge (11601, rank 5) | 121 ± 19.7%, 5 Rage, 5 s, GCD, Defensive Stance; tooltip's AP ratio since 70707 reads 22.5% | 97–145 + 28% AP, threat 2.25 × damage, no flat | AP: T2 log fit 0.277 ± 0.009; threat: T2 capture | ok; **call** (AP 22.5% vs 28%) |
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
| Reckless Swings (talent, no client spell; placeholder 1399001) | the notes: 15 s, 25% of base health, the next 4 melee attacks +5 Rage each, 8 s | off the GCD, landed white swings (Heroic Strike included), flat 5; **its health cost taken since 2026-11-04** (damage taken for the score; it was left out for the DPS) | notes; GCD and which swings count are the round's reading | notes; **fixed** (health cost) |
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
| Reckless Swings (Fury row 4, the notes) | above, under Abilities | 16 Fury points reach it with Shield Slam (35 Protection, no Arms) | notes |
| Piercing Howl (12323, Fury) | snare | not simmed | n/a |

Not reachable by a tank build: Gore Drinker (the notes move it to row 5 under Blood Craze, and it makes no threat: 20 Fury
points and Blood Craze before it, so with Shield Slam's 31 it needs 53 of 51 points; a tank without Shield Slam loses far
more threat than its 0.5 / 1% heals give back, so it stays unsimmed), Death Wish, Flurry (row 6, no Death Wish needed now,
but 25 Fury points), Improved Whirlwind, Mortal Strike, Sweeping Strikes and the deep Arms and Fury talents are DPS
talents the talent search may still try.

**Battleplate of Glory** (set 2103, Forever's Tier 1 Protection set, items 280905 to 280910): 2-piece +7 Defense
(1300970), 3 Intervene's cooldown −10 s (not simmed: no Intervene), 4 −1.2% dodged and parried (not simmed: the core has
no parry reduction), 5 Shield Wall −30 s. The synthetic set carries the 2-piece only (standards 6).

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
- `TestForeverTankRecklessSwings` (2026-11-04): 15 s off the GCD, 25% of base health, 4 landed swings at 5 Rage more each,
  none without the talent.
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

## Presets (2026-11-04)

Searched with `tools/rotopt -template warrior_tank -tank <preset>` on the six scoring encounters (their mean), the
synthetic Level 60 set, 600 iterations a candidate and 3,000 to confirm (Max survival 2,000 and 6,000: its paired
noise is about three times Balanced's). Talent and knob searches alternated until neither moved. The raw output is in
`orchestration/tank_handoff/warrior/` (`ts_*`, `knobs_*`, `races_*`).

| Preset | Talents | Knobs that differ from Balanced | Race | Score (Undead, confirm) |
|---|---|---|---|---|
| Max threat (w = 1/3) | `3-05050015-250430020330001351`: Improved Heroic Strike 3 in place of Deflection, Blood Craze 1, Anticipation 4, Improved Sunder Armor 3 | `tc=0` (no Thunder Clap), `hsRage=30`, `revengeFirst=1` | Undead | 4,709.0 ± 5.3 |
| Balanced (w = 1, default) | `05-05050005-250530020310001351`: Deflection 5; Cruelty, Unbridled Wrath, Enrage 5/5; Protection 31 with Shield Slam, Improved Sunder Armor 1, no Last Stand or Vanguard | (the base) | Undead | 3,795.3 ± 7.5 |
| Max survival (w = 3) | Balanced's: the talent search from it found no move past its noise (best +5.3, noise 29.4) | none (the knob search kept every knob: `demo=1` -62 ± 34, `hsRage=30` -67 ± 26) | Undead | 1,485.0 ± 19.2 (6,000) |

Balanced's knobs: `shieldBlock=1` (whenever its buff is down), `sbRage=0`, `demo=0`, `tc=1`, `tcRefresh=2`,
`debuffsFirst=0`, `revengeFirst=0`, `sunderUp=1`, `sunderRage=30`, `hsRage=50`, `bloodrageRage=30`, `opener=0`,
`chargeFrom=0`, `potion=1`, `shieldWall=2` (the last 12 s), `wallAt=12`, `trinketAt=0`, `recklessSwingsRage=0`,
`slamRage=0`. Max survival's APL is Balanced's (same talents and knobs); it gets its own file so the UI keeps three.

Races (rotopt `-eval`, the preset's talents and knobs; Max threat and Balanced at 3,000 iterations, Max survival at 6,000):

| Race | Max threat | Balanced | Max survival |
|---|---|---|---|
| Undead | **4,709.0** | **3,795.3** | **1,485.0** |
| Tauren | 4,598.1 | 3,680.6 | 1,368.9 |
| Human | 4,595.7 | 3,637.4 | 1,204.2 |
| the Skyborne (both) | 4,603.1 | 3,626.1 | 1,174.7 |
| Night Elf | 4,569.6 | 3,625.4 | 1,240.6 |
| Troll | 4,578.1 | 3,622.9 | 1,202.2 |
| Orc | 4,574.0 | 3,617.2 | 1,187.3 |
| Dwarf | 4,571.1 | 3,613.3 | 1,176.6 |
| Gnome | 4,565.0 | 3,610.8 | 1,186.9 |

Undead's lead at every weight is Touch of the Grave's drain healing the tank, modelled at its "up to 5% of max health"
cap (`sim/core/racials_forever.go`): a call (below). Without it, Tauren leads at
Balanced and Max survival, and the Skyborne and Tauren tie at Max threat.

What the searches dropped: Slam (15 s under Forever) at every weight (-150 to -380), Reckless Swings (no searched build
keeps it: a 16-point Fury start traded it for Deflection), Demoralizing Shout (the tank pays for it, `tankscore.OwnDebuffs`),
Vanguard and the Charge opener (Charge from Defensive Stance needs Vanguard; `opener=0` scores the same as the old Bloodrage-and-Charge opener), Last Stand (its health change isn't
followed by the replay, see Rotation template), Gore Drinker (unreachable with Shield Slam).

Files: `tools/tankbench/req/warrior_protection_{max_threat,balanced,max_survival}.json`,
`ui/tank_warrior/apls/level60_{max_threat,balanced,max_survival}.apl.json`, `ui/tank_warrior/presets.ts` (the three
builds, Balanced first). `tools/tankbench/req/warrior_protection.json` is left as it was (`35--251533120331221351`,
Orc, its old gear): `sim/raid_healing_test.go`, `tools/healbench`, `tools/encounters/from_logs.py` and
`TestForeverTankScore` read it, and replacing it breaks `TestForeverTankScore` (no iteration needs 500 HPS on the
synthetic set) and `TestHealingFightHealComm`. Whether the healing tools move to the Balanced request is the
orchestrator's to decide with the healing session.

## Numbers (2026-11-04)

`tools/tankbench`, 1,000 iterations, the presets' requests on all seven Anniversary encounters (Ragnaros benched, not
scored), Undead. The score at each weight (`Max threat`, `Balanced`, `Max survival` columns), SafeDPS, healing needed
(per second at 5% deaths), the death floor, TPS, DPS and DTPS. Raw output: `orchestration/tank_handoff/warrior/bench_table.txt`.

**Max threat preset**

| Encounter | w = 1/3 | w = 1 | w = 3 | SafeDPS | Healing needed | Death floor | TPS | DPS | DTPS |
|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 4,882 | 4,228 | 2,264 | 2,605 | 281 | 0.0% | 2,427 | 674 | 323 |
| Golemagg | 4,800 | 3,995 | 1,578 | 2,602 | 346 | 0.0% | 2,447 | 682 | 365 |
| Onyxia | 4,768 | 3,698 | 487 | 2,651 | 310 | 0.0% | 2,478 | 686 | 362 |
| Broodlord | 4,474 | 2,929 | -1,707 | 2,623 | 773 | 0.0% | 2,462 | 682 | 443 |
| Vaelastrasz | 5,043 | 4,534 | 3,008 | 2,649 | 605 | 0.0% | 2,403 | 653 | 608 |
| Chromaggus | 4,281 | 2,637 | -2,296 | 2,552 | 830 | 0.0% | 2,448 | 686 | 654 |
| Ragnaros | 2,889 | -1,429 | -14,385 | 2,524 | 2,051 | 0.0% | 2,393 | 667 | 1,088 |
| **Mean of six** | **4,708.0** | 3,670.2 | 555.7 | 2,613.7 | 524.2 | 0.0% | 2,444.2 | 677.2 | 459.2 |

**Balanced preset** (Max survival's request is the same build and scores the same, row for row)

| Encounter | w = 1/3 | w = 1 | w = 3 | SafeDPS | Healing needed | Death floor | TPS | DPS | DTPS |
|---|---|---|---|---|---|---|---|---|---|
| Magmadar | 4,648 | 4,229 | 2,973 | 2,429 | 180 | 0.0% | 2,266 | 629 | 253 |
| Golemagg | 4,626 | 4,097 | 2,509 | 2,445 | 227 | 0.0% | 2,291 | 637 | 280 |
| Onyxia | 4,623 | 3,809 | 1,365 | 2,515 | 236 | 0.0% | 2,345 | 648 | 313 |
| Broodlord | 4,375 | 3,214 | -269 | 2,477 | 580 | 0.0% | 2,319 | 642 | 361 |
| Vaelastrasz | 4,858 | 4,435 | 3,164 | 2,535 | 504 | 0.0% | 2,300 | 622 | 555 |
| Chromaggus | 4,253 | 2,986 | -815 | 2,443 | 640 | 0.0% | 2,355 | 658 | 507 |
| Ragnaros | 2,788 | -1,384 | -13,903 | 2,437 | 1,982 | 0.3% | 2,320 | 646 | 1,044 |
| **Mean of six** | 4,563.8 | **3,795.0** | **1,487.8** | 2,474.0 | 394.5 | 0.0% | 2,312.7 | 639.3 | 378.2 |

The other tanks, as their docs on this branch stand (Balanced, the mean of the six): the Warden shaman 501
(`docs/tank_shaman.md`), the bear 1,833.6 before its presets (`docs/tank_druid.md`); the orchestrator lines them up in
`docs/tank_sim.md` with their final numbers.

**What the update moved** (Balanced, the mean of the six): 3,360.8 before it and 3,447.0 after on the same build
(`35-05-251530120331201351`, Orc, 3,000 iterations: TPS +45% from the stance and Defiance, SafeDPS +2%), then 3,795.3 with
the Balanced preset (new talents +170, Undead +178).

## Calls for the user

1. **Touch of the Grave's drain** (Undead): modelled at its "up to 5% of max health" cap, which makes Undead the best race
   at every weight (+115 Balanced, +106 Max threat, +116 Max survival over the next race). Default: Undead. At a lower
   drain the presets would take Tauren (Balanced 3,680.6, Max survival 1,368.9; Max threat the Skyborne 4,603.1).
2. **Shield Slam's threat** (flat **508** on top of damage): the client says "very high threat" with no number. Default
   the sim's 508 (Rules audit, Shield Slam row).
3. **Revenge's attack power ratio**: the tooltip reads 22.5% since 70707; the sim keeps 28% (T2, a log fit of 0.277 ± 0.009).
4. **Thunder Clap's threat** (2.5 × damage): T3 (Classic); wowsims/forever leaves it at 1× unmeasured. Max threat drops
   Thunder Clap (`tc=0`) either way; Balanced keeps it up.
5. **Max survival is Balanced's build**: no knob or talent of the template moves it past its noise. Its real levers react to
   health (Last Stand, Shield Wall on low health), which the score can't see (brief step 2): they're in the UI for the
   player, not in the presets. The alternative, a Shield Wall timer, is in the template (`shieldWall=2`, the last 12 s).
