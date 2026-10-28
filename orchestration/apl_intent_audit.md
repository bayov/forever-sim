# APL intent audit (standards 31)

Scope: every rotation template in `tools/rotopt/` (warrior, rogue, mage, druid, hunter, paladin, priest, shaman, warlock), the
Level 60 preset knob sets (`*_presets.sh`), the report requests (`tools/dps_report/req/*.json`) and the race rotations
(`req/races/*.json`). The question for each line is whether it gets its timing from what it says it waits for, or from the
current fight length, another line's timing or line order.

Method: paired sims with `wowsimcli` from `tools/dps_report/bin`. Each variant is the preset's request rotation with one
condition rewritten. Each cell is 8 seeds x 2500 iterations; "±" is the standard error of the per-seed paired difference.
"180±20" is the report's fight-length variation and the other lengths are fixed (variation 0). A positive number means the
intent-coded line does better than the preset. Scripts and raw output are in `/tmp/claude-0/apl_audit/` (`results.txt`).
No template, preset or data file was changed.

Summary: 9 fragile lines, 7 of them confirmed by sims at more than 2 DPS at some fight length.

| Class | Fragile lines | Largest stake |
|---|---|---|
| Warrior | 4 (+1 minor) | Night Elf Fury on Battle Shout duty: +9.4 at 180±20 |
| Paladin | 1 | Holy Shock presets: +12 at 220-300 s |
| Mage | 2 | Frostfire: +10.1 at 210 s, +7.4 at 240 s |
| Shaman | 1 | Elemental: +9.2 at 300 s |
| Warlock | 1 (minor) | Troll Demonology: +2.1 at 180±20 |
| Rogue, Druid, Hunter, Priest | 0 | - |

## Warrior

| # | Line (template, knob) | Presets | Races | Now | Failure case (sim) | Intent | Proposed condition / change |
|---|---|---|---|---|---|---|---|
| W1 | Recklessness, `recklessness=3` (`warriorCooldowns`, warrior_common.go) | warrior__fury, warrior__fury-duty-battle-shout, and the Night Elf race file | all | "while Death Wish is up" | Death Wish goes at the pull once two uses fit (fights of 210 s or more: holding it would cost the second use), and Recklessness, once a fight, goes with it at the pull instead of with the fight's last Death Wish. Orc: 0.00 at 120/180/180±20, **+7.4±0.5 at 240, +6.7±0.3 at 300**. Night Elf race file: +13.9±0.4 at 240, +11.8±0.4 at 300 (this includes W2) | Recklessness is the end-of-fight dump, together with the last Death Wish | `and(auraIsActive(deathWish), eq(spellUsesRemaining(deathWish, 30s), 0))`, meaning this Death Wish is the fight's last. It is identical to now at 210 s or less. |
| W2 | Elune's Light, Night Elf race file `warrior__fury.NightElf.json` (preset knobs with `racialHoldForDw=0`) | warrior__fury (Night Elf row) | Night Elf | no condition (at the pull and on cooldown) | It misses Recklessness only because Recklessness waits for the end of a fight of 200 s or less. At 210 s or more, Recklessness goes at the pull (W1) and Elune's 10% crit lands inside its 100%. With "not during Recklessness nor 15 s before it" and Recklessness left where it is: +0.00 at 120/180, +0.13±0.01 at 180±20, **+7.4±0.4 at 240, +3.9±0.3 at 300**. With W1 fixed, the pull Elune is safe at every length (W1+W2 = W1 alone) | With the cooldowns (trinket, potion, Death Wish), never during Recklessness | `and(not(auraIsActive(recklessness)), or(eq(spellUsesRemaining(recklessness, 0), 0), not(reckDue(k, 15s))))`. Here `reckDue` is a new warrior_common.go helper built from the same knobs as the Recklessness line: recklessness=1 at the pull; =2 `le(remainingTime(), recklessAt+15)`; =3 Death Wish up, or Death Wish's own line true now, or `le(remainingTime(), dwEnd+15)`. Use it for every Elune line (W2-W4) instead of `hold(elunesLight, …)` and instead of the race file's bare line. |
| W3 | Elune's Light, `racialHoldForDw=1` (`hold(elunesLight)`), Fury Battle Shout duty | warrior__fury-duty-battle-shout | Night Elf | held for Death Wish | This is the original bug, still live. The duty variant has no race file of its own and cannot take the base's, because its rotation differs (race_override.py). So Elune goes with Death Wish and Recklessness: 13.5 of its 15 s inside 100% crit. W2's fix: **+11.5±0.3 at 120, +8.3±0.2 at 180, +9.4±0.2 at 180±20**, +4.7±0.4 at 240, +2.4±0.2 at 300. With W1 too: +10.7±0.5 at 240, +8.6±0.5 at 300 | as W2 | W2's condition in the template, so the base and duty presets both get it without a race file |
| W4 | Elune's Light, `racialHoldForDw=2` (two-hand Fury: Elune still uses `hold(elunesLight, k["racialHoldForDw"] >= 1)`) | warrior__fury_2h, warrior__fury_2h-duty-battle-shout | Night Elf | held for Death Wish (last 36 s) | Recklessness goes in the last 30 s, so Elune's Death Wish window (144-159 s at 180 s) overlaps Recklessness (150-165 s) for 9 s at every length. "Not during Recklessness nor within 15 s before it, otherwise on cooldown": **+7.9±0.5 at 120, +4.9±0.4 at 180, +6.7±0.3 at 180±20**, +3.7±0.1 at 240, +2.4±0.2 at 300. Duty: +7.3±0.3 at 180±20 | as W2 | W2's condition. With `recklessness=2` it is `gt(remainingTime(), recklessAt+15)` or Recklessness spent. |
| W5 (minor) | Elune's Light, `racialHoldForDw=0` (Arms: on cooldown) | warrior__arms, -duty, _dragonkin | Night Elf | on cooldown | The second use at 180 s overlaps Recklessness (last 25 s) only in fights of 180-205 s: +0.10±0.00 at 180±20, +0.01 at 195, 0 elsewhere | as W2 | the same helper. It is worth doing only because it comes free with W2. |

