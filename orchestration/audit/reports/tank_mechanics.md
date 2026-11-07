# tank_mechanics pass 6 (2026-11-07, master 12aa2796f, client 1.60.1.70862+hf)

Due on a new client build, 1.60.1.70790 -> 1.60.1.70862. Result: findings (one, low).

**Getting current:** wago's latest is 1.60.1.70862. Blizzard's development notes thread (t/2360696) has a new post 10 (2026-11-04T19:35Z): it only says a UI section was added to post 9, and post 9's edit at 19:35 is that UI section (swing timer, bags, nameplates). Nothing for tanks. The repo range c67d9e014..12aa2796f is an audit-loop commit and an items fold-in; no tank ability, talent or engine file changed.

**Client diff** (`wago.py diff 1.60.1.70790+hf 1.60.1.70862+hf --by-spell`, 441 spell rows; nearly all new creature spells and flight-path dummies). Player tank rows that changed:

| Spell | Field | 70790 | 70862 | Sim | Verdict | Tier |
|---|---|---|---|---|---|---|
| 7376 Defensive Stance Passive | #2 A_MOD_THREAT (misc 127) | 30 | 80 | x1.8 under Forever (stances.go:117) | match, now client | 1 |
| 21178 Bear Form (Passive2) | #0 A_MOD_THREAT (misc 127) | 50 | 115 | x2.15 under Forever (bear.go:58) | match, now client | 1 |
| 12792 Defiance r3 | #0 A_MOD_THREAT | 15 | 15 | 20% (notes) | unchanged, stays notes-only | notes |

Not tank: 57596 Mangle! (creature spell, not in the sim), 20575 Command (Orc pet racial), 48198 Righteous Fire (targeting), The Quick and the Dead 126231x (attribute flags only; the notes' "Righteous Fury made free" is server side), Soulstone aura-option rows removed.

**Engine:** no engine commit in range; the changed rows touch no attack-table, armor, block or Rage data. No sim run needed: both changed values were already in the sim.

**New finding:** tank_mechanics-20261107-1 (low): sim comments and class docs still say the client disagrees with the 80% / 115% threat values.
