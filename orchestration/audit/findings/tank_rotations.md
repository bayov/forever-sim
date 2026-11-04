# Findings: tank_rotations

## Open

### tank_rotations-20261104-1: Bear's Demoralizing Roar takes the pull's first GCD and costs about a tenth of the score
- severity: high
- affects: Feral bear, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: Roar is the first priority line, so it takes the first GCD and 10 Rage ahead of Primal Bite and Lacerate, while SafeDPS is set in the first seconds (tools/rotopt/druid_tank.go:103 appends it above them; ui/feral_tank_druid/apls/*.apl.json:9, max_survival :10; tools/tankbench/req/druid_bear_{balanced,max_threat}.json:141, druid_bear_max_survival.json:164). Moving the line to just below "Lacerate to 5 stacks" (paired, `tools/audit/tank_rotations/abtest`, scoring mean): Balanced +197.5 ± 16.7 (1k; 1855 → ~2053, +10.6%), Max threat +207.8 ± 11.7 (1k), Max survival +176.9 ± 30.0 (3k). Every scoring boss gains (+160 to +244 at Balanced); Ragnaros +128 / +150. Healing needed moves -5 to +29 per boss. Removing Roar wins more (+224) but breaks the 11-02 ruling that the tank applies its own Roar, so it isn't proposed. Roar right after Primal Bite: only +40. Sure: high.
- proposed fix: Roar's line just below the Lacerate keep-up line in all three APLs and requests. Give the template a knob (e.g. `roarAfterLacerate`), since the hard-coded order kept rotopt from finding it. Reproduce B: `tools/audit/tank_rotations/mk_bear.py balanced B.json 'r_=pl.pop(1); pl.insert(3,r_)'` (`pop(2)/insert(4)` for max_survival).
- needs: worker fix

### tank_rotations-20261104-2: Warrior's Shield Wall in the last 12 s is too late; the last 16–18 s wins at Balanced and Max survival
- severity: high
- affects: Protection warrior Balanced, Max survival
- found: 2026-11-04, master f62bfc74
- evidence: line `remainingTime <= 12s` (ui/tank_warrior/apls/level60_{balanced,max_survival}.apl.json:11, tools/tankbench/req/warrior_protection_{balanced,max_survival}.json:162). rotopt's `wallAt` grid is Min 12, Step 12 (tools/rotopt/warrior_tank.go:67), so 13–23 s was never searched (contradicts docs/tank_warrior.md:283-285, "no template knob moves Max survival"). Paired, 3k, scoring mean: 18 s gives Balanced +46.0 ± 7.1 (+1.2%) and Max survival +175.3 ± 19.7 (+11.8%); 16 s gives +37.1 ± 5.8 and +137 ± 16; 20 s +20 / +110; 22 s loses. Vaelastrasz (17 ± 8 s fight) goes the other way: -107 / -283 at 18 s, -56 / -132 at 16 s. Ragnaros +13 / +52 (noise). Max threat at 18 s is -2.4 ± 3.9, so leave it at 12 s.
- proposed fix: 18 s (or 16 s, which costs Vaelastrasz half as much) for Balanced and Max survival, and a finer grid in the template (Min 6, Step 2) with a re-search. Reproduce B: `tools/audit/tank_rotations/mk_warrior.py tools/tankbench/req/warrior_protection_balanced.json B.json setcond 4 '{"cmp":{"lhs":{"remainingTime":{}},"op":"OpLe","rhs":{"const":{"val":"18s"}}}}'`
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18 (warrior): Shield Wall in the last 17 s for Balanced and Max survival (+20.7 ± 5.4, +113.2 ± 13.5 paired 3k vs 12 s; Vaelastrasz -112 / -163), Max threat stays 12 s; template grid 1 s (d9e001609).

### tank_rotations-20261104-3: Paladin's Templar's Bulwark timer was searched on a 10 s grid and misses an optimum at 13 s, which mostly fits Broodlord's script
- severity: high (by the numbers; see the caveat)
- affects: Protection paladin, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: the Bulwark line's const `10s` (ui/protection_paladin/apls/level60_*.apl.json:8); `bulwarkAt` Step 10 (tools/rotopt/paladin_tank.go:48). Paired, 3k: 13 s gives Balanced +73.3 ± 5.0 (+2.9%), Max survival +219.9 ± 14.9 (+17%), Max threat +23.1 ± 1.5; 12 s gives +19.5 / +58.5 / +5.6, 14 s +47.7 (Balanced), 15 s -52.9. Almost all of the gain is Broodlord (+437 at Balanced): 13 s covers his first Mortal Strike, Knock Away and Blast Wave (tools/encounters/broodlord.json firstAt 25±5, 22.5±6.5, 20±7.5, timers the file calls "probably wrong" vmangos values). Vaelastrasz -62, Ragnaros -20 to -61.
- proposed fix: the knob gets Step 1–2 and a re-search. Adopt 13 s only if Broodlord's first-ability timers are trusted; otherwise keep 10 s and fix the timers first. Reproduce B: `tools/audit/tank_rotations/mk_paladin.py A.json B.json bulwark_13s`.
- needs: user ruling
- left open: 2026-11-04 (tank-paladin, e9751cc). bulwarkAt now steps by 1 sec, and rotopt holds it at 10 (`-freeze bulwarkAt`). Paired, 3,000 iterations, 13 s minus 10 s: on the five scoring bosses without Broodlord Max threat -1.1 ± 1.1, Balanced -6.2 ± 3.0, Max survival -18.5 ± 9.0 (Vaelastrasz -24 / -66 / -199, Onyxia +9 / +26 / +79); with Broodlord +16.5 ± 1.4 / +49.3 ± 3.9 / +147.9 ± 11.7 (Broodlord +105 / +327 / +980). 13 s wins only on Broodlord's suspect timers, so 10 s stays. The user's call (docs/tank_paladin.md, Calls): 13 s if Broodlord's timers are trusted, and Bulwark on a timer every pull at all.

### tank_rotations-20261104-4: Warden's Stormstrike sits above Earth Shock, which makes more threat per GCD
- severity: high (Max threat); medium elsewhere
- affects: Warden shaman, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: Earth Shock does about 2,225 threat a cast against Stormstrike's 1,500 (Golemagg, -actions), and with Stormstrike first it waits whenever both are ready. Order hard-coded at tools/rotopt/shaman_tank.go:150; Stormstrike at tools/tankbench/req/shaman_warden_*.json:194/204/208. Swapping them, paired, 3k, scoring mean: Max threat +62.0 ± 16.2 (+2.4%); Balanced +51.2 ± 48.6; Max survival -18.7 ± 148. SafeDPS rises on every scoring boss at all three weights (Vaelastrasz +75, Onyxia +43); the Balanced and Max survival error comes from Broodlord's death floor near 5% (without Broodlord: +67 / +48).
- proposed fix: emit the shock before Stormstrike in the template (swaps lines 12↔13 in level60_max_threat/max_survival, 11↔12 in level60_balanced, and the requests). Confirm Balanced and Max survival with a rotopt -eval first. Reproduce: `tools/audit/tank_rotations/mk_warden.py <preset> swap 'i=[j for j,a in enumerate(pl) if a["action"].get("castSpell",{}).get("spellId",{}).get("spellId")==17364][0]; pl[i],pl[i+1]=pl[i+1],pl[i]'`.
- needs: worker fix
- fixed: 2026-11-04, 952a7a4 (template knob `shockFirst`, default 1) and dd4b82a (the three presets re-searched with it; paired: Max threat +91.6 ± 10.6, Balanced +105.0 ± 12.5, Max survival +129.3 ± 33.1).

### tank_rotations-20261104-5: Bear Max survival's Enrage at -3 s loses to Balanced's -1 s
- severity: medium (+46, about 5% of a negative score of -844)
- affects: Feral bear Max survival
- found: 2026-11-04, master f62bfc74
- evidence: prepull shift -4.5 s and Enrage -3 s (tools/tankbench/req/druid_bear_max_survival.json:111, :125), reasoned in docs/tank_druid.md:224 as "so its armor cut ends sooner". Shift -2.5 s and Enrage -1 s, paired, 3k: +45.6 ± 16.6 on top of finding 1, +60.1 ± 17.6 on today's preset. Chromaggus -6 and Magmadar -13 (noise); Ragnaros -50 ± 200.
- proposed fix: shift at -2.5 s and Enrage at -1 s, as Balanced does, and fix the doc's reason. Reproduce: `tools/audit/tank_rotations/mk_bear.py max_survival B.json 'r_=pl.pop(2); pl.insert(4,r_); pre[0]["doAtValue"]["const"]["val"]="-2.5s"; pre[1]["doAtValue"]["const"]["val"]="-1s"'`.
- needs: worker fix

### tank_rotations-20261104-6: Paladin's on-use trinket waits behind every GCD line
- severity: medium
- affects: Protection paladin, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: `autocastOtherCooldowns` is the last line (ui/protection_paladin/apls/level60_*.apl.json:15, tools/rotopt/paladin_tank.go:117), noted "on a free GCD", but trinket 990716 (+165 Str, 20 s) is off the GCD, so the line only delays it. Moved to the top, paired, 3k: Balanced +13.9 ± 0.6 (+0.5%), Max threat +14.1 ± 0.2, Max survival +13.4 ± 1.8 (+1.1%); mostly Vaelastrasz (+44). Ragnaros -2 to -21.
- proposed fix: move the line to the top, with the note "Trinkets on cooldown from the pull, off the global cooldown." Reproduce: `tools/audit/tank_rotations/mk_paladin.py A.json B.json autocast_top`.
- needs: worker fix
- fixed: 2026-11-04, 86b9631 (the presets; rotopt's template in e9751cc).

### tank_rotations-20261104-7: Warrior Max threat's Sunder filler at 30 Rage is a little high
- severity: low
- affects: Protection warrior Max threat
- found: 2026-11-04, master f62bfc74
- evidence: priorityList[10] `currentRage >= 30`; `>= 15` gives +3.9 ± 1.1 (3k, paired, every boss ≥ 0, Ragnaros +1.9 ± 10.6). Bloodrage on cooldown gives +1.4 ± 0.5 too (noted only).
- proposed fix: Sunder filler at 15 Rage, if the template is re-searched anyway.
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18: Sunder filler at 15 Rage on the re-searched Max threat talents (+6.9 ± 1.6; -0.3 ± 0.8 on the old ones).

### tank_rotations-20261104-8: Rotation notes that don't say what the line does (standards 31)
- severity: low
- affects: Warden shaman, Protection warrior, Feral bear
- found: 2026-11-04, master f62bfc74
- evidence: (a) Warden's shield line, a Lightning Shield cast, is noted "Water Shield needs the Restoration talent" (tools/rotopt/shaman_tank.go:118, every level60_*.apl.json). (b) Warrior's Defensive Stance line (spell 71) has no note in all three APLs. (c) docs/tank_druid.md:217 says "Enrage at the pull and 1 sec before it", but the APL casts it only before the pull (an in-fight Enrage loses -5.3 ± 1.3). Bear Max survival's Frenzied Regeneration at 100 Rage fires about once, at the pull's Rage peak, not "on cooldown" as its note says (the line itself is right: removing it costs 432).
- proposed fix: (a) "Lightning Shield whenever its orbs are gone: their threat beats the 370 mana." (b) "Defensive Stance kept up: the tank's stance." (c) the doc: "Enrage 1 sec before the pull"; the Frenzied Regeneration note: "at the pull's Rage peak".
- needs: worker fix
- fixed (shaman part, (a)): 2026-11-04, 952a7a4 (the template's shield note; the APLs carry it since af78f8a).
- fixed (warrior part b): 2026-11-04, c588675a6 and 61db5cccb. (a) and (c) are the shaman and druid workers'.

## Resolved
