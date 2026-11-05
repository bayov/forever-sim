# Mechanics pass 67 (2026-11-05, master a07941531)

Quiet: no new finding. One open finding updated (mechanics-20261104-1, Energy part fixed).

## Range

dc303b026..a07941531 (9 watched files). Commits on sim / mksettings:

| Commit | What | Mechanics check |
|---|---|---|
| ff56f8ddc (sim/druid/ferocious_bite.go) | Forever damage a point of extra Energy 0.7 / 1.2 / 1.8 / 2.5 / 3.1 / 3.3 (hotfix push 113403) | Matches 70790 SpellEffect #1 bp 70 / 120 / 180 / 250 / 310 / 330 for 1328618 / 22568 / 22827 / 22828 / 22829 / 31018 (tier 1). Updated mechanics-20261104-1; base and per combo point parts still on the 1.25 scale. |
| 22a7d7913 (consumes.go, character.go) | Sapper, Land Mine, EZ-Thro bombs usable in a form; dynamite, grenades, Holy Hand Grenade leave it | 70790 SpellMisc Attributes_0: 13241, 4100, 1269334, 1269330 = 0x0; 12419, 23063, 4068, 19769, 1317268 = 0x10000 (NOT_SHAPESHIFTED). Matches. New verdicts row. Only potions keep their own canUseItemsInForm path. |
| 464deda7b (major_cooldown.go, spell_result.go) | BonusArmor on-use trinkets are survival cooldowns; heal threat in a tank-score replay takes the whole heal | Tank scoring and cooldown classification; no DPS mechanic. |
| aa7cd24dc | Test-only set filter | n/a |

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70790 still newest (`wago.py latest --probe 300`). |
| Dev notes thread | Post 10 (Kaivax, 2026-11-04T19:35:16Z) and post 9 edit (19:35:27Z): a User Interface and Accessibility section only (swing timer fade, bags, nameplates, keybinds). No mechanics. |
| Blue tracker | Linxy newest 2026-11-03, Aggrend 2026-10-22; Kaivax/Gnedi/Brenoch activity not public via user_actions. Beta discussion category: nothing blue in the newest topics. |
| Daily update | forever-update/2026-11-04 merged (33dc144a1); TASKS.md afdf88334 (hotfix pushes 113282/113403 follow-ups, Endless consumables, 70755 item cache). 113403's Ferocious Bite part landed in ff56f8ddc. |
| Hotfix pushes | wago.tools logs about 2,400+ SpellMisc rows for build 70790 (region 91). `hotfixes.py fetch 1.60.1.70790 SpellEffect SpellMisc` ran an hour without finishing SpellMisc and was stopped. Open item: finish the fetch (resumable cache) and run `pushes` / `check` for shared rows next pass. Likely another attribute bake like 70753's, unverified. |

## Conclusions

No shared mechanic moved against tier 1 data. The explosives-in-form rule now follows the client flag. Open item: the 70790 hotfix rows (above).