## Paladin

| # | Line | Presets | Races | Now | Failure case (sim) | Intent | Proposed |
|---|---|---|---|---|---|---|---|
| P1 | Light's Vigil and the Holy Shock low-mana branch, `vigilLongFight` (`paladinRet.Build`, `shortFight := lt(remainingTime(), vigilLongFight - now)`) | level60_shock and level60_shock_crusader_duty (210 s), level60_shock_undead (240 s) | all (paladin__level_60_shock, -duty-crusader, _shock_undead) | a 30% mana floor in fights shorter than the switch, 60% in longer ones | The switch is on the fight's total length, not on the mana the rest of the fight needs. On the shock and crusader-duty presets it loses wherever it acts. Keeping the 30% floor instead: shock **+12.8±0.2 at 220, +8.7±0.2 at 240, +11.8±0.2 at 300** (0 at 180/200); crusader duty +11.3±0.2 at 220, +12.0±0.1 at 300. On the Undead preset (Exorcism and Holy Wrath, so more mana-hungry) it helps: dropping it is -10.8±0.3 at 260, -4.2±0.1 at 300 | Vigil only while the mana lasts the rest of the fight | Replace `shortFight` with `ge(currentMana, remainingTime() × vigilManaPerSec)` (a new knob, like hunter `sniperManaPerSec`). Shock at 15 mana/s: 0.0 at 180, +12.8 at 220, +11.4 at 300, the same as dropping the switch. For Undead, 15-35 mana/s still loses 5-10 at 260 s, so that preset needs its own rate search at 120/180/240/300 (or keeps the switch until then). |

Resolved P1 (2026-10-28): `vigilLongFight`/`vigilLongMana` are gone. Shock and its Crusader duty keep the flat 30% floor: no per-second rate helped (10-15 mana/s never binds, 20+ loses at 220-300 s); paired 8x5000 +0.0 at 120/180/180±20, +12.6/+8.4/+11.5 at 220/240/300 (duty +11.1/+3.9/+11.9). Undead: what the switch bought is "the Vigil only gets the top of the bar (60%) once the mana can't keep up", and a bar-only rate failed because holding the Vigil at the pull delays the first Major Mana Potion so the third never comes. The new rule (`vigilManaPerSec=55`, `vigilTopMana=60`, `vigilHammerManaPerSec=20`, `vigilExecuteLead=10`) counts the potions and runes still to come (with 15 s left) and adds Hammer of Wrath's mana from 10 s before the execute phase: +0.46/-0.10/+4.19/+0.43/+1.44 at 120/180/210/260/300, -0.16±0.07 at 180±20 and -0.13±0.02 at 240 (20 seeds), both under the report's 0.23 standard error.

## Mage

