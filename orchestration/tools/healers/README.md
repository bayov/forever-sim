# Healer gear scripts

The Restoration shaman's healer stat mix and synthetic set (orchestration/gear_targets.md, Healers;
standards 35). Copy and adapt them per healer spec.

- `bis_parse.py`: reads a Wowhead guide page's slot tables (`[h3 toc=...]` with ranked `[item=...]` rows).
- `shaman_restoration_mix.py FOREVER_GUIDE.html CLASSIC_GUIDE.html`: each slot's first BiS pick of both
  guides, priced at the standards 1b weights with healing on top of spell power at 0.45, and their average.
  Run from the repo root (it reads assets/database/db.json). The Classic picks are listed by hand in it.
- `shaman_restoration_synth.py`: the set from the row's shares: each piece's lines by hand, an IPF pass to the
  shares, each piece's shape fixed within itself, the staff taking what the set still needs.

Fetch the guide pages with tools/database/forever_wowhead/wowhead.py (`Wowhead().page(path)`), never curl;
the pages are cached in ~/.cache/wow-data/wowhead/pages for 7 days.
