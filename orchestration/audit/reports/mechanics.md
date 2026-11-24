# Mechanics pass 101 (2026-11-24)

Master c78d983ac..566067a39. Client 1.60.1.71037 still the newest (wago probe 300). Result: quiet.

## Get current

| Source | Newest read | New since pass 100 |
|---|---|---|
| Dev notes thread 2360696 | 8 posts, last 2026-11-04 | none |
| Hotfix thread 2403040 | post 14, 2026-11-21 | none |
| Blue tracker | Kaivax 2026-11-23 18:55Z, topic 2424420 (Epic Mount Riding Skill) | not a combat mechanic |
| Daily update | no TASKS.md or update_ledger.md commit in range | none |

## Commits

| Commit | Area | Verdict |
|---|---|---|
| cd71e78c2 | Warrior: Heroic Strike / Cleave spend one Flurry charge a cast under Forever (landed or not, once per Cleave, outside the 0.5 s white window); Reckless Swings spends on every HS / Cleave cast, white swings only when landed; Unbridled Wrath skips Shield Slam, Revenge, Shield Bash | Class rules from wowsims/forever logs EV-0138 / EV-0148 (after push 114115 / 114122). The Flurry part reverses 30a645072 and the November 16 hotfix wording ("Flurry stacks were consumed by Cleave or Heroic Strike"), but the log postdates the hotfix and the user ruling of 2026-11-22 (a rule measured from a Forever log wins) settles it. Not a finding. Windfury Totem's swing still skipped for warrior (mechanics-20261118-1 stays open for shaman). |
| be773f5eb | Druid: Rip tick 0.0114 AP a combo point (fifth as fourth), Rake tick 0.065, both outside nov2BleedScale | Log-sourced (EV-0031..0036, builds 70862 / 70888, after the Nov 2 notes), same ruling; the client leaves these shares to the server. Note added under mechanics-20261104-1 (Ferocious Bite and Starfire parts unchanged). |
| b82ad9a1e | Hunter: pet summoned under a prepull Aspect of the Hawk gets 10% of the aspect's RAP | User-ruled default (2026-11-23), class rule; no shared formula touched. |

No commit touched `sim/core` or `tools/rotopt/mksettings`; no shared-raid, racial or consumable value moved.

## Open findings unchanged

mechanics-20261123-1, -20261120-2, -20261118-2, -20261118-1 (shaman half), -20261113-2, -20261104-1.