| # | Line | Presets | Races | Now | Failure case (sim) | Intent | Proposed |
|---|---|---|---|---|---|---|---|
| M1 | Evocation `evocationMinTime=90` against the Rank 2 Frostfire Bolt line's `or(not(spellIsReady(evocation)), lt(remainingTime(), 60s))` (`lowManaRank`, mage_frostfire.go) | mage__frostfire | all (the Orc, Troll, Undead and High Order race files keep the same knobs) | Evocation below 10% mana with more than 90 s left. The cheap rank only when Evocation is on cooldown or less than 60 s are left. | The two lines encode one decision ("is Evocation still worth it?") with different constants. With 60-90 s left and Evocation ready, neither fires: full-rank bolts drain the mana and the wand finishes. The preset does worse at 240 s (1058.4) than at 300 s (1068.7). Both at 75 s: 0 at 120/150/300, +0.8±0.3 at 180±20, **+10.1±0.6 at 210, +7.4±0.2 at 240**. Evocation at 30/60 s alone loses at 180±20 (-21.7 and -3.1). | One cutoff T: with more than T left Evocation refills, with less the mana goes into rank 2 bolts | `lowManaRank`'s `lt(remainingTime(), 60s)` becomes `le(remainingTime(), seconds(k["evocationMinTime"]))`, so both lines read one knob. Re-search T (75 tested). |
| M2 | Troll Berserking through `autocastOtherCooldowns` (mage_common.go, `racialAfterScorch=0`) | mage__fire | Troll | at the pull | The pull is the Scorch ramp (5 Scorches at 1.5 s, held by the GCD), so the 10% cast speed buys nothing there. The template's `racialAfterScorch=1` line (Berserking once Fire Vulnerability has 5 stacks): **+2.5±0.2 at 120, +2.5±0.5 at 180, +1.4±0.4 at 300** | haste where casts are longer than the GCD | `racialAfterScorch=1` in the Fire preset (it adds nothing for the preset's Human), or a Troll race file. Frostfire's Troll file already does this. |

## Shaman

| # | Line | Presets | Races | Now | Failure case (sim) | Intent | Proposed |
|---|---|---|---|---|---|---|---|
| S1 | Lower-rank Lightning Bolt, `lightningBoltDownrankMana=20` (shaman_ele.go) | elemental-level60, -mana-tide | all (the Dwarf and Troll files keep it) | downrank below 20% mana | A percent floor stands in for "the mana will not last the fight". At 300 s it downranks too late: "below 25 mana per second of fight left" is **+9.2±0.2 at 300**, +0.47±0.01 at 180±20, 0 at 120, -0.5±0.1 at 240 | downrank once the mana won't last what is left | `lt(currentMana, remainingTime() × downrankManaPerSec)` (as Smite's `smiteManaRate`). Search the rate at 120/180/240/300: 25 is not yet right at 240. |

## Warlock

| # | Line | Presets | Races | Now | Failure case (sim) | Intent | Proposed |
|---|---|---|---|---|---|---|---|
| L1 (minor) | Troll Berserking through `autocastOtherCooldowns` (warlock.go) | all four warlock specs and their duty variants | Troll | at the pull | The pull is the bane, Corruption, Immolate and Conflagrate opener, mostly instants held by the GCD. Berserking with the GCD free right before a filler (as `eureka=1`) at 180±20: demonology **+2.1±0.3**, destruction_fire +0.8±0.1, affliction +0.9±0.9, destruction +0.1±0.3 | haste on the casts | a `berserking` line like `eureka=1` (`gcdIsReady()` just above the filler). It is a small gain. |

## Interactions

- Elune's Light and Recklessness: Elune's +10% crit is worthless inside Recklessness's +100%. Every warrior Elune line (Fury, Fury duty, two-hand Fury and duty, Arms) must read where Recklessness will be, so the fix is one helper (`reckDue`) derived from the Recklessness knobs, not a per-race file. The Night Elf race file's "at the pull" (W2) is right at every length only once W1 makes Recklessness wait for the fight's last Death Wish.
- Race files and duty variants: a duty variant whose rotation differs from its base takes no rotation from the base's race file (race_override.py). So the Night Elf fix on warrior__fury never reached warrior__fury-duty-battle-shout (W3, +9.4 at 180±20). The same will happen to any race fix on a base with duty variants (two-hand Fury, Arms, the warlock curse duties, the druid Faerie Fire duties, the paladin Crusader duties). This is one more reason to put intent in the template rather than in race files.
- Death Wish's wait value: `warriorCooldowns` computes `dwWait = max(spellTimeToReady(deathWish), remainingTime - dwEnd)`, which reads "207 s" at the pull of a 240 s fight in which Death Wish actually goes now (its own line gives up the hold). The racial holds survive this only because they check `auraIsActive(deathWish)` first and Death Wish sits above them in the list. A Recklessness-due helper must not reuse `dwWait` as "when Death Wish goes" (an early version of the W2 test did, and changed nothing).
- Troll Berserking on GCD-held phases: Fire's Scorch ramp (M2, +2.5), the warlock DoT opener (L1, up to +2.1). Elemental Mastery's 1.5 s bolts and the Frostfire Scorch ramp are already handled by race files. Shadow priest is unaffected: holding Berserking for Mind Flay gave -0.03±0.02 (Shadow) and -0.01±0.02 (Shadow + Holy).
- Gnome Eureka! covers only some spells: the rogue and warlock Gnome files already code it to what it covers (warlock: not Incinerate or the DoTs, hence `eureka=2` on Fire Destruction). The Arcane mage preset is a Gnome whose Eureka! goes out at the pull through the autocast. That was not tested.
- Frostfire's Evocation and rank 2 lines (M1) are two lines making one decision. Any knob that moves one (`evocationMinTime`, `lowManaRank`) must move the other.

