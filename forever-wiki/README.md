# WoW Forever wiki

Everything we could find about World of Warcraft: Forever as of 2026-09-14, two days after the BlizzCon 2026 reveal and three days before the beta opens. The purpose is to feed a DPS sim, so the focus is on classes, talents, racials, stats and gear rules.

## Files

- [overview.md](overview.md): what Forever is, dates, rulesets, zones, dungeons, raids, roadmap, the Skyborne.
- [stats-and-gear.md](stats-and-gear.md): itemization and combat stat changes (unified hit and crit, weapon skill, healing gear, new avoidance stat, trinkets).
- [racials.md](racials.md): every race, every racial, Forever text next to Classic text, plus race and class combinations.
- [talent-system.md](talent-system.md): tree structure, milestone rows, per-class counts, removed talents, baseline moves, design patterns.
- [classes/](classes/): one file per class with the full spellbook changes and every talent in all three trees (Forever text, Classic text when changed, ranks, requirements, notes).
  - [warrior](classes/warrior.md), [paladin](classes/paladin.md), [hunter](classes/hunter.md), [rogue](classes/rogue.md), [priest](classes/priest.md), [shaman](classes/shaman.md), [mage](classes/mage.md), [warlock](classes/warlock.md), [druid](classes/druid.md)
- [professions.md](professions.md): profession perks and the camping system.
- [legacy-tree.md](legacy-tree.md): the account-wide Legacy Tree nodes.
- [references.md](references.md): source URLs and related projects (including an existing wowsims fork for Forever).
- [sim-status.md](sim-status.md): what this repo's sim models of Forever, the assumptions behind it, and how to regenerate the talent trees.
- [data/](data/): `abilities.json`, every class ability with its text per rank, generated from the beta client's tables (`tools/sources/abilities.py`).

## How much to trust the numbers

Nothing here comes from a shipped client. The talent and ability text was transcribed by fans from BlizzCon 2026 show-floor footage and streams, where the demo characters were level 38. That means:

- Absolute numbers on abilities (damage, mana cost, attack power on Windfury) are level 38 numbers, not level 60.
- The stream only ever showed rank 1 of most talents. Higher ranks on new or changed multi-rank talents were extrapolated linearly by the dataset authors. The per-talent notes say when that happened.
- Talents with no Classic equivalent have no real spell id yet.
- Wording and values can still change between the demo build and the beta. The beta client (from 2026-09-17, level cap 30 in beta) will be the first hard data.
- Gear is the biggest unknown. Blizzard described the itemization rules but published no item numbers. See [stats-and-gear.md](stats-and-gear.md).

Where the fan dataset marks something as "not yet verified" or "Classic placeholder", it means the Classic Era 1.15.9 text is shown and nobody has seen the Forever version yet.

## Regenerating the class files

The class pages were generated from wowforevertalents.com's BlizzCon transcription (`talents.json` and `abilities.json`, embedded in its page bundle), with a generator that lived in a session scratchpad, not the repo. That data is gone from the repo now. The beta client's trees are generated from its tables into `tools/forever_talents/client/` (`tools/forever_talents/import_db2.py`), and `data/abilities.json` from its spell tables (`tools/sources/abilities.py`). The Rogue, Shaman and Paladin pages open with what changed against the BlizzCon data. The other class pages are still the BlizzCon text, so read the client files for current numbers.
