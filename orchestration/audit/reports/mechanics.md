# Mechanics pass 90 (2026-11-18, master a6677362fc)

Range 8219f1083..a6677362fc (six sim commits, one TASKS.md commit). Client unchanged at 1.60.1.70959.
One new finding (low, user ruling). Open findings mechanics-20261118-1, mechanics-20261113-2 and
mechanics-20261104-1 stay open (18-1 gains a note: the Nov 17 Flurry fix is "announced, not in the game yet").

## Get current

| Source | Newest | New since pass 89 |
|---|---|---|
| Client build | 1.60.1.70959 (gate) | none |
| Dev notes t/2360696 | post 10, 2026-11-04 19:35Z | none (no edits since) |
| Hotfix post t/2403040 | post 12, 2026-11-18 03:24Z (post 1 edited 03:24:32Z) | none |
| Blue tracker | Forever hotfixes Nov 17 (2026-11-18 03:24Z) | nothing newer for Forever |
| Daily update / TASKS | a6677362f: Flurry/WF Totem fix pending live; Dungeon Set 1 follow-ups (item data) | noted on mechanics-20261118-1 |

## Commits

| Commit | Change | Verdict |
|---|---|---|
| ee35da473 | Forever spell GCD = 1.5 / (1 + cast speed); Warchief's Blessing +15% cast speed under Forever | Tier 1 log (user's Enhancement, 2026-11-18). 16609's effects are all aura 4 (dummy) in 70959, so the client can't confirm or contradict the cast speed. The spell test (`hasSpellGCD`) is school-based; client damage class says shouts and aspects are magic: mechanics-20261118-2. |
| c2efe4ae3 | Floor every hasted spell GCD at 1s | User ruling 2026-11-18; no spell of ours has a base GCD under 1s. Not a finding. |
| c8b5275b1 | DPS Report patches folded into env-var switches (report_switches.go, racials_forever.go, rage.go, spell_outcome.go, debuffs.go, wrath.go, hunter.go) | Default paths read unchanged: racial parts gated by `on()` that is true when DASH_RACIALS_OFF is unset; rage scaling and damage-taken factor default to the constants; tick crit multiplier unchanged unless bleed_tick_crit is set. haste_gcd knob absent from sim/ at master, so no double haste. |
| 7199c8f9c | Healer idle skip wakes for hits seen a reaction time late | Engine scheduling for heal rotations; no mechanic. |
| cc5a55f47 | Faster start-up (item set index, slimmer db decode), labeled roll cache | Same results by design; no mechanic. |
| bbfd014e4 | healbench speed-ups (heal sight cache, health history search, spell metrics) | Same results to the bit per commit; no mechanic. |

## Client damage class (SpellCategories.DefenseType, 70959)

| Class | Spells |
|---|---|
| 1 magic | Battle Shout 25289, Demoralizing Shout 27579, Aspects Hawk 25296 / Monkey 13163 / Wild 20190, Earth Shock 10414, Corruption 11672, Seal of Command 20375 |
| 0 none | Death Wish 12328, Berserker Rage 18499, Slice and Dice 6774, Blade Flurry 13877, Hunter's Mark 1213268, Searing Totem 10438, Judgement 20271 |
| 2 melee | Stormstrike 17364, Recklessness 13847 |

## Not done

- No sims run: the new finding's DPS size is an estimate (under 0.05%), stated as such.