## Lines checked and judged fine

- Warrior Death Wish held for the last `dwEnd` s (33/36): an end-of-fight dump. Held for the 20% execute phase instead, it loses 21.8±0.9 (Fury) and 29.0±0.6 (two-hand) at 120 s, 2.9±0.2 at 180±20 (Fury), and 8.3±0.3 at 300 s (two-hand).
- Warrior Recklessness at `recklessAt` (two-hand, Arms), Blood Fury and Berserking at `racialAt` (two-hand), `trinketAt`, `potionAt` plus `potionPull`: end-of-fight dumps that give up via `spellUsesLostByDelay`. All key on the end, so they stay together at any length.
- Warrior Blood Fury and Berserking held for Death Wish (`racialHoldForDw=1`): they say what they wait for and give up via `spellUsesLostByDelay`. Haste and AP stack with Recklessness.
- Warrior Battle Shout refresh at 35 s left: it refreshes at about 142 s and only when the fight outlasts the shout, so it never lands in the execute phase at any length.
- Arms `reckSwap` stance dance: tied to Recklessness's own clock.
- Rogue Blood Fury, Elune's Light and trinkets `alignedWith` Adrenaline Rush; Blade Flurry and Berserking on cooldown (stated: attack speed does not feed the energy window); Cold Blood and Eureka! before a 5 point Eviscerate; Gnome files with Eureka! on cooldown; the end-of-fight Eviscerate; Vanish and Preparation with 10 s or more left.
- Mage Arcane, no Evocation: adding one at 20% loses at every length (-1.5/-15.3/-5.1/-2.3 at 120/180/240/300).
- Mage Fire and Frost `evocationMinTime=90` (no downrank line to disagree with): 30 s loses at 180 (-12.0 Fire) and at 240/300 for Frost (-2.5/-7.4); Fire +1.8 at 300 only.
- Mage Arcane Power and Presence of Mind with the trinket; Combustion, trinket and racials after 5 Scorch stacks; the finisher instants; Frostfire's Fire Blast surplus and per-second rules.
- Druid Balance, no Innervate: Innervate at 20% or 40% loses at every length (-0.2 to -5.8). Elune's Light, potion and trinket pre-pull, later Elune held for the trinket (`holdFor`); end-of-fight Moonfire; DoT minimum times.
- Druid Feral Elune's Light outside Berserk (states the clash); Berserk held for the trinket; Shadowmeld and Ferocious Bite at the end.
- Hunter Rapid Fire and Elune's Light `alignedWith` the trinket; Sniper Shot and the weave on mana per second of fight left; Hunter's Mark refresh; Survival mana floors (percent, at 30-50%, on fillers only).
- Paladin trinket at the pull, then held for the execute phase with `spellUsesLostByDelay`; Consecration above 20% mana (a per-second rule does no better: 0 at 120/180, -1.3 at 300).
- Priest Shadow DoT minimum times, Shadowguard and the wand with 60 s or more left; Smite's downrank per second of fight left; Shadow Word: Death at the end. Smite's Devouring Plague above 80% mana: a per-second rule gains 3.4 at 120 s but loses 5-19 at 180 s, so the floor is not worse.
- Shaman Elemental Troll Berserking after Elemental Mastery (race file, stated); totem, Flame Shock and shield minimum times; end-of-fight Earth Shock. Enhancement Rage of the Farseer with the trinket, racials with Rage, Fire Nova by mana per second.
- Warlock Bane of Doom with more than 60 s left, otherwise Agony; DoT minimum times; end-of-fight instants; Gnome Eureka! lines (`eureka=1/2`) state the spells they are for.

## Unsure

- P1 on the Undead Holy Shock preset: the long-fight switch is right in direction, but no per-second rate tested reproduces it at 260 s. That preset needs a search.
- S1: the rate (25) is not right at every length (-0.5 at 240). The finding is the 300 s loss; the right rate needs a search.
- W2-W4 were simmed as JSON edits of the request rotation, not through the template. `reckDue` for `recklessness=3` was tested in the form "Death Wish up, or its line true now, or 48 s or less left", which matches what a template helper would build.
