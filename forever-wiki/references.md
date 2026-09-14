# References

Collected 2026-09-14.

## Official (Blizzard)

- Deep Dive panel recap: https://news.blizzard.com/en-us/article/24303313/world-of-warcraft-forever-deep-dive-panel-recap (camping, Legacy, rulesets, transmog, combat pacing, itemization, racials, talent structure, Paladin example)
- What's Next panel recap: https://news.blizzard.com/en-us/article/24303862/world-of-warcraft-forever-whats-next-panel-recap (dates, zones, dungeons, raids, Skyborne, PvP)
- Pre-purchase and editions: https://news.blizzard.com/en-us/article/24301508/pre-purchase-world-of-warcraft-forever-upgrades-and-begin-your-next-journey-in-azeroth

## Fan datasets (primary source for the class files)

- wowforevertalents.com by Milenkas of Nihilum. Talent calculator, spellbooks, racials, professions, Legacy Tree, roadmap. Data is Classic Era 1.15.9.69722 client tables with a per-talent overlay from BlizzCon footage, each change tagged with its screenshot. https://wowforevertalents.com/ (talents at /warrior/ etc, spellbooks at /abilities/warrior/ etc, /racials/, /professions/, /legacy/, /what-to-expect/, /about/)
- nikftw talent calculator, the screenshot source most of the above cites: https://nikftw.github.io/forevertalent/
- classicwowforever.com talent evidence images: https://classicwowforever.com/talents/
- wowforever.quest change summary (the removed talent lists): https://wowforever.quest/guides/wow-forever-talent-changes
- Wowhead Forever section and talent calculator (JS driven, not scraped): https://www.wowhead.com/forever and https://www.wowhead.com/forever/talent-calc
- zockify talent calculator: https://www.zockify.com/forever/talents/
- wowtbc.gg what we know: https://wowtbc.gg/warcraftforever/news/what-we-know/

## Racials cross-checks

- Method: https://www.method.gg/wow-classic/all-new-racial-abilities-in-world-of-warcraft-forever
- Mobalytics: https://mobalytics.gg/gamebase/guides/wow-forever-racial-abilities
- Warcraft Tavern: https://www.warcrafttavern.com/forever/news/all-racial-abilities-in-world-of-warcraft-forever/

## Press

- Game Informer: https://gameinformer.com/blizzcon-2026/2026/09/12/blizzard-announces-world-of-warcraft-forever-expanding-vanilla-wow-with
- Massively OP reveal: https://massivelyop.com/2026/09/12/blizzcon-2026-world-of-warcraft-forever-announced-as-a-permanent-home-for-classic-players/
- Massively OP deep dive: https://massivelyop.com/2026/09/13/blizzcon-2026-world-of-warcraft-forevers-deep-dive-panel-talks-group-play-progression-and-item-updates/
- Kotaku: https://kotaku.com/world-of-warcraft-forever-classic-plus-revealed-2000733774
- Wccftech: https://wccftech.com/world-of-warcraft-forever-blizzcon-2026-release-date/

## Existing sim work

ElliotWood/Forever (https://github.com/ElliotWood/Forever) is a fork of wowsims/classic already being converted to Forever, active as of 2026-09-14 (50 merged PRs in two days). Worth reading before duplicating effort. What it has:

- A `SimOptions.ruleset` switch (`RulesetClassic` default on the wire, `RulesetForever` default in the UI). Rules gated on it: periodic crits (dots and bleeds roll crit), unified hit and crit from gear, bonus healing grants a third as spell damage, Forever racials, the Skyborne and six new race and class pairings.
- Forever talent trees for all nine classes under both rulesets (talent proto is positional, so trees are not switchable). Imported from the nikftw tooltip dataset via `tools/forever_talents/`. 58 of 469 talents have extrapolated rank scaling.
- Phases renamed to Forever tiers (Launch, Tier1 to Tier3), every preset at Launch with generated pre-raid gear (`tools/launch_gear`, highest EP per slot from non-raid items, using Classic item stats since Forever has none published).
- Onyxia as the one tier 1 encounter.
- Known gaps it lists: Classic items, weapon skill still priced at Classic value, tank warrior runs Fury rotation, mixed raid sizes.

Note on "periodic critical strikes": that rule is in the fork but we did not find it in any Blizzard statement or fan transcription. Treat it as the fork author's reading until confirmed.
